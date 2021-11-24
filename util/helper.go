package util

import (
	"os"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

	"github.com/gonejack/inostar-heroku/dbx"
	"github.com/gonejack/inostar-heroku/model"
)

func SaveArticle(a *model.Article) (f *files.FileMetadata, err error) {
	htm := model.NewHTML(a)
	eml := model.NewEmail(os.Getenv("EML_FROM"), os.Getenv("EML_TO"), a.Title, htm)
	name := eml.Filename()
	datp := eml.Render()
	if os.Getenv("EML_ZIP") == "1" {
		name = name + ".gz"
		datp = dbx.NewZipper(datp)
	}
	return dbx.Upload(name, 0, datp)
}
