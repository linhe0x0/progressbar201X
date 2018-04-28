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
		log.WithError(err).Error("get progress of this year")
		return
	}

	artile, err := progressbar201X.NewArticle(2018, progress)

	if err != nil {
		log.WithError(err).Error("create article")
		return
	}

	mediaId, err := progressbar201X.UploadArticle(artile)

	if err != nil {
		log.WithError(err).Error("upload article")
		return
	}

	log.Info("upload article successfully, the article's mediaId is " + mediaId)

	err = progressbar201X.BetchPostArticle(mediaId)

	if err != nil {
		log.WithError(err).Error("send article")
		return
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

	log.Info("timed task: 0 41 1 * * *")

	// The server is using UTC time.
	// A fixed 8 hour UTC offset is used in China.
	// In order to perform the task at 09:00am, Beijing time, it needs to be fixed.
	c.AddFunc("0 41 1 * * *", func() {
		broadcast()
	})

	c.Start()

	progressbar201X.StartServer()
}
