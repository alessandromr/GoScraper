package URLsQueue

import ()

type Queue struct {
	URLs []Url
	Qlen int
}

type Url struct {
	Domain string
	Clean  string
}

func (q Queue) Push(newUrl Url) []Url {
	q.URLs = append(q.URLs, newUrl)
	if len(q.URLs) > q.Qlen {
		q.URLs = q.URLs[1:]
	}
	return q.URLs
}
