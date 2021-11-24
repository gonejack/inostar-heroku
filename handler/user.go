package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/gonejack/inostar-heroku/config"
	"github.com/gonejack/inostar-heroku/reader"
	"github.com/gonejack/inostar-heroku/util"
)

func UserLogin(c *gin.Context) {
	c.Redirect(http.StatusFound, config.OAuth2.AuthCodeURL("state", oauth2.AccessTypeOnline))
}

func UserInfo(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		logrus.Errorf("empty code")
		c.String(http.StatusBadRequest, "empty code")
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute)
	defer cancel()

	tok, err := config.OAuth2.Exchange(ctx, code)
	if err != nil {
		logrus.Errorf("get token failed: %s", err)
		c.String(http.StatusBadGateway, err.Error())
		return
	}
	c.String(http.StatusOK, "got token %s", util.JsonDump(tok))

	reader.Reset(tok)
}
