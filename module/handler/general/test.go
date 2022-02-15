package general

import (
	"net/http"

	"github.com/gin-gonic/gin"

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
