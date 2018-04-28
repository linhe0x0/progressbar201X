package controller

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"

	. "github.com/sqrthree/progressbar201X/internal/config"
	"github.com/sqrthree/progressbar201X/internal/timeline"
	"github.com/sqrthree/progressbar201X/internal/wechat"
)

func Pong(w http.ResponseWriter, r *http.Request) {
	echostr := r.URL.Query().Get("echostr")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	signature := r.URL.Query().Get("signature")

	if strings.TrimSpace(echostr) == "" {
		http.Error(w, "parameter `echostr` is invalid.", http.StatusBadRequest)
		return
	}

	expectedSignature := wechat.Sign(Config.Wechat.Token, timestamp, nonce, "")

	if signature != expectedSignature {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, echostr)
}

func HandleEvents(w http.ResponseWriter, r *http.Request) {
	encryptType := r.URL.Query().Get("encrypt_type")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	signature := r.URL.Query().Get("signature")

	if encryptType != "aes" {
		http.Error(w, "Unsupported encryption type", http.StatusBadRequest)
		return
	}

	expectedSignature := wechat.Sign(Config.Wechat.Token, timestamp, nonce, "")

	if signature != expectedSignature {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	message, err := wechat.ParseXML(body)

	if err != nil {
		log.WithError(err).Error("parse cryptographic xml")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	encrypt, ok := message["Encrypt"].(string)

	if !ok {
		log.Error("stringify encrypt")
		http.Error(w, "Invalid value of encrypt", http.StatusBadRequest)
		return
	}

	_, rawXMLMsg, err := wechat.DecryptMsg(Config.Wechat.AppId, encrypt, Config.Wechat.AESKey)

	if err != nil {
		log.WithError(err).Error("decrypt message")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.WithField("rawXMLMsg", string(rawXMLMsg)).Debug("rawXMLMsg")

	data, err := wechat.ParseXML(rawXMLMsg)

	if err != nil {
		log.WithError(err).Error("parse unencrypted xml")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var contentOfResponse string

	switch data["EventKey"] {
	case "month":
		contentOfResponse, err = responseOfEventMonth()
	default:
		log.Error("Unsupported event type")
		http.Error(w, "Unsupported event type", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.WithError(err).Error("get progress of current month")
		http.Error(w, "Unsupported event type", http.StatusBadRequest)
		return
	}

	random := RandomStr(16)
	timestampOfTheMoment := strconv.Itoa(int(time.Now().Unix()))
	rawXMLResponse := []byte(fmt.Sprintf("<xml><ToUserName>%s</ToUserName><FromUserName>%s</FromUserName><CreateTime>%s</CreateTime><MsgType>text</MsgType><Content>%s</Content></xml>", value2CDATA(data["FromUserName"]), value2CDATA(data["ToUserName"]), value2CDATA(timestampOfTheMoment), value2CDATA(contentOfResponse)))

	log.WithField("rawXMLResponse", string(rawXMLResponse)).Debug("raw response XML")

	ciphertext, err := wechat.EncryptMsg(random, rawXMLResponse, Config.Wechat.AppId, Config.Wechat.AESKey)

	if err != nil {
		log.WithError(err).Error("encrypt xml")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msgSignature := wechat.Sign(Config.Wechat.Token, timestamp, nonce, string(ciphertext))

	XMLResponse := fmt.Sprintf("<xml><Encrypt>%s</Encrypt><MsgSignature>%s</MsgSignature><TimeStamp>%s</TimeStamp><Nonce>%s</Nonce></xml>", value2CDATA(string(ciphertext)), value2CDATA(msgSignature), value2CDATA(timestamp), value2CDATA(nonce))

	fmt.Fprintln(w, XMLResponse)
}

func value2CDATA(v interface{}) string {
	return fmt.Sprintf("<![CDATA[%s]]>", v.(string))
}

func RandomStr(length int) []byte {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)

	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}

	return result
}

func responseOfEventMonth() (string, error) {
	progress, err := getProgressOfCurrentMonth()

	if err != nil {
		return "", nil
	}

	p := math.Floor(progress * 100)

	return fmt.Sprintf("本月已经过去了 %v%s。", p, "%"), nil
}

func getProgressOfCurrentMonth() (progress float64, err error) {
	now := time.Now().UTC().Add(8 * time.Hour)

	progress, err = timeline.NewWithMonth(now)
	return
}
