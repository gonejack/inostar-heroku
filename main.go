package main

import (
	"github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/gonejack/inostar-heroku/config"
	"github.com/gonejack/inostar-heroku/module/handler"
	"github.com/gonejack/inostar-heroku/module/handler/donwload"
	"github.com/gonejack/inostar-heroku/module/handler/general"
	"github.com/gonejack/inostar-heroku/module/handler/oauth2"
	"github.com/gonejack/inostar-heroku/module/handler/webhook"
	"github.com/gonejack/inostar-heroku/module/worker"
)

func init() {
	logrus.SetFormatter(&formatter.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		NoColors:        true,
		HideKeys:        false,
		CallerFirst:     true,
	})
	level, err := logrus.ParseLevel(config.LogLevel)
	if err == nil {
		logrus.SetLevel(level)
	}
	if !logrus.IsLevelEnabled(logrus.DebugLevel) {
		gin.SetMode(gin.ReleaseMode)
	}
	worker.InitToken()
}
func main() {
	r := gin.New()
	{
		r.Use(gin.Recovery())
	}
	idx := r.Group("/")
	{
		idx.Any("/", handler.Index)
		idx.POST("star", general.Star)
		idx.Any("test", general.Test)
		idx.GET("masky_debug", general.Debug)
	}
	oauth := r.Group("/oauth2")
	{
		oauth.GET("auth", oauth2.Auth)
		oauth.GET("callback", oauth2.Callback)
	}
	hook := r.Group("/webhook")
	{
		hook.Any("dropbox", webhook.Dropbox)
	}
	down := r.Group("download")
	{
		down.Any("download", donwload.Download)
	}
	if e := r.Run(config.Port); e != nil {
		logrus.Fatal(e)
	}
}
