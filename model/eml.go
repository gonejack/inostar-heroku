package model

import (
	"context"
	"crypto/md5"
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
func (e *Email) Render() io.ReadCloser {
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
	html, _ := e.htm.Render()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logrus.Errorf("cannot parse html %s: %s", e.htm, err)
		return
	}

	var media []textproto.MIMEHeader
	var cids = make(map[string]bool)
	doc.Find("img,video,source").Each(func(i int, e *goquery.Selection) {
		src, _ := e.Attr("src")
		switch {
		case strings.HasPrefix(src, "http://"):
			fallthrough
		case strings.HasPrefix(src, "https://"):
			cid := md5str(src)
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

	html, _ = doc.Html()
	e.writeHTML(html)

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
	timeout, cancel := context.WithTimeout(context.TODO(), time.Minute*3)
	defer cancel()

	ref := h.Get("content-location")

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		start := time.Now()
		logrus.Debugf("get media %s", ref)
		defer func() {
			logrus.Debugf("get media %s done: %s", ref, time.Now().Sub(start).Round(time.Millisecond))
		}()
	}

	rsp, err := e.request(timeout, ref)
	if err != nil {
		logrus.Errorf("downolad %s failed: %s", ref, err)
		return
	}
	defer func() {
		io.Copy(io.Discard, rsp.Body)
		rsp.Body.Close()
	}()

	h.Set("content-type", Fallback(rsp.Header.Get("content-type"), "application/octet-stream"))
	h.Set("content-disposition", "inline")
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

func Fallback(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
func md5str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
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
