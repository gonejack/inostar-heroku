package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gonejack/inostar-heroku/dbx"
)

func Test(c *gin.Context) {
	_, err := dbx.Upload(c.Query("name"), c.Request.ContentLength, c.Request.Body)
	if err == nil {
		c.String(http.StatusOK, "done")
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
