package reader

import (
	"encoding/json"
	"os"

	"golang.org/x/oauth2"
)

var r reader

func ResetByEnv() {
	js := os.Getenv("TOKEN")
	if js == "" {
		return
	}
	var tok oauth2.Token
	err := json.Unmarshal([]byte(js), &tok)
	if err == nil {
		Reset(&tok)
	}
}

func Reset(tok *oauth2.Token) {
	r.reset(tok)
}
