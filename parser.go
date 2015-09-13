package ruslanparser

import (
	"encoding/xml"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ungerik/go-dry"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	Api struct {
		mx         sync.Mutex
		Ssid       string
		SsidFile   string
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

	BY_AUTHOR = "1003"
	BY_TITLE  = "4"
)

func Search(q, filter, ssid string, offset, limit int) (books []Book, numResult int, e error) {
	if limit == 0 {
		limit = 10
	}
	switch filter {
	case "author":
		filter = BY_AUTHOR
	default:
		filter = BY_TITLE
	}

	var data = searchValues(q, ssid)
	query := data.Encode()
	var (
		bts []byte
	)
	bts, e = get(BASE_URL + "?" + query)
	if e != nil {
		return
	}

	maxrecords := resultMax(bts)
	numResult = maxrecords
	if maxrecords > limit {
		maxrecords = limit
	}

	records, e := getBooksData(ssid, offset, maxrecords)
	for _, v := range records.Records {
		books = append(books, v.ToBook())
	}
	return
}

func NewApi(opts ...*Option) *Api {
	a := &Api{}
	a.MAXRECORDS = 20
	err := a.loadOpts(opts...)
	if err != nil {
		log.Println(err)
		return nil
	}
	go a.GetSsid()
	return a
}

func (a *Api) loadOpts(opts ...*Option) error {
	for _, v := range opts {
		switch v.Type() {
		case OptTypeSessionFile:
			a.mx.Lock()
			defer a.mx.Unlock()
			str, err := dry.FileGetString(v.Value())
			log.Println("LOAD Ssid from file:", str)
			if err != nil {
				if !dry.FileExists(v.Value()) {
					err := ioutil.WriteFile(v.Value(), []byte(""), 0777)
					if err != nil {
						return err
					}
				}
				return err
			}
			a.SsidFile = v.Value()
			a.Ssid = str
			a.SsidAssign = time.Now()
		}
	}
	return nil
}

func (a *Api) ProxySearch(q string) []Book {
	data := a.searchValues(q)
	query := data.Encode()
	get(BASE_URL + "?" + query)

	var res []Book

	records, _ := getBooksData(a.Ssid, 0, 10)
	for _, v := range records.Records {
		res = append(res, v.ToBook())
	}
	return res
}

func (a *Api) SearchByAuthor(author string) ([]Book, error) {
	return a.BySearch(author, BY_AUTHOR)
}

func (a *Api) SearchByTitle(title string) ([]Book, error) {
	return a.BySearch(title, BY_TITLE)
}

func (a *Api) BySearch(q string, by string) ([]Book, error) {
	data := a.searchValues(q)
	data.Set("USE_1", by)

	query := data.Encode()
	get(BASE_URL + "?" + query)

	var res []Book

	records, _ := getBooksData(a.Ssid, 0, 10)
	for _, v := range records.Records {
		res = append(res, v.ToBook())
	}
	return res, nil
}

func searchValues(q, ssid string) url.Values {
	var data = url.Values{}
	data.Set("TERM_1", q)
	data.Set("SESSION_ID", ssid)

	data.Set("HOST", HOST)
	data.Set("PORT", PORT)
	data.Set("LANG", LANG)
	data.Set("ACTION", SEARCH)
	data.Set("ESNAME", ESNAME)
	data.Set("MATERIAL_TYPE", MATERIAL_TYPE)
	data.Set("USE_1", USE_1)
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

func (a *Api) searchValues(q string) url.Values {
	return searchValues(q, a.GetSsid())
}

func (a *Api) Search(q string) []Book {
	var data = a.searchValues(q)
	query := data.Encode()
	rb, err := get(BASE_URL + "?" + query)
	if err != nil {
		panic(err)
	}

	maxrecords := resultMax(rb)

	if !a.GetAllRecords {
		maxrecords = a.MAXRECORDS
	}

	var res []Book

	records, _ := getBooksData(a.Ssid, 0, maxrecords)
	for _, v := range records.Records {
		res = append(res, v.ToBook())
	}
	return res
}

func (a *Api) fetch(q string) Result {
	return Result{}
}

func (a *Api) GetSsid() string {
	a.mx.Lock()
	defer a.mx.Unlock()
	if a.Ssid != "" && time.Now().Sub(a.SsidAssign).Minutes() < 5.0 {
		return a.Ssid
	}
	d, err := goquery.NewDocument(BASE_URL + "?" + INDEX_QUERY)
	log.Println("GET", BASE_URL+"?"+INDEX_QUERY)
	if err != nil {
		log.Println(err)
		return ""
	}
	s, _ := d.Find("input[name=SESSION_ID]").Eq(0).Attr("value")
	a.Ssid = s
	a.SsidAssign = time.Now()
	if a.SsidFile != "" {
		err := ioutil.WriteFile(a.SsidFile, []byte(s), 0777)
		if err != nil {
			log.Println(err)
			return ""
		}
	}
	fmt.Println(s)
	return s
}

func getSsid(uri string) (string, error) {
	d, err := goquery.NewDocument(uri)
	if err != nil {
		return "", err
	}
	s, _ := d.Find("input[name=SESSION_ID]").Eq(0).Attr("value")
	return s, nil
}

func GetBooksData(session string) Records {
	var v = Records{}
	base := BASE_URL + "?" + PRESENT_FORMAT
	start := 1
	lim := 20
	u := fmt.Sprintf(base, session, start, lim)
	b, err := get(u)
	if err != nil {
		panic(err)
	}
	b = careBad(b)
	var vt Records
	err = xml.Unmarshal(b, &vt)
	if err != nil {
		fmt.Println(string(b))
		panic(err)
	}
	v.Records = append(v.Records, vt.Records...)
	return v
}

func getBooksData(session string, offset, limit int) (Records, error) {
	var v = Records{}
	base := BASE_URL + "?" + PRESENT_FORMAT
	start := offset + 1
	max := limit
	lim := 56
	if max < lim {
		lim = max
	}
	fmt.Printf("Start fetching. Results: %d\n", max)
	for start < max {
		fmt.Printf("Start fetching %d,%d\n", start, lim)
		u := fmt.Sprintf(base, session, start, lim)
		b, err := get(u)
		if err != nil {
			return Records{}, err
		}
		b = careBad(b)
		var vt Records
		err = xml.Unmarshal(b, &vt)
		if err != nil {
			return Records{}, err
		}
		v.Records = append(v.Records, vt.Records...)
		if len(vt.Records) < lim {
			return v, nil
		}
		if start+lim > max {
			lim = max - start
		}
		start += lim
	}
	return v, nil
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

func get(u string) ([]byte, error) {
	fmt.Println("GET", u)
	return dry.FileGetBytes(u)
}
