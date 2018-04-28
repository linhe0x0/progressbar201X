package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
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
	"github.com/clbanning/mxj"
)

// Sign 将 token、timestamp、nonce, msg_encrypt 四个参数进行字典序排序，
// 之后拼接成一个字符串进行 sha1 加密，作为签名的结果。
func Sign(token, timestamp, nonce, msg_encrypt string) string {
	strs := sort.StringSlice{token, timestamp, nonce, msg_encrypt}

	strs.Sort()

	buf := make([]byte, 0, len(token)+len(timestamp)+len(nonce)+len(msg_encrypt))

	buf = append(buf, strs[0]...)
	buf = append(buf, strs[1]...)
	buf = append(buf, strs[2]...)
	buf = append(buf, strs[3]...)

	hashsum := sha1.Sum(buf)

	return hex.EncodeToString(hashsum[:])
}

func ParseXML(xml []byte) (map[string]interface{}, error) {
	m, err := mxj.NewMapXml(xml)

	if err != nil {
		return nil, err
	}

	if _, ok := m["xml"]; !ok {
		err := errors.New("Invalid message")

		return nil, err
	}

	message, ok := m["xml"].(map[string]interface{})

	if !ok {
		err := errors.New("Invalid field `xml` Type")

		return nil, err
	}

	return message, nil
}

func DecryptMsg(appID, encryptedMsg, aesKey string) (random, rawMsgXMLBytes []byte, err error) {
	var encryptedMsgBytes, key, getAppIDBytes []byte

	encryptedMsgBytes, err = base64.StdEncoding.DecodeString(encryptedMsg)

	if err != nil {
		return
	}

	key, err = aesKeyDecode(aesKey)

	if err != nil {
		return
	}

	random, rawMsgXMLBytes, getAppIDBytes, err = AESDecryptMsg(encryptedMsgBytes, key)

	if err != nil {
		err = fmt.Errorf("the message decryption failed, error message: %v", err)
		return
	}

	if appID != string(getAppIDBytes) {
		err = errors.New("invalid app id")
		return
	}

	return
}

func aesKeyDecode(encodedAESKey string) (key []byte, err error) {
	if len(encodedAESKey) != 43 {
		err = fmt.Errorf("the length of encodedAESKey must be equal to 43")
		return
	}

	key, err = base64.StdEncoding.DecodeString(encodedAESKey + "=")

	if err != nil {
		return
	}

	if len(key) != 32 {
		err = fmt.Errorf("encodingAESKey invalid")
		return
	}

	return
}

// ciphertext = AES_Encrypt[random(16B) + msg_len(4B) + rawXMLMsg + appId]
func AESDecryptMsg(ciphertext []byte, aesKey []byte) (random, rawXMLMsg, appId []byte, err error) {
	const (
		BLOCK_SIZE = 32             // PKCS#7
		BLOCK_MASK = BLOCK_SIZE - 1 // BLOCK_SIZE 为 2^n 时, 可以用 mask 获取针对 BLOCK_SIZE 的余数
	)

	if len(ciphertext) < BLOCK_SIZE {
		err = fmt.Errorf("the length of ciphertext too short: %d", len(ciphertext))
		return
	}

	if len(ciphertext)&BLOCK_MASK != 0 {
		err = fmt.Errorf("ciphertext is not a multiple of the block size, the length is %d", len(ciphertext))
		return
	}

	plaintext := make([]byte, len(ciphertext)) // len(plaintext) >= BLOCK_SIZE

	// 解密
	block, err := aes.NewCipher(aesKey)

	if err != nil {
		return
	}

	mode := cipher.NewCBCDecrypter(block, aesKey[:16])
	mode.CryptBlocks(plaintext, ciphertext)

	// PKCS#7 去除补位
	amountToPad := int(plaintext[len(plaintext)-1])

	if amountToPad < 1 || amountToPad > BLOCK_SIZE {
		err = fmt.Errorf("the amount to pad is incorrect: %d", amountToPad)
		return
	}

	plaintext = plaintext[:len(plaintext)-amountToPad]

	// len(plaintext) == 16+4+len(rawXMLMsg)+len(appId)
	if len(plaintext) <= 20 {
		err = fmt.Errorf("plaintext too short, the length is %d", len(plaintext))
		return
	}

	rawXMLMsgLen := int(decodeNetworkByteOrder(plaintext[16:20]))

	if rawXMLMsgLen < 0 {
		err = fmt.Errorf("incorrect msg length: %d", rawXMLMsgLen)
		return
	}

	appIdOffset := 20 + rawXMLMsgLen

	if len(plaintext) <= appIdOffset {
		err = fmt.Errorf("msg length too large: %d", rawXMLMsgLen)
		return
	}

	random = plaintext[:16:20]
	rawXMLMsg = plaintext[20:appIdOffset:appIdOffset]
	appId = plaintext[appIdOffset:]

	return
}

func EncryptMsg(random, rawXMLMsg []byte, appID, aesKey string) (encrtptMsg []byte, err error) {
	var key []byte

	key, err = aesKeyDecode(aesKey)

	if err != nil {
		return
	}

	ciphertext := AESEncryptMsg(random, rawXMLMsg, appID, key)
	encrtptMsg = []byte(base64.StdEncoding.EncodeToString(ciphertext))
	return
}

// ciphertext = AES_Encrypt[random(16B) + msg_len(4B) + rawXMLMsg + appId]
func AESEncryptMsg(random, rawXMLMsg []byte, appId string, aesKey []byte) (ciphertext []byte) {
	const (
		BLOCK_SIZE = 32             // PKCS#7
		BLOCK_MASK = BLOCK_SIZE - 1 // BLOCK_SIZE 为 2^n 时, 可以用 mask 获取针对 BLOCK_SIZE 的余数
	)

	appIdOffset := 20 + len(rawXMLMsg)
	contentLen := appIdOffset + len(appId)
	amountToPad := BLOCK_SIZE - contentLen&BLOCK_MASK
	plaintextLen := contentLen + amountToPad

	plaintext := make([]byte, plaintextLen)

	// 拼接
	copy(plaintext[:16], random)
	encodeNetworkByteOrder(plaintext[16:20], uint32(len(rawXMLMsg)))
	copy(plaintext[20:], rawXMLMsg)
	copy(plaintext[appIdOffset:], appId)

	// PKCS#7 补位
	for i := contentLen; i < plaintextLen; i++ {
		plaintext[i] = byte(amountToPad)
	}

	// 加密
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		panic(err)
	}
	mode := cipher.NewCBCEncrypter(block, aesKey[:16])
	mode.CryptBlocks(plaintext, plaintext)

	ciphertext = plaintext
	return
}

// 从 4 字节的网络字节序里解析出整数
func decodeNetworkByteOrder(b []byte) (n uint32) {
	return uint32(b[0])<<24 |
		uint32(b[1])<<16 |
		uint32(b[2])<<8 |
		uint32(b[3])
}

// 把整数 n 格式化成 4 字节的网络字节序
func encodeNetworkByteOrder(b []byte, n uint32) {
	b[0] = byte(n >> 24)
	b[1] = byte(n >> 16)
	b[2] = byte(n >> 8)
	b[3] = byte(n)
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

	log.Debugf("selected %vth from %v materials", random+1, len)

	return
}

// func GetAllTags(client *Client) (tags []Tag, err error) {
// 	tags, err = fetchAllTags(client)

// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	return
// }

// func UploadArticleMaterial(client *Client, article *ArticleMaterial) (mediaId string, err error) {
// 	mediaId, err = uploadArticleMaterial(client, article)

// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	return
// }

// func BetchPostArticle(client *Client, mediaId string) (err error) {
// 	err = betchPostArticle(client, mediaId)

// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	return
// }
