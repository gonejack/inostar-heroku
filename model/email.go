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
	headers textproto.MIMEHeader
	html    *HTML
	build   struct {
		writer *multipart.Writer
	}
	output struct {
		reader io.ReadCloser
		writer io.WriteCloser
	}
	renderError error
}

func (e *Email) Filename() string {
	return strings.TrimSuffix(e.html.Filename(), ".html") + ".embed.eml"
}
func (e *Email) Build() io.ReadCloser {
	go e.render()
	return e.output.reader
}
func (e *Email) RenderErr() error {
	return e.renderError
}

func (e *Email) render() {
	defer e.output.writer.Close()
	defer e.build.writer.Close()

	e.renderHeader()
	e.renderContent()
}
func (e *Email) renderHeader() {
	e.writeHeader(e.headers)
	e.write("\r\n")
}
func (e *Email) renderContent() {
	dat, _ := e.html.Render()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(dat))
	if err != nil {
		logrus.Errorf("cannot parse html %s: %s", e.html, err)
		return
	}

	var listOfMedia []textproto.MIMEHeader
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

			headerOfMedia := textproto.MIMEHeader{}
			headerOfMedia.Set("content-id", fmt.Sprintf("<%s>", cid))
			headerOfMedia.Set("content-location", src)
			listOfMedia = append(listOfMedia, headerOfMedia)
		}
	})
	doc.Find("iframe").Each(func(i int, iframe *goquery.Selection) {
		src, _ := iframe.Attr("src")
		if src == "" {
			return
		}
		a := &html.Node{
			Type: html.ElementNode,
			Data: atom.A.String(),
			Attr: []html.Attribute{{Key: atom.Src.String(), Val: src}},
		}
		a.AppendChild(&html.Node{Type: html.TextNode, Data: src})
		iframe.ReplaceWithNodes(a)
	})

	dat, _ = doc.Html()
	e.writeHTML(dat)

	for _, mediaHeader := range listOfMedia {
		err := e.writeMedia(mediaHeader)
		if err != nil {
			e.renderError = err
			break
		}
	}
}

func (e *Email) write(a ...interface{}) {
	fmt.Fprint(e.output.writer, a...)
}
func (e *Email) writeHeader(header textproto.MIMEHeader) {
	for field, values := range header {
		for _, value := range values {
			e.write(field, ": ")
			switch {
			case field == "Content-Type" || field == "Content-Disposition":
				e.write(value)
			case field == "From" || field == "To" || field == "Cc" || field == "Bcc":
				participants := strings.Split(value, ",")
				for i, v := range participants {
					addr, err := mail.ParseAddress(v)
					if err != nil {
						continue
					}
					participants[i] = addr.String()
				}
				e.write(strings.Join(participants, ", "))
			default:
				e.write(mime.QEncoding.Encode("utf-8", value))
			}
			e.write("\r\n")
		}
	}
}
func (e *Email) writeBase64(writer io.Writer, reader io.Reader) (err error) {
	printer := linesprinter.NewLinesPrinter(writer, 76, []byte("\r\n"))
	defer printer.Close()
	encoder := base64.NewEncoder(base64.StdEncoding, printer)
	defer encoder.Close()
	_, err = io.Copy(encoder, reader)
	if err != nil {
		err = fmt.Errorf("copy failed: %s", err)
	}
	return
}
func (e *Email) writeHTML(html string) {
	header := textproto.MIMEHeader{}
	header.Set("content-type", "text/html; charset=utf-8")
	header.Set("content-transfer-encoding", "base64")

	partWriter, _ := e.build.writer.CreatePart(header)
	e.writeBase64(partWriter, strings.NewReader(html))
}
func (e *Email) writeMedia(header textproto.MIMEHeader) error {
	ref := header.Get("content-location")

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		start := time.Now()
		logrus.Debugf("get media %s", ref)
		defer func() {
			logrus.Debugf("get media %s done: %s", ref, time.Now().Sub(start).Round(time.Millisecond))
		}()
	}

	retry := 0
request:
	timeout, cancel := context.WithTimeout(context.TODO(), time.Minute*5)
	rsp, err := e.request(timeout, ref)
	if err == nil {
		defer cancel()
	} else {
		cancel()
		if retry += 1; retry < 3 {
			goto request
		}
		return fmt.Errorf("downolad %s failed: %s", ref, err)
	}

	defer func() {
		io.Copy(io.Discard, rsp.Body)
		rsp.Body.Close()
	}()

	header.Set("content-type", util.Fallback(rsp.Header.Get("content-type"), "application/octet-stream"))
	header.Set("content-disposition", fmt.Sprintf(`inline; filename="%s"`, filename(rsp)))
	header.Set("content-transfer-encoding", "base64")

	partWriter, _ := e.build.writer.CreatePart(header)
	return e.writeBase64(partWriter, rsp.Body)
}
func (e *Email) request(ctx context.Context, src string) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, src, nil)
	if err != nil {
		return
	}

	req.Header.Set("referer", e.html.article.Href)
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:94.0) Gecko/20100101 Firefox/94.0")

	return client.Do(req.WithContext(ctx))
}

func NewEmail(from string, to string, subject string, html *HTML) (email *Email) {
	email = new(Email)

	email.output.reader, email.output.writer = io.Pipe()
	email.build.writer = multipart.NewWriter(email.output.writer)

	email.headers = textproto.MIMEHeader{}
	email.headers.Set("Mime-Version", "1.0")
	email.headers.Set("Content-Type", "multipart/related;\r\n boundary="+email.build.writer.Boundary())
	email.headers.Set("Subject", subject)
	email.headers.Set("From", from)
	email.headers.Set("To", to)
	email.headers.Set("Date", html.article.StarTime().Format(time.RFC1123Z))
	id, err := messageId()
	if err == nil {
		email.headers.Set("Message-Id", id)
	}
	email.html = html

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
