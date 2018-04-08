package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type accessTokenResponse struct {
	WechatGlobalError
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func fetchAccessToken(c *http.Client, appId, appSecret string) (result *accessTokenResponse, err error) {
	res, err := c.Get("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + appId + "&secret=" + appSecret)
	defer res.Body.Close()

	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", res.Status)
		return
	}

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return
	}

	if result.ErrCode != 0 {
		err = errors.New(fmt.Sprintf("[wechat global error]: errcode: %d, errmsg: %s", result.ErrCode, result.ErrMsg))
		return
	}

	return
}

func fetchAllTags(client *Client) (tags []Tag, err error) {
	apiURL := "https://api.weixin.qq.com/cgi-bin/tags/get"

	var result struct {
		WechatGlobalError
		Tags []Tag `json:"tags"`
	}

	if err = client.Get(apiURL, "", &result); err != nil {
		return
	}

	if result.ErrCode != 0 {
		err = &result.WechatGlobalError
		return
	}

	tags = result.Tags
	return
}

func fetchImageMaterial(client *Client) (materials []Material, err error) {
	apiURL := "https://api.weixin.qq.com/cgi-bin/material/batchget_material"

	var result struct {
		WechatGlobalError
		TotalCount int        `json:"total_count"`
		ItemCount  int        `json:"item_count"`
		Item       []Material `json:"item"`
	}

	var data = struct {
		Type   string `json:"type"`
		Offset int    `json:"offset"`
		Count  int    `json:"count"`
	}{
		Type:   "image",
		Offset: 0,
		Count:  20,
	}

	if err = client.Post(apiURL, &data, &result); err != nil {
		return
	}

	if result.ErrCode != 0 {
		err = &result.WechatGlobalError
		return
	}

	materials = result.Item
	return
}

func UploadArticleMaterial(client *Client, article *ArticleMaterial) (mediaId string, err error) {
	apiURL := "https://api.weixin.qq.com/cgi-bin/material/add_news"

	var result struct {
		WechatGlobalError
		Type      string `json:"type"`
		MediaId   string `json:"media_id"`
		CreatedAt string `json:"created_at"`
	}

	var data = struct {
		Articles []*ArticleMaterial `json:"articles"`
	}{}

	data.Articles = []*ArticleMaterial{article}

	if err = client.Post(apiURL, &data, &result); err != nil {
		return
	}

	if result.ErrCode != 0 {
		err = &result.WechatGlobalError
		return
	}

	mediaId = result.MediaId
	return
}

func BetchPostArticle(client *Client, mediaId string) (err error) {
	apiURL := "https://api.weixin.qq.com/cgi-bin/message/mass/sendall"

	var result struct {
		WechatGlobalError
		MsgId     int64 `json:"msg_id"`
		MsgDataId int64 `json:"msg_data_id"`
	}

	var data = struct {
		Filter struct {
			IsToAll bool `json:"is_to_all"`
			TagId   int  `json:"tag_id"`
		} `json:"filter"`
		Mpnews struct {
			MediaId string `json:"media_id"`
		} `json:"mpnews"`
		MsgType string `json:"msgtype"`
	}{}

	data.Filter.IsToAll = false
	data.Filter.TagId = 2
	data.Mpnews.MediaId = mediaId
	data.MsgType = "mpnews"

	if err = client.Post(apiURL, &data, &result); err != nil {
		return
	}

	if result.ErrCode != 0 {
		err = &result.WechatGlobalError
		return
	}

	return
}
