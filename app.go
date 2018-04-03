package progressbar201X

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/apex/log"

	. "github.com/sqrthree/progressbar201X/internal/config"
	"github.com/sqrthree/progressbar201X/internal/timeline"
	"github.com/sqrthree/progressbar201X/internal/wechat"
)

var (
	appId     = Config.Wechat.AppId
	appSecret = Config.Wechat.AppSecret
)

var (
	accessTokenServer wechat.AccessTokenServer = wechat.NewDefautlAccessToeknServer(appId, appSecret)
	wechatClient      *wechat.Client           = wechat.NewClient(accessTokenServer)
)

// Handle the request.
func handle(w http.ResponseWriter, r *http.Request) {
	for _, route := range Routes {
		if route.url == r.URL.Path {
			route.handle(w, r)
			return
		}
	}

	http.NotFound(w, r)
}

// Start func starts a server to handle requests.
func StartServer() {
	port := strconv.FormatUint(Config.Server.Port, 10)

	if port == "" {
		fmt.Println("Port is invaild.")
		return
	}

	server := http.Server{
		Addr:         ":" + port,
		Handler:      http.HandlerFunc(handle),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	fmt.Printf("Server is running at http://127.0.0.1:%s\n", port)

	err := server.ListenAndServe()

	if err != nil {
		fmt.Println("error:", err)
	}
}

func GetProgressOfCurrentYear() (progress float64, err error) {
	now := time.Now().UTC().Add(8 * time.Hour)

	progress, err = timeline.NewWithYear(now)
	return
}

type Article struct {
	Title   string
	Content string
}

func NewArticle(year int, progress float64) *Article {
	p := math.Floor(progress * 100)

	log.Debugf("create article with progress value [%v]", p)

	article := Article{
		Title:   fmt.Sprintf("%v 年已经过去了 %v%s 啦", year, p, "%"),
		Content: fmt.Sprintf("%v 年已经过去了 %v%s 啦", year, p, "%"),
	}

	log.WithFields(log.Fields{
		"title":   article.Title,
		"content": article.Content,
	}).Debug("new article")

	return &article
}

func UploadArticle(article *Article) (mediaId string, err error) {
	material, err := wechat.GetRandomImageMaterial(wechatClient)

	if err != nil {
		return
	}

	newArticle := wechat.ArticleMaterial{
		ThumbMediaId: material.MediaId,
		Title:        article.Title,
		Content:      article.Content,
	}

	log.WithFields(log.Fields{
		"thumb_media_id": newArticle.ThumbMediaId,
		"title":          newArticle.Title,
		"content":        newArticle.Content,
	}).Info("create new article")

	mediaId, err = wechat.UploadArticleMaterial(wechatClient, &newArticle)

	return
}

func BetchPostArticle(mediaId string) error {
	err := wechat.BetchPostArticle(wechatClient, mediaId)

	if err != nil {
		return err
	}

	return nil
}
