package oauth2

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/gonejack/inostar-heroku/config"
)

func Auth(c *gin.Context) {
	c.Redirect(http.StatusFound, config.OAuth2.AuthCodeURL("state", oauth2.AccessTypeOnline))
}
