package model

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
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
	a.fullBody()
}

func (a *Article) fullBody() {
	u, err := url.Parse(a.Href)
	if err != nil {
		logrus.Errorf("cannot parse link %s", a.Href)
		return
	}

	switch {
	case strings.HasSuffix(u.Host, "sspai.com"):
		full, err := a.grabDoc()
		if err != nil {
			logrus.Errorf("cannot parse content from %s", a.Href)
			return
		}
		ct := full.Find("div.article-body div.content").First()
		ct.Find("*").RemoveAttr("style").RemoveAttr("class")
		htm, err := ct.Html()
		if err != nil {
			logrus.Errorf("cannot generate content of %s", a.Href)
			return
		}
		a.Article = htm
	case strings.HasSuffix(u.Host, "leimao.github.io"):
		full, err := a.grabDoc()
		if err != nil {
			logrus.Errorf("cannot parse content from %s", a.Href)
			return
		}
		ct := full.Find("article.article div.content").First()
		ct.Find("*").RemoveAttr("style").RemoveAttr("class")
		htm, err := ct.Html()
		if err != nil {
			logrus.Errorf("cannot generate content of %s", a.Href)
			return
		}
		a.Article = htm
	case strings.HasSuffix(u.Host, "thoughtworks.cn"):
		full, err := a.grabDoc()
		if err != nil {
			logrus.Errorf("cannot parse content from %s", a.Href)
			return
		}
		ct := full.Find("article.post div.entry-wrap").First()
		ct.Find("*").RemoveAttr("style").RemoveAttr("class")
		htm, err := ct.Html()
		if err != nil {
			logrus.Errorf("cannot generate content of %s", a.Href)
			return
		}
		a.Article = htm
	case strings.HasSuffix(u.Host, "huxiu.com"):
		full, err := a.grabDoc()
		if err != nil {
			logrus.Errorf("cannot parse content from %s", a.Href)
			return
		}
		js := full.Find("div.js-video-play-log-report-wrap script").Text()
		if js == "" {
			return
		}
		ms := regexp.MustCompile(`'(https://.*video\.huxiucdn\.com/[^']+)'`).FindStringSubmatch(js)
		if len(ms) > 0 {
			tpl := `<video autoplay controls width="100%%"><source src="%s" type="video/mp4"></video>`
			video := fmt.Sprintf(tpl, ms[1])
			a.Article = video + a.Article
		}
	}
}
func (a *Article) grabDoc() (doc *goquery.Document, err error) {
	timeout, cancel := context.WithTimeout(context.TODO(), time.Second*15)
	defer cancel()

	req, err := http.NewRequestWithContext(timeout, http.MethodGet, a.Href, nil)
	if err != nil {
		return
	}
	req.Header.Set("referer", a.Href)
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:94.0) Gecko/20100101 Firefox/94.0")

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("cannot grab link %s", a.Href)
		return
	}
	defer rsp.Body.Close()
	dat, err := io.ReadAll(rsp.Body)
	if err != nil {
		return
	}
	htm := strings.ReplaceAll(string(dat), "<!--!-->", "")
	return goquery.NewDocumentFromReader(strings.NewReader(htm))
}
func (a *Article) shouldUnescape(s string) bool {
	return regexp.MustCompile(`(&#\d{2,6};){2,}`).MatchString(s)
}
