package ruslanparser

import (
	"encoding/xml"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	Api struct {
		Ssid       string
		SsidAssign time.Time

		InitialUrl string
	}
)

var (
	HOST          = "192.168.0.1"
	PORT          = "210"
	LANG          = "rus"
	ACTION        = "SEARCH"
	ESNAME        = "B"
	MATERIAL_TYPE = ""
	DBNAMES       = []string{"BOOK", "REF", "STAT", "ЭЛЕКТР_РЕСУРСЫ", "ФРК"}
	USE_1         = "1035"
	TERM_1        = "%D0%BC%D0%B0%D1%82%D0%B5%D0%BC%D0%B0%D1%82%D0%B8%D0%BA%D0%B0"
	BOOLEAN_OP1   = "AND"
	USE_2         = "4"
	TERM_2        = ""
	BOOLEAN_OP2   = "AND"
	USE_3         = "21"
	TERM_3        = ""
	SHOW_HOLDINGS = "on"
	MAXRECORDS    = "60"
	SEARCH        = "SEARCH"
)

func NewApi(initialurl string, host string, dbnames []string) *Api {
	a := &Api{}
	HOST = host
	DBNAMES = dbnames
	a.InitialUrl = initialurl
	return a
}

func (a *Api) Search(q string) []Book {
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
	query := data.Encode()
	r, err := http.Get("http://92.241.99.100/Scripts/zgate.exe?" + query)

	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	d, err := goquery.NewDocumentFromResponse(r)
	if err != nil {
		panic(err)
	}

	var res []Book

	d.Find(".recordmenu a").Each(func(i int, s *goquery.Selection) {
		a, ok := s.Attr("href")
		if ok {
			b := getBookData(a)
			res = append(res, b.ToBook())
		}
	})
	return res
}

func (a *Api) fetch(q string) Result {
	return Result{}
}

func (a *Api) GetSsid() string {
	if a.Ssid != "" && time.Now().Sub(a.SsidAssign).Minutes() < 30.0 {
		return a.Ssid
	}
	d, err := goquery.NewDocument(a.InitialUrl)
	if err != nil {
		panic(err)
	}
	s, _ := d.Find("input[name=SESSION_ID]").Eq(0).Attr("value")
	a.Ssid = s
	a.SsidAssign = time.Now()
	return s
}

func getBookData(u string) Result {
	base := "http://92.241.99.100/Scripts/"
	u = strings.Replace(u, "+F+", "+X+", -1)
	d, err := http.Get(base + u)
	if err != nil {
		panic(err)
	}
	defer d.Body.Close()
	b, err := ioutil.ReadAll(d.Body)
	if err != nil {
		panic(err)
	}
	var v Result
	err = xml.Unmarshal(b, &v)
	if err != nil {
		panic(err)
	}
	return v
}
