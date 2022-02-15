package model

import (
	"encoding/json"
	"html"
	"io"
	"regexp"
	"time"
)

type Article struct {
	FeedTitle   string `json:"feed_title"`
	Title       string `json:"title"`
	StarTimeRaw string `json:"star_time"`
	Href        string `json:"href"`
	Article     string `json:"article"`
}

func (a *Article) StarTime() time.Time {
	t, _ := time.Parse("01/02/2006 3:04:05 PM", a.StarTimeRaw)
	return t
}
func (a *Article) From(r io.ReadCloser) (err error) {
	defer r.Close()

	err = json.NewDecoder(r).Decode(a)
	if err == nil {
		a.Refine()
	}

	return
}
func (a *Article) Refine() {
	if a.shouldUnescape(a.Title) {
		a.Title = html.UnescapeString(a.Title)
	}
	if a.shouldUnescape(a.FeedTitle) {
		a.FeedTitle = html.UnescapeString(a.FeedTitle)
	}
}
func (a *Article) shouldUnescape(s string) bool {
	return regexp.MustCompile(`(&#\d{2,6};){2,}`).MatchString(s)
}
