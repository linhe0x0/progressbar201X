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
	Digest  string
	Content string
}

type ReferenceOption struct {
	Body      string
	Author    string
	Reference string
}

type CustomizedOptions struct {
	Digests    []string
	References []ReferenceOption
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
func renderArticle(contentTitle string, reference ReferenceOption) (string, error) {
	t, err := getArticleTemplate()

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	type ContentOption struct {
		Title     string
		Content   ReferenceOption
	}

	err = t.Execute(&buf, ContentOption{contentTitle, reference})

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// getCustomizedOptions fetches latest options from `articleOptionsURL`
func getCustomizedOptions() (CustomizedOptions, error) {
	var options CustomizedOptions

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

// ▓▓▓▓░░░░░░
func GenerateBar(p float64) string {
	var str string

	completed := int(p / 10)

	for i := 0; i < 10; i++ {
		if i < completed {
			str += "\u2593"
		} else {
			str += "\u2591"
		}
	}

	return str
}

// New creates a new article with specified value
func New(year int, p float64) (*Article, error) {
	pageTitle := GenerateBar(p)
	contentTitle := fmt.Sprintf("%v 年已经过去了 %v%s 啦", year, p, "%")

	options, err := getCustomizedOptions()

	if err != nil {
		log.WithError(err).Error("fetch customized options")

		return nil, err
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	referenceIndex := r.Intn(len(options.References))
	digestIndex := r.Intn(len(options.Digests))

	articleContent, err := renderArticle(contentTitle, options.References[referenceIndex])

	if err != nil {
		log.WithError(err).Error("render article")

		return nil, err
	}

	article := Article{
		Title:   pageTitle,
		Digest:  options.Digests[digestIndex],
		Content: articleContent,
	}

	log.WithFields(log.Fields{
		"title":  article.Title,
		"digest": article.Digest,
	}).Debug("new article")

	return &article, nil
}
