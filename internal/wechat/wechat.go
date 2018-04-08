package wechat

import (
	"fmt"
	"net/http"
)

type AccessTokenServer interface {
	Token() (token string, err error)
	RefreshToken() (token string, err error)
}

type Client struct {
	AccessTokenServer
	HttpClient *http.Client
}

type WechatGlobalError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (err *WechatGlobalError) Error() string {
	return fmt.Sprintf("[wechat global error]: errcode: %d, errmsg: %s", err.ErrCode, err.ErrMsg)
}

type Tag struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Material struct {
	MediaId    string `json:"media_id"`
	Name       string `json:"name"`
	UpdateTime int64  `json:"update_time"`
	Url        string `json:"url"`
}

type ArticleMaterial struct {
	ThumbMediaId string `json:"thumb_media_id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Digest       string `json:"digest"`
	ShowCoverPic int    `json:"show_cover_pic"`
}
