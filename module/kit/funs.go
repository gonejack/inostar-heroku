package kit

import (
	"io"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

	"github.com/gonejack/inostar-heroku/config"
	"github.com/gonejack/inostar-heroku/model"
	"github.com/gonejack/inostar-heroku/module/dropbox"
	"github.com/gonejack/inostar-heroku/util"
)

func SaveAsEmail(a *model.Article) (f *files.FileMetadata, err error) {
	html := model.NewHTML(a)
	email := model.NewEmail(config.EmailFrom, config.EmailTo, a.Title, html)
	name, body := email.Filename(), email.Build()
	if config.EmailZip {
		name, body = name+".gz", util.NewZipper(body)
	}
	f, err = dropbox.Upload(name, 0, body)
	if err == nil {
		err = email.RenderErr()
	}
	return
}

func SaveAsHTML(a *model.Article) (f *files.FileMetadata, err error) {
	html := model.NewHTML(a)
	content, _ := html.Render()
	return dropbox.Upload(html.Filename(), int64(len(content)), io.NopCloser(strings.NewReader(content)))
}
