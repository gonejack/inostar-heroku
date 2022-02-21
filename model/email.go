package model

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"math/big"
	"mime"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gonejack/linesprinter"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/gonejack/inostar-heroku/util"
)

type Email struct {
	pr io.ReadCloser
	pw io.WriteCloser
	mw *multipart.Writer
	h  textproto.MIMEHeader

	htm *HTML

	boundary string
}

func (e *Email) Filename() string {
	return strings.TrimSuffix(e.htm.Filename(), ".html") + ".embed.eml"
}
func (e *Email) Build() io.ReadCloser {
	go e.render()
	return e.pr
}

func (e *Email) render() {
	defer e.pw.Close()
	defer e.mw.Close()

	e.renderHeader()
	e.renderContent()
}
func (e *Email) renderHeader() {
	e.writeHeader(e.h)
	e.write("\r\n")
}
func (e *Email) renderContent() {
	htm, _ := e.htm.Render()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htm))
	if err != nil {
		logrus.Errorf("cannot parse html %s: %s", e.htm, err)
		return
	}

	var media []textproto.MIMEHeader
	var cids = make(map[string]bool)
	doc.Find("img,video,source").Each(func(i int, e *goquery.Selection) {
		src, _ := e.Attr("src")
		e.SetAttr("data-inostar-src", src)
		switch {
		case src == "":
			return
		case strings.HasPrefix(src, "http://"):
			fallthrough
		case strings.HasPrefix(src, "https://"):
			cid := util.MD5(src)
			u, err := url.Parse(src)
			if err == nil {
				cid += filepath.Ext(u.Path)
			}
			e.SetAttr("src", fmt.Sprintf("cid:%s", cid))

			if cids[cid] {
				return
			}
			cids[cid] = true

			h := textproto.MIMEHeader{}
			h.Set("content-id", fmt.Sprintf("<%s>", cid))
			h.Set("content-location", src)
			media = append(media, h)
		}
	})
	doc.Find("iframe").Each(func(i int, iframe *goquery.Selection) {
		src, _ := iframe.Attr("src")
		if src == "" {
			return
		}
		atag := &html.Node{
			Type: html.ElementNode,
			Data: atom.A.String(),
			Attr: []html.Attribute{{Key: atom.Src.String(), Val: src}},
		}
		atag.AppendChild(&html.Node{Type: html.TextNode, Data: src})
		iframe.ReplaceWithNodes(atag)
	})

	htm, _ = doc.Html()
	e.writeHTML(htm)

	for _, m := range media {
		e.writeMedia(m)
	}
}

func (e *Email) write(a ...interface{}) {
	fmt.Fprint(e.pw, a...)
}
func (e *Email) writeHeader(header textproto.MIMEHeader) {
	for f, vs := range header {
		for _, v := range vs {
			e.write(f, ": ")
			switch {
			case f == "Content-Type" || f == "Content-Disposition":
				e.write(v)
			case f == "From" || f == "To" || f == "Cc" || f == "Bcc":
				participants := strings.Split(v, ",")
				for i, v := range participants {
					addr, err := mail.ParseAddress(v)
					if err != nil {
						continue
					}
					participants[i] = addr.String()
				}
				e.write(strings.Join(participants, ", "))
			default:
				e.write(mime.QEncoding.Encode("utf-8", v))
			}
			e.write("\r\n")
		}
	}
}
func (e *Email) writeBase64(w io.Writer, r io.Reader) {
	prt := linesprinter.NewLinesPrinter(w, 76, []byte("\r\n"))
	enc := base64.NewEncoder(base64.StdEncoding, prt)
	_, err := io.Copy(enc, r)
	if err != nil {
		logrus.Errorf("copy failed: %s", err)
	}
	enc.Close()
	prt.Close()
}
func (e *Email) writeHTML(html string) {
	h := textproto.MIMEHeader{}
	h.Set("content-type", "text/html; charset=utf-8")
	h.Set("content-transfer-encoding", "base64")

	pw, _ := e.mw.CreatePart(h)
	e.writeBase64(pw, strings.NewReader(html))
}
func (e *Email) writeMedia(h textproto.MIMEHeader) {
	ref := h.Get("content-location")

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		start := time.Now()
		logrus.Debugf("get media %s", ref)
		defer func() {
			logrus.Debugf("get media %s done: %s", ref, time.Now().Sub(start).Round(time.Millisecond))
		}()
	}

	retry := 0
request:
	timeout, cancel := context.WithTimeout(context.TODO(), time.Minute*3)
	rsp, err := e.request(timeout, ref)
	if err == nil {
		defer cancel()
	} else {
		cancel()
		if retry += 1; retry < 3 {
			goto request
		}
		logrus.Errorf("downolad %s failed: %s", ref, err)
		return
	}

	defer func() {
		io.Copy(io.Discard, rsp.Body)
		rsp.Body.Close()
	}()

	h.Set("content-type", util.Fallback(rsp.Header.Get("content-type"), "application/octet-stream"))
	h.Set("content-disposition", fmt.Sprintf(`inline; filename="%s"`, filename(rsp)))
	h.Set("content-transfer-encoding", "base64")

	pw, _ := e.mw.CreatePart(h)
	e.writeBase64(pw, rsp.Body)
}
func (e *Email) request(ctx context.Context, src string) (resp *http.Response, err error) {
	r, err := http.NewRequest(http.MethodGet, src, nil)
	if err != nil {
		return
	}

	r.Header.Set("referer", e.htm.a.Href)
	r.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:94.0) Gecko/20100101 Firefox/94.0")

	return client.Do(r.WithContext(ctx))
}

func NewEmail(from string, to string, subject string, html *HTML) (e *Email) {
	e = new(Email)
	e.pr, e.pw = io.Pipe()
	e.mw = multipart.NewWriter(e.pw)

	e.h = textproto.MIMEHeader{}
	e.h.Set("Mime-Version", "1.0")
	e.h.Set("Content-Type", "multipart/related;\r\n boundary="+e.mw.Boundary())
	e.h.Set("Subject", subject)
	e.h.Set("From", from)
	e.h.Set("To", to)
	e.h.Set("Date", html.a.StarTime().Format(time.RFC1123Z))
	id, err := messageId()
	if err == nil {
		e.h.Set("Message-Id", id)
	}
	e.htm = html

	return
}

func messageId() (string, error) {
	t := time.Now().UnixNano()
	pid := os.Getpid()
	rint, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return "", err
	}
	h, err := os.Hostname()
	if err != nil {
		h = "localhost.localdomain"
	}
	return fmt.Sprintf("<%d.%d.%d@%s>", t, pid, rint, h), nil
}

func filename(rsp *http.Response) string {
	if cd := rsp.Header.Get("content-disposition"); cd != "" {
		_, ps, _ := mime.ParseMediaType(cd)
		if ps["filename"] != "" {
			return ps["filename"]
		}
	}

	name := util.MD5(rsp.Request.RequestURI)
	ext := ".dat"
	ct, _, _ := mime.ParseMediaType(rsp.Header.Get("content-type"))
	switch ct {
	case "":
	case "application/javascript", "application/x-javascript":
		ext = ".js"
	case "image/jpeg":
		ext = ".jpg"
	case "font/opentype":
		ext = ".otf"
	default:
		exs, _ := mime.ExtensionsByType(ct)
		if len(exs) > 0 {
			ext = exs[0]
		}
	}

	return name + ext
}
