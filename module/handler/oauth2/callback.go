package oauth2

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/gonejack/inostar-heroku/config"
	"github.com/gonejack/inostar-heroku/module/worker"
	"github.com/gonejack/inostar-heroku/util"
)

func Callback(c *gin.Context) {
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

	worker.Reset(tok)
}
