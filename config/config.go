package config

import (
	"fmt"
	"os"
	"path"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"golang.org/x/oauth2"

	"github.com/gonejack/inostar-heroku/util"
)

var LogLevel = os.Getenv("LOG_LEVEL")
var Port = fmt.Sprintf(":%s", util.Fallback(os.Getenv("PORT"), "8080"))
var Token = os.Getenv("TOKEN")
var (
	EmailFrom = os.Getenv("EML_FROM")
	EmailTo   = os.Getenv("EML_TO")
	EmailZip  = os.Getenv("EML_ZIP") == "1"
)

var Dropbox = dropbox.Config{
	Token:    os.Getenv("DROPBOX_TOKEN"),
	LogLevel: dropbox.LogInfo, // if needed, set the desired logging level. Default is off
}

var OAuth2 = &oauth2.Config{
	ClientID:     os.Getenv("INOREADER_CLIENT_ID"),
	ClientSecret: os.Getenv("INOREADER_CLIENT_SECRET"),
	Scopes:       []string{"read"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://www.inoreader.com/oauth2/auth",
		TokenURL: "https://www.inoreader.com/oauth2/token",
	},
	RedirectURL: path.Join(os.Getenv("HOST"), "/oauth2/callback"),
}
