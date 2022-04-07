package donwload

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gonejack/inostar-heroku/module/dropbox"
	"github.com/gonejack/inostar-heroku/util"
)

func Download(c *gin.Context) {
	var download struct {
		URL string `json:"url"`
	}

	err := c.BindJSON(&download)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	req, err := http.NewRequest(http.MethodGet, download.URL, nil)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:94.0) Gecko/20100101 Firefox/94.0")
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, "saving")
	go func() {
		defer rsp.Body.Close()
		dropbox.Upload(util.MD5(download.URL), rsp.ContentLength, rsp.Body)
	}()
}
