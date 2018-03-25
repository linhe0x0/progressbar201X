package wechat

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/apex/log"
)

// Sign 将 token、timestamp、nonce 三个参数进行字典序排序，
// 之后拼接成一个字符串进行 sha1 加密，作为签名的结果。
func Sign(token, timestamp, nonce string) string {
	strs := sort.StringSlice{token, timestamp, nonce}

	strs.Sort()

	buf := make([]byte, 0, len(token)+len(timestamp)+len(nonce))

	buf = append(buf, strs[0]...)
	buf = append(buf, strs[1]...)
	buf = append(buf, strs[2]...)

	hashsum := sha1.Sum(buf)

	return hex.EncodeToString(hashsum[:])
}

func NewClient(ats AccessTokenServer) *Client {
	c := &Client{
		AccessTokenServer: ats,
		HttpClient:        http.DefaultClient,
	}

	return c
}

func (client *Client) Get(apiURL string, querystring string, response interface{}) (err error) {
	token, err := client.Token()

	if err != nil {
		return
	}

	uri := apiURL + fmt.Sprintf("?access_token=%s", token)

	if querystring != "" {
		uri += "&" + querystring
	}

	logRequest("GET", uri, []byte{})

	res, err := client.HttpClient.Get(uri)
	defer res.Body.Close()

	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", res.Status)
		return
	}

	responseBody, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	logResponse(res, responseBody)

	if err = json.Unmarshal(responseBody, response); err != nil {
		return
	}

	return nil
}

func (client *Client) Post(apiURL string, data interface{}, response interface{}) (err error) {
	token, err := client.Token()

	if err != nil {
		return err
	}

	uri := apiURL + fmt.Sprintf("?access_token=%s", token)

	initialBuffer := []byte{}
	buf := bytes.NewBuffer(initialBuffer)

	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)

	if err = encoder.Encode(data); err != nil {
		return
	}

	requestBodyBytes := buf.Bytes()
	requestBody := bytes.NewReader(requestBodyBytes)

	logRequest("POST", uri, requestBodyBytes)

	res, err := client.HttpClient.Post(uri, "application/json; charset=utf-8", requestBody)
	defer res.Body.Close()

	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", res.Status)
		return
	}

	responseBody, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	logResponse(res, responseBody)

	if err = json.Unmarshal(responseBody, response); err != nil {
		return
	}

	return nil
}

// logRequest logs the request
func logRequest(method, uri string, body []byte) {
	if i := len(body) - 1; i >= 0 && body[i] == '\n' {
		body = body[:i] // Remove \n at the end of the line
	}

	log.WithFields(log.Fields{
		"method": method,
		"uri":    uri,
		"body":   string(body),
	}).Debug("<= request")
}

// logResponse logs the response.
func logResponse(res *http.Response, bodyBytes []byte) {
	logger := log.WithFields(log.Fields{
		"status": res.StatusCode,
		"body":   string(bodyBytes),
	})

	switch {
	case res.StatusCode >= 500:
		logger.Error("=> response")
	case res.StatusCode >= 400:
		logger.Warn("=> response")
	default:
		logger.Debug("=> response")
	}
}

func GetRandomImageMaterial(client *Client) (randomMaterial Material, err error) {
	materials, err := fetchImageMaterial(client)

	if err != nil {
		return
	}

	len := len(materials)

	if len == 0 {
		err = errors.New("No image materials is available.")
		return
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := r.Intn(len)

	randomMaterial = materials[random]

	log.Debugf("selected %vth from %v materials", random + 1, len)

	return
}

func GetAllTags(client *Client) (tags []Tag, err error) {
	tags, err = fetchAllTags(client)

	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func UploadArticleMaterial(client *Client, article *ArticleMaterial) (mediaId string, err error) {
	mediaId, err = uploadArticleMaterial(client, article)

	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func BetchPostArticle(client *Client, mediaId string) (err error) {
	err = betchPostArticle(client, mediaId)

	if err != nil {
		fmt.Println(err)
		return
	}

	return
}
