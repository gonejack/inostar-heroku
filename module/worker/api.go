package worker

import (
	"encoding/json"

	"golang.org/x/oauth2"

	"github.com/gonejack/inostar-heroku/config"
)

var w worker

func InitToken() {
	if config.Token == "" {
		return
	}
	var token oauth2.Token
	e := json.Unmarshal([]byte(config.Token), &token)
	if e == nil {
		Reset(&token)
	}
}
func Reset(tok *oauth2.Token) {
	w.reset(tok)
}
