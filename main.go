package main

import (
	"fmt"
	"os"

	"github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/gonejack/inostar-heroku/handler"
	"github.com/gonejack/inostar-heroku/reader"
	"github.com/gonejack/inostar-heroku/util"
)

func init() {
	logrus.SetFormatter(&formatter.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        true,
		HideKeys:        false,
		CallerFirst:     true,
	})
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err == nil {
		logrus.SetLevel(level)
	}
	if !logrus.IsLevelEnabled(logrus.DebugLevel) {
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	server := gin.New()
	server.Use(gin.Recovery())

	basic := server.Group("/")
	{
		basic.GET("/", handler.Hello)
		basic.POST("star", handler.Star)
		basic.POST("test", handler.Test)
	}
	oauth := server.Group("/oauth2")
	{
		oauth.GET("user_login", handler.UserLogin)
		oauth.GET("user_info", handler.UserInfo)
	}

	reader.ResetByEnv()

	err := server.Run(fmt.Sprintf(":%s", util.Env("PORT", "8080")))
	if err != nil {
		logrus.Fatal(err)
	}
}
