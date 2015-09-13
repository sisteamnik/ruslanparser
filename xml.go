package ruslanparser

import (
	"encoding/xml"
	"github.com/sisteamnik/guseful/md5"
	"strconv"
	"time"
)

type Records struct {
	XMLName xml.Name `xml:"records"`
	Records []Result `xml:"record"`
}

type Result struct {
	XMLName             xml.Name            `xml:"record"`
	BibliographicRecord BibliographicRecord `xml:"bibliographicRecord"`
	HoldingsData        HoldingsData        `xml:"holdingsData"`
}

type BibliographicRecord struct {
	Record Record `xml:"record"`
}

type Record struct {
	Leader Leader  `xml:"leader"`
	Field  []Field `xml:"field"`
}

type Field struct {
	Id        string      `xml:"id,attr"`
	Val       string      `xml:",chardata"`
	Indicator []Indicator `xml:"indicator"`
	Subfield  []Subfield  `xml:"subfield"`
}

type Indicator struct {
	Id  string `xml:"id,attr"`
	Val string `xml:",chardata"`
}

type Subfield struct {
	Id  string `xml:"id,attr"`
	Val string `xml:",chardata"`
}

type Leader struct {
	Length   string `xml:"length"`
	Status   string `xml:"status"`
	Type     string `xml:"type"`
	Leader07 string `xml:"leader07"`
	EntryMap string `xml:"entryMap"`
}

type HoldingsData struct {
	HoldingsAndCirc []HoldingsAndCirc `xml:"holdingsAndCirc"`
}

type HoldingsAndCirc struct {
	NucCode         string          `xml:"nucCode"`
	LocalLocation   string          `xml:"localLocation"`
	CallNumber      string          `xml:"callNumber"`
	ShelvingData    string          `xml:"shelvingData"`
	CopyNumber      string          `xml:"copyNumber"`
	CirculationData CirculationData `xml:"circulationData"`
}

type CirculationData struct {
	CircRecord CircRecord `xml:"circRecord"`
}

type CircRecord struct {
	AvailableNow int `xml:"availableNow"`
}

func (r Result) ToBook() Book {
	b := Book{}
	b.Title = getTitle(r)
	b.Author = getAuthor(r)
	b.PublicationYear = getPublicationYear(r)
	b.City = getCity(r)
	b.Tags = getTags(r)
	b.Publishing = getPublishing(r)
	b.Genre = getGenre(r)
	b.Edition = getEdition(r)
	b.Series = getSeries(r)
	b.Places = getPlaces(r)
	b.LastMod = getLastMod(r)
	b.SourceId = getSourceId(r)
	b.Fields = map[string]map[string]string{}

	for _, v := range r.BibliographicRecord.Record.Field {
		if b.Fields[v.Id] == nil {
			b.Fields[v.Id] = map[string]string{}
		}
		for _, j := range v.Subfield {
			b.Fields[v.Id][j.Id] = j.Val
		}
	}
	return b
}

func getTitle(r Result) string {
	f := r.BibliographicRecord.Record.Field
	for _, v := range f {
		if v.Id == "200" {
			for _, k := range v.Subfield {
				if k.Id == "a" {
					return k.Val
				}
			}
		}
	}
	return ""
}

func getAuthor(r Result) string {
	var ln string
	var on string
	var onlong string
	f := r.BibliographicRecord.Record.Field
	for _, v := range f {
		if v.Id == "700" {
			for _, k := range v.Subfield {
				if k.Id == "a" {
					ln = k.Val
				}
				if k.Id == "b" {
					on = k.Val
				}
				if k.Id == "g" {
					onlong = k.Val
				}
			}
		}
	}
	name := ln + " " + on
	if onlong != "" {
		name = ln + " " + onlong
	}
	return name
}

func getPublicationYear(r Result) int {
	y := r.getSubfield("210", "d")
	i, _ := strconv.Atoi(y)
	return i
}

func getCity(r Result) string {
	return r.getSubfield("210", "a")
}

func getTags(r Result) []string {
	var res = []string{}
	f := r.BibliographicRecord.Record.Field
	for _, v := range f {
		if v.Id == "606" {
			for _, k := range v.Subfield {
				if k.Id == "a" || k.Id == "x" {
					res = append(res, k.Val)
				}
			}
		}
	}
	return res
}

func getPublishing(r Result) string {
	f := r.BibliographicRecord.Record.Field
	for _, v := range f {
		if v.Id == "210" {
			for _, k := range v.Subfield {
				if k.Id == "c" {
					return k.Val
				}
			}
		}
	}
	return ""
}

func getGenre(r Result) string {
	f := r.BibliographicRecord.Record.Field
	for _, v := range f {
		if v.Id == "610" {
			for _, k := range v.Subfield {
				if k.Id == "a" {
					return k.Val
				}
			}
		}
	}
	return ""
}

func getEdition(r Result) string {
	f := r.BibliographicRecord.Record.Field
	for _, v := range f {
		if v.Id == "205" {
			for _, k := range v.Subfield {
				if k.Id == "a" {
					return k.Val
				}
			}
		}
	}
	return ""
}

func getSeries(r Result) string {
	return r.getSubfield("200", "e")
}

func (r Result) getSubfield(fId, sId string) string {
	f := r.BibliographicRecord.Record.Field
	for _, v := range f {
		if v.Id == fId {
			for _, k := range v.Subfield {
				if k.Id == sId {
					return k.Val
				}
			}
		}
	}
	return ""
}

func getPlaces(r Result) []Place {
	var res = []Place{}
	f := r.HoldingsData.HoldingsAndCirc
	for _, v := range f {
		var p Place
		p.Available = v.CirculationData.CircRecord.AvailableNow
		p.Name = v.LocalLocation
		p.ShelvingIndex = v.ShelvingData
		p.Cipher = v.CallNumber
		res = append(res, p)
	}
	return res
}

func getLastMod(r Result) string {
	var t time.Time
	var err error
	f := r.BibliographicRecord.Record.Field
	for _, v := range f {
		if v.Id == "005" {
			t, err = time.Parse("20060102150405.0", v.Val)
			if err != nil {
				panic(err)
			}
		}
	}
	return t.Format(time.RFC3339)
}

func getSourceId(r Result) string {
	f := r.BibliographicRecord.Record.Field
	for _, v := range f {
		if v.Id == "001" {
			return md5.Hash(v.Val)
		}
	}
	return ""
}
