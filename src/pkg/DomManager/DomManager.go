package DomManager

import (
	"github.com/PuerkitoBio/goquery"
)

type DomManager struct {
	Document goquery.Document
}

/**
* Get Links from a given document
* @parameter goquery.document
* @return []string
 */
func (dm DomManager) GetURLFromDocument() []string {
	var links []string
	dm.Document.Find("a").Each(func(index int, element *goquery.Selection) {
		href, exists := element.Attr("href")
		if exists {
			links = append(links, href)
		}
	})
	return links
}
