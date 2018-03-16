package main

import (
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/robfig/cron"

	"github.com/sqrthree/progressbar201X"
	. "github.com/sqrthree/progressbar201X/internal/config"
)

func broadcast() {
	progress, err := progressbar201X.GetProgressOfCurrentYear()

	if err != nil {
		log.WithError(err).Fatal("get progress of this year")
	}

	artile := progressbar201X.NewArticle(2018, progress)

	mediaId, err := progressbar201X.UploadArticle(artile)

	if err != nil {
		log.WithError(err).Fatal("upload article")
	}

	log.Info("upload article successfully, the article's mediaId is " + mediaId)

	err = progressbar201X.BetchPostArticle(mediaId)

	if err != nil {
		log.WithError(err).Fatal("send article")
	}

	log.Infof("Article %s has been sent.\n", mediaId)
}

func main() {
	logLevel := log.InfoLevel

	if Config.App.Debug {
		logLevel = log.DebugLevel
	}

	log.SetHandler(text.Default)
	log.SetLevel(logLevel)

	c := cron.New()

	c.AddFunc("0 41 9 * * *", func() {
		broadcast()
	})

	c.Start()

	select {}

	// progressbar201X.StartServer()
}
