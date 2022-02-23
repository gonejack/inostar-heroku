package general

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/gonejack/inostar-heroku/module/dropbox"
)

func Test(c *gin.Context) {
	_, err := dropbox.Upload(c.Query("name"), c.Request.ContentLength, c.Request.Body)
	if err == nil {
		c.String(http.StatusOK, "done")
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

var gif, _ = base64.RawStdEncoding.DecodeString("R0lGODlhAQABAJAAAP8AAAAAACH5BAUQAAAALAAAAAABAAEAAAICBAEAOw==")

func Debug(c *gin.Context) {
	logrus.Infof("接收到请求: %s: %s", c.Request.RemoteAddr, c.Request.RequestURI)
	c.Data(http.StatusOK, "image/gif", gif)
}
