package wechat

import (
	"net/http"
	"net/url"
	"time"

	"github.com/apex/log"
)

type accessToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
}

type refreshTokenResult struct {
	token string
	err   error
}

type DefaultAccessTokenServer struct {
	appId                    string
	appSecret                string
	httpClient               *http.Client
	refreshTokenRequestChan  chan bool
	refreshTokenResponseChan chan refreshTokenResult
	tokenCache               string
}

func NewDefautlAccessToeknServer(appId, appSecret string) *DefaultAccessTokenServer {
	if appId == "" || appSecret == "" {
		panic("appId or appSecret is invalid.")
	}

	server := &DefaultAccessTokenServer{
		appId:                    url.QueryEscape(appId),
		appSecret:                url.QueryEscape(appSecret),
		httpClient:               http.DefaultClient,
		refreshTokenRequestChan:  make(chan bool),
		refreshTokenResponseChan: make(chan refreshTokenResult),
	}

	go server.tokenUpdateDaemon(2 * time.Hour)

	return server
}

func (s DefaultAccessTokenServer) Token() (string, error) {
	if s.tokenCache != "" {
		log.Debug("load token from cache")

		return s.tokenCache, nil
	}

	log.Debug("load token from wechat server")
	return s.RefreshToken()
}

func (s *DefaultAccessTokenServer) RefreshToken() (string, error) {
	s.refreshTokenRequestChan <- true

	result := <-s.refreshTokenResponseChan

	return result.token, result.err
}

func (s *DefaultAccessTokenServer) tokenUpdateDaemon(initTickDuration time.Duration) {
	tickDuration := initTickDuration

NEW_TICK_DURATION:
	ticker := time.NewTicker(tickDuration)

	for {
		select {
		case <-s.refreshTokenRequestChan:
			accessToken, err := s.updateToken()

			if err != nil {
				s.refreshTokenResponseChan <- refreshTokenResult{err: err}
				break
			}

			s.refreshTokenResponseChan <- refreshTokenResult{token: accessToken.Token}

			tickDuration = time.Duration(accessToken.ExpiresIn) * time.Second
			ticker.Stop()

			goto NEW_TICK_DURATION
		case <-ticker.C:
			accessToken, err := s.updateToken()

			if err != nil {
				break
			}

			newTimeDuration := time.Duration(accessToken.ExpiresIn) * time.Second

			if abs(tickDuration-newTimeDuration) > 5*time.Second {
				tickDuration = newTimeDuration
				ticker.Stop()

				goto NEW_TICK_DURATION
			}
		}
	}
}

func abs(x time.Duration) time.Duration {
	if x >= 0 {
		return x
	}

	return -x
}

func (s *DefaultAccessTokenServer) updateToken() (accessToken, error) {
	response, err := fetchAccessToken(s.httpClient, s.appId, s.appSecret)

	if err != nil {
		log.WithError(err).Error("load token")
		return accessToken{}, err
	}

	log.Infof("fetched new token: %s", response.AccessToken)

	s.tokenCache = response.AccessToken

	return accessToken{response.AccessToken, response.ExpiresIn}, nil
}
