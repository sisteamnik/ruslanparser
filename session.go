package ruslanparser

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
)

func NewSessionSearchUrl(iurl string) (string, error) {
	d, err := goquery.NewDocument(iurl)
	if err != nil {
		return "", err
	}
	s, ok := d.Find("input[name=SESSION_ID]").Eq(0).Attr("value")
	if !ok {
		return "", fmt.Errorf("session not found")
	}
	return s, nil
}
