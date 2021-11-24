package config

import (
	"os"
	"path"

	"golang.org/x/oauth2"
)

var OAuth2 = &oauth2.Config{
	ClientID:     os.Getenv("INOREADER_CLIENT_ID"),
	ClientSecret: os.Getenv("INOREADER_CLIENT_SECRET"),
	Scopes:       []string{"read"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://www.inoreader.com/oauth2/auth",
		TokenURL: "https://www.inoreader.com/oauth2/token",
	},
	RedirectURL: path.Join(os.Getenv("HOST"), "/oauth2/user_info"),
}
