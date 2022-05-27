package general

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/gonejack/inostar-heroku/module/dropbox"
)

var gif, _ = base64.RawStdEncoding.DecodeString("R0lGODlhAQABAJAAAP8AAAAAACH5BAUQAAAALAAAAAABAAEAAAICBAEAOw==")

func Debug(c *gin.Context) {
	logrus.Infof("接收到请求: %s: %s", c.Request.RemoteAddr, c.Request.RequestURI)
	c.Data(http.StatusOK, "image/gif", gif)
}

func Test(c *gin.Context) {
	logrus.Infof("接收到请求: %s: %s", c.Request.RemoteAddr, c.Request.RequestURI)
	name := c.Query("name")
	logrus.Infof("请求文件名: %s", name)
	if dropbox.Exist(name) {
		c.String(http.StatusOK, "exist")
	} else {
		c.String(http.StatusOK, "not exist")
	}
}
