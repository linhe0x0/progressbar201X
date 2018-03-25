package main

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/logfmt"
	"github.com/robfig/cron"
	"github.com/sqrthree/debugfmt"

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
		log.SetHandler(debugfmt.New(os.Stdout))
	} else {
		log.SetHandler(logfmt.Default)
	}

	log.SetLevel(logLevel)

	c := cron.New()

	log.Info("timed task: 0 41 9 * * *")

	c.AddFunc("0 41 9 * * *", func() {
		broadcast()
	})

	c.Start()

	select {}

	// progressbar201X.StartServer()
}
