package Requester

import (
	"../LinkSystem"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"time"
)

type Requester struct {
	StartUrl  string
	domainUrl string
	timeout   time.Duration
}

/**
* Exposed function to check url and return links
* @parameter null
* @return []string, []string
 */
func (req Requester) Check() ([]string, []string) {
	linkSystem := LinkSystem.LinkSystem{}
	var external []string
	var internal []string

	document := req.makeCheck()
	links := linkSystem.Clean(req.getLinks(document))

	external, internal = linkSystem.DivideByType(links, linkSystem.GetDomainUrl(req.StartUrl))
	external, _ = linkSystem.IsSubdomain(external)

	return external, internal
}

/**
* Make HTTP Request to a given url
* @parameter null
* @return goquery.document
 */
func (r Requester) makeCheck() goquery.Document {
	client := &http.Client{
		Timeout: r.timeout * time.Second,
	}

	// Create and modify HTTP request before sending
	request, err := http.NewRequest("GET", r.StartUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("User-Agent", "Not Firefox")

	// Make request
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}
	return *document
}

/**
* Get Links from a given document
* @parameter goquery.document
* @return []string
 */
func (r Requester) getLinks(d goquery.Document) []string {
	var links []string
	d.Find("a").Each(func(index int, element *goquery.Selection) {
		href, exists := element.Attr("href")
		if exists {
			links = append(links, href)
		}
	})
	return links
}
