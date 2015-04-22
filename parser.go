package ruslanparser

import (
	"encoding/xml"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type (
	Api struct {
		Ssid       string
		SsidAssign time.Time

		InitialUrl string
		PresentUrl string

		GetAllRecords bool
		MAXRECORDS    int
	}
)

var (
	BASE_URL       = "http://92.241.99.100/Scripts/zgate.exe"
	INDEX_QUERY    = "Init+test.xml,simple.xsl+rus"
	PRESENT_FORMAT = "present+%s+default+%d+%d+X+1.2.840.10003.5.102+rus"
	HOST           = "192.168.0.1"
	PORT           = "210"
	LANG           = "rus"
	ACTION         = "SEARCH"
	ESNAME         = "B"
	MATERIAL_TYPE  = ""
	DBNAMES        = []string{"BOOK", "REF", "STAT", "ЭЛЕКТР_РЕСУРСЫ", "ФРК"}
	USE_1          = "1035"
	TERM_1         = "%D0%BC%D0%B0%D1%82%D0%B5%D0%BC%D0%B0%D1%82%D0%B8%D0%BA%D0%B0"
	BOOLEAN_OP1    = "AND"
	USE_2          = "4"
	TERM_2         = ""
	BOOLEAN_OP2    = "AND"
	USE_3          = "21"
	TERM_3         = ""
	SHOW_HOLDINGS  = "on"
	MAXRECORDS     = "60"
	SEARCH         = "SEARCH"
)

func NewApi() *Api {
	a := &Api{}
	a.MAXRECORDS = 20
	return a
}

func (a *Api) ProxySearch(q string) []Book {
	data := a.searchValues(q)
	query := data.Encode()
	r, err := http.Get(BASE_URL + "?" + query)

	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var res []Book

	records := getBooksData(a.Ssid, 10)
	for _, v := range records.Records {
		res = append(res, v.ToBook())
	}
	return res
}

func (a *Api) searchValues(q string) url.Values {
	var data = url.Values{}
	data.Set("HOST", HOST)
	data.Set("PORT", PORT)
	data.Set("SESSION_ID", a.GetSsid())
	data.Set("LANG", LANG)
	data.Set("ACTION", SEARCH)
	data.Set("ESNAME", ESNAME)
	data.Set("MATERIAL_TYPE", MATERIAL_TYPE)
	data.Set("USE_1", USE_1)
	data.Set("TERM_1", q)
	data.Set("BOOLEAN_OP1", BOOLEAN_OP1)
	data.Set("USE_2", USE_2)
	data.Set("TERM_2", TERM_2)
	data.Set("BOOLEAN_OP2", BOOLEAN_OP2)
	data.Set("USE_3", USE_3)
	data.Set("TERM_3", TERM_3)
	data.Set("SHOW_HOLDINGS", SHOW_HOLDINGS)
	data.Set("MAXRECORDS", MAXRECORDS)
	data.Set("SEARCH", SEARCH)
	for _, v := range DBNAMES {
		data.Add("DBNAME", v)
	}
	return data
}

func NewSessionSearchUrl(iurl string) string {
	d, err := goquery.NewDocument(iurl)
	if err != nil {
		return ""
	}
	s, _ := d.Find("input[name=SESSION_ID]").Eq(0).Attr("value")
	return s
}

func (a *Api) Search(q string) []Book {
	var data = a.searchValues(q)
	query := data.Encode()
	r, err := http.Get(BASE_URL + "?" + query)

	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	rb, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	maxrecords := resultMax(rb)

	if !a.GetAllRecords {
		maxrecords = a.MAXRECORDS
	}

	var res []Book

	records := getBooksData(a.Ssid, maxrecords)
	for _, v := range records.Records {
		res = append(res, v.ToBook())
	}
	return res
}

func (a *Api) fetch(q string) Result {
	return Result{}
}

func (a *Api) GetSsid() string {
	if a.Ssid != "" && time.Now().Sub(a.SsidAssign).Minutes() < 5.0 {
		return a.Ssid
	}
	d, err := goquery.NewDocument(BASE_URL + "?" + INDEX_QUERY)
	if err != nil {
		panic(err)
	}
	s, _ := d.Find("input[name=SESSION_ID]").Eq(0).Attr("value")
	a.Ssid = s
	a.SsidAssign = time.Now()
	fmt.Println(s)
	return s
}

func getBooksData(session string, max int) Records {
	var v = Records{}
	base := BASE_URL + "?" + PRESENT_FORMAT
	start := 1
	lim := 56
	if max < lim {
		lim = max
	}
	fmt.Printf("Start fetching. Results: %d\n", max)
	for start < max {
		fmt.Printf("Start fetching %d,%d\n", start, lim)
		u := fmt.Sprintf(base, session, start, lim)
		d, err := http.Get(u)
		if err != nil {
			panic(err)
		}
		defer d.Body.Close()
		b, err := ioutil.ReadAll(d.Body)
		if err != nil {
			panic(err)
		}
		b = careBad(b)
		var vt Records
		err = xml.Unmarshal(b, &vt)
		v.Records = append(v.Records, vt.Records...)
		if err != nil {
			fmt.Println(string(b))
			panic(err)
		}
		if start+lim > max {
			lim = max - start
		}
		start += lim
	}
	return v
}

func careBad(x []byte) []byte {
	s := string(x)
	s = strings.Replace(s, "?>", "?><records>", -1)
	s = s + "\n" + "</records>"
	return []byte(s)
}

func resultMax(r []byte) int {
	red := strings.NewReader(string(r))
	d, err := goquery.NewDocumentFromReader(red)
	if err != nil {
		panic(err)
	}
	t := d.Find("span.succ").Text()
	reg, err := regexp.Compile("Записи с [\\d]* по [\\d]* из ([\\d]*)")
	if err != nil {
		panic(err)
	}
	bts := reg.FindAllStringSubmatch(t, -1)
	if len(bts) > 0 {
		if len(bts[0]) > 1 {
			n, _ := strconv.Atoi(bts[0][1])
			return n
		}
	}
	return 0
}

func RM(r []byte) int {
	return resultMax(r)
}
