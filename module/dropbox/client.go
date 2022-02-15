package dropbox

import (
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

	"github.com/gonejack/inostar-heroku/config"
)

var client = files.New(config.Dropbox)
