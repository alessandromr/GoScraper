package LinkSystem

import (
	"regexp"
	"strings"
)

type LinkSystem struct {
}

/**
* Divide link by external and internal
* @parameter []string, string
* @return map[string]int, map[string]int
 */
func (l LinkSystem) DivideByType(links []string, domainUrl string) ([]string, []string) {
	var external = make(map[string]int)
	var internal = make(map[string]int)

	isValid, _ := regexp.Compile("(http(s)?:\\/\\/)([\\w.-]+)(\\.[\\w\\.-]+)+(\\/)?([\\w\\-\\._~:\\/?#[\\]@!\\$&'\\(\\)\\*\\+,;=.]+)$")
	isInternal, _ := regexp.Compile("(http(s)?:\\/\\/)(([\\w\\.-]+)?)" + domainUrl + "(.*)")

	for _, element := range links {
		element = strings.TrimSuffix(element, "/")
		if isValid.MatchString(element) {
			if isInternal.MatchString(element) {
				if _, ok := internal[element]; !ok {
					internal[element] = 1
				}
			} else {
				if _, ok := external[element]; !ok {
					external[element] = 1
				}
			}
		}
	}
	return l.mapToSlice(external), l.mapToSlice(internal)
}

/**
* Clean Links
* @parameter []string
* @return []string
 */
func (l LinkSystem) Clean(links []string) []string {
	for _, element := range links {
		element = strings.TrimSuffix(element, "/")
	}
	return links
}

/**
* Get DomainUrl from given start url
* @parameter
* @return
 */
func (l LinkSystem) GetDomainUrl(startUrl string) string {
	regex, _ := regexp.Compile("(http(s)?:\\/\\/)([\\w.-]+)(\\.[\\w\\.-]+)")
	found := regex.FindStringSubmatch(startUrl)
	return found[3] + found[4]
}

/**
* Check if link is not subdomain
* @parameter []string, []string
* @return []string
 */
func (l LinkSystem) IsSubdomain(external []string) ([]string, []string) {
	var validExternal []string
	var invalid []string
	isSubdomain, _ := regexp.Compile("(http(s)?:\\/\\/)([a-z0-9]+.{1}[a-z0-9]+)(\\/?)((\\/[a-z0-9]+)?)$")

	for _, element := range external {
		if isSubdomain.MatchString(element) {
			validExternal = append(validExternal, element)
		} else {
			invalid = append(invalid, element)
		}
	}
	return validExternal, invalid
}

/**
* Convert map to slice
* @parameter map[string]int
* @return []string
 */
func (l LinkSystem) mapToSlice(m map[string]int) []string {
	var slice []string
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}
