// Package article implements some useful methods about Aritcle.
package article

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"regexp"
	"text/template"
	"time"

	"github.com/apex/log"
)

const articleOptionsURL = "https://raw.githubusercontent.com/sqrthree/progressbar201X/quotations/main.json"

type Article struct {
	Title   string
	Content string
}

type articleOption struct {
	Body      string
	Author    string
	Reference string
}

// getArticleTemplate reads the content of `PROJECT_ROOT/article_template.html`
// and use it as the template of article.
func getArticleTemplate() (*template.Template, error) {
	file, err := filepath.Abs("./article_template.html")

	log.Debug("read article template from " + file)

	if err != nil {
		return nil, err
	}

	conts, err := ioutil.ReadFile(file)

	if err != nil {
		return nil, err
	}

	temp, err := template.New("article").Parse(compressHTMLString(string(conts)))

	if err != nil {
		return nil, err
	}

	return temp, nil
}

// compressHTMLString compresses HTML code and retrun the compressed string.
func compressHTMLString(s string) string {
	return regexp.MustCompile("\\s*(<[^><]*>)\\s*").ReplaceAllString(s, "$1")
}

// renderArticle renders a article template with random options from `articleOptionsURL`.
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

// getArticleOptions fetches latest options from `articleOptionsURL`
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

// New creates a new article with specified title
func New(title string) (*Article, error) {
	articleContent, err := renderArticle()

	if err != nil {
		return nil, err
	}

	article := Article{
		Title:   title,
		Content: articleContent,
	}

	log.WithFields(log.Fields{
		"title": article.Title,
	}).Debug("new article")

	return &article, nil
}
