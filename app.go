package progressbar201X

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"text/template"
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

const articleOptionsURL = "https://raw.githubusercontent.com/sqrthree/progressbar201X/quotations/main.json"

type articleOption struct {
	Body      string
	Author    string
	Reference string
}

func getArticleTemplate() (*template.Template, error) {
	file, err := filepath.Abs("./article_template.html")

	if err != nil {
		return nil, err
	}

	conts, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, err
	}

	temp, err := template.New("article").Parse(string(conts))

	if err != nil {
		return nil, err
	}

	return temp, nil
}

func renderArticle() (string, error) {
	t, err := getArticleTemplate()

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	options, err := getArticleOptions()

	if err != nil {
		return "", nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := r.Intn(len(options))

	err = t.Execute(&buf, options[random])

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getArticleOptions() ([]articleOption, error) {
	var options []articleOption

	res, err := http.Get(articleOptionsURL)
	defer res.Body.Close()

	if err != nil {
		return options, err
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("fetch article options, got http.Status: %s", res.Status)
		return options, err
	}

	if err = json.NewDecoder(res.Body).Decode(&options); err != nil {
		return options, err
	}

	return options, nil
}

func NewArticle(year int, progress float64) (*Article, error) {
	p := math.Floor(progress * 100)

	log.Debugf("create article with progress value [%v]", p)

	articleContent, err := renderArticle()

	if err != nil {
		return nil, err
	}

	article := Article{
		Title:   fmt.Sprintf("%v 年已经过去了 %v%s 啦", year, p, "%"),
		Content: articleContent,
	}

	log.WithFields(log.Fields{
		"title": article.Title,
	}).Debug("new article")

	return &article, nil
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
