package webhook

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Dropbox(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		c.Header("content-type", "text/plain")
		c.Header("X-Content-Type-Options", "nosniff")
		c.String(http.StatusOK, c.Query("challenge"))
	default:
		io.Copy(io.Discard, c.Request.Body)
		c.String(http.StatusOK, "ok")
	}
}
