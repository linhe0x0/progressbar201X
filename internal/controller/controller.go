package controller

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/sqrthree/progressbar201X/internal/config"
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

	expectedSignature := wechat.Sign(Config.Wechat.Token, timestamp, nonce)

	if signature != expectedSignature {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, echostr)
}
