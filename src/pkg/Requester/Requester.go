package Requester

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"regexp"
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
func (r Requester) Check() ([]string, []string) {
	r.getDomainUrl()

	document := r.makeCheck()
	links := r.getLinks(document)

	return r.checkLinkType(links)
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

/**
* Get DomainUrl from given start url
* @parameter
* @return
 */
func (r *Requester) getDomainUrl() {
	regex, _ := regexp.Compile("(http(s)?:\\/\\/)([\\w.-]+)(\\.[\\w\\.-]+)")
	found := regex.FindStringSubmatch(r.StartUrl)
	r.domainUrl = found[3] + found[4]
}

/**
* Check link type
* @parameter []string
* @return
 */
func (r Requester) checkLinkType(links []string) ([]string, []string) {
	var external []string
	var internal []string
	isValid, _ := regexp.Compile("(http(s)?:\\/\\/)([\\w.-]+)(\\.[\\w\\.-]+)+(\\/)?([\\w\\-\\._~:\\/?#[\\]@!\\$&'\\(\\)\\*\\+,;=.]+)$")
	isInternal, _ := regexp.Compile("(http(s)?:\\/\\/)" + r.domainUrl + "(.*)")

	for _, element := range links {
		//If String is valid url
		//Anchor are excluded
		if isValid.MatchString(element) {
			if isInternal.MatchString(element) {
				internal = append(internal, element)
			} else {
				external = append(external, element)
			}
		}
	}
	return external, internal
}
