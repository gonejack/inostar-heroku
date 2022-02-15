package kit

import (
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

	"github.com/gonejack/inostar-heroku/config"
	"github.com/gonejack/inostar-heroku/model"
	"github.com/gonejack/inostar-heroku/module/dropbox"
	"github.com/gonejack/inostar-heroku/util"
)

func SaveArticle(a *model.Article) (f *files.FileMetadata, err error) {
	html := model.NewHTML(a)
	email := model.NewEmail(config.EmailFrom, config.EmailTo, a.Title, html)
	name, content := email.Filename(), email.Build()
	if config.EmailZip {
		name, content = name+".gz", util.NewZipper(content)
	}
	return dropbox.Upload(name, 0, content)
}
