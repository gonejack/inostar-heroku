package model

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"
)

var client http.Client

type HTML struct {
	a *Article
}

func (h *HTML) Filename() string {
	feed := maxLen(h.a.FeedTitle, 30)
	item := maxLen(h.a.Title, 30)
	return safetyName(fmt.Sprintf("[%s][%s][%s].html", feed, h.a.StarTime().Format("2006-01-02 15.04.05"), item))
}
func (h *HTML) Render() (htm string, err error) {
	content := h.content()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return
	}

	h.cleanDoc(doc)

	meta := fmt.Sprintf(`<meta name="inostar:publish" content="%s"/>`, h.a.StarTime().Format(time.RFC1123Z))
	doc.Find("head").AppendHtml(meta)
	doc.Find("head").AppendHtml(`<meta charset="utf-8"/>`)
	if doc.Find("title").Length() == 0 {
		doc.Find("head").AppendHtml("<title></title>")
	}
	if doc.Find("title").Text() == "" {
		doc.Find("title").SetText(h.a.Title)
	}

	return doc.Html()
}

func (h *HTML) content() string {
	content := fmt.Sprintf("%s %s %s", h.contentHeader(), h.a.Article, h.contentFooter())

	return gohtml.Format(content)
}
func (h *HTML) contentHeader() string {
	const tpl = `
<p>
	<a title="Published: {published}" href="{link}" style="display:block; color: #000; padding-bottom: 10px; text-decoration: none; font-size:1em; font-weight: normal;">
		<span style="display: block; color: #666; font-size:1.0em; font-weight: normal;">{origin}</span>
		<span style="font-size: 1.5em;">{title}</span>
	</a>
</p>`

	rpl := strings.NewReplacer(
		"{link}", html.EscapeString(h.a.Href),
		"{origin}", html.EscapeString(h.a.FeedTitle),
		"{published}", h.a.StarTime().Format("2006-01-02 15:04:05"),
		"{title}", html.EscapeString(h.a.Title),
	)

	return rpl.Replace(tpl)
}
func (h *HTML) contentFooter() string {
	const tpl = `
<br/><br/>
<a style="display: block; display: inline-block; border-top: 1px solid #ccc; padding-top: 5px; color: #666; text-decoration: none;"
   href="{link}">{link_text}</a>
<p style="color:#999;">Save with <a style="color:#666; text-decoration:none; font-weight: bold;" 
									href="https://github.com/gonejack/inostar">inostar</a>
</p>`

	text, err := url.QueryUnescape(h.a.Href)
	if err != nil {
		text = h.a.Href
	}
	rpl := strings.NewReplacer(
		"{link}", html.EscapeString(h.a.Href),
		"{link_text}", html.EscapeString(text),
	)

	return rpl.Replace(tpl)
}
func (h *HTML) cleanDoc(doc *goquery.Document) *goquery.Document {
	// remove inoreader ads
	doc.Find("body").Find(`div:contains("ads from inoreader")`).Closest("center").Remove()

	// remove solidot.org ads
	doc.Find("img[src='https://img.solidot.org//0/446/liiLIZF8Uh6yM.jpg']").Remove()

	// remove 36kr ads
	doc.Find("img[src='https://img.36krcdn.com/20191024/v2_1571894049839_img_jpg']").Closest("p").Remove()

	// remove zaobao ads
	doc.Find("img[src='https://www.zaobao.com.sg/themes/custom/zbsg2020/images/default-img.png']").Closest("p").Remove()

	// remove cnbeta ads
	doc.Find(`strong:contains("访问：")`).Closest("div").Remove()

	// remove empty div
	doc.Find("div:empty").Remove()

	// fix image
	doc.Find("img,video").Each(func(i int, img *goquery.Selection) {
		w, _ := img.Attr("width")
		if w == "0" {
			img.RemoveAttr("width")
		}
		h, _ := img.Attr("height")
		if h == "0" {
			img.RemoveAttr("height")
		}
	})

	return doc
}

func NewHTML(a *Article) *HTML {
	return &HTML{a}
}

func safetyName(name string) string {
	return regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(name, ".")
}
func maxLen(str string, max int) string {
	var rs []rune
	for i, r := range []rune(str) {
		if i >= max {
			if i > 0 {
				rs = append(rs, '.', '.', '.')
			}
			break
		}
		rs = append(rs, r)
	}
	return string(rs)
}
