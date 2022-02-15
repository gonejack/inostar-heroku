package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	io.Copy(io.Discard, c.Request.Body)
	c.String(http.StatusOK, "I am running")
}
