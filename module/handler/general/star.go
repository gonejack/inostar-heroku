package general

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/gonejack/inostar-heroku/model"
	"github.com/gonejack/inostar-heroku/module/kit"
)

func Star(c *gin.Context) {
	logrus.Debugf("processing %s", c.Request.URL)

	var art model.Article
	err := art.From(c.Request.Body)
	if err == nil {
		go kit.SaveArticle(&art)
		c.String(http.StatusOK, "ok")
	} else {
		logrus.Errorf("parse json failed: %s", err)
		c.String(http.StatusBadRequest, err.Error())
	}
}
