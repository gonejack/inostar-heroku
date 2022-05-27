package kit

import (
	"fmt"
	"io"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

	"github.com/gonejack/inostar-heroku/config"
	"github.com/gonejack/inostar-heroku/model"
	"github.com/gonejack/inostar-heroku/module/dropbox"
	"github.com/gonejack/inostar-heroku/util"
)

func SaveAsEmail(a *model.Article) (meta *files.FileMetadata, err error) {
	email := model.NewEmail(config.EmailFrom, config.EmailTo, a.Title, model.NewHTML(a))
	name := email.Filename()
	if config.EmailZip {
		name = name + ".gz"
	}
	if dropbox.Exist(name) {
		return nil, fmt.Errorf("file exist")
	}
	body := email.Build()
	if config.EmailZip {
		body = util.NewZipper(body)
	}
	meta, err = dropbox.Upload(name, 0, body)
	if err == nil {
		err = email.RenderErr()
	}
	return
}

func SaveAsHTML(a *model.Article) (f *files.FileMetadata, err error) {
	html := model.NewHTML(a)
	dat, _ := html.Render()
	return dropbox.Upload(html.Filename(), int64(len(dat)), io.NopCloser(strings.NewReader(dat)))
}
