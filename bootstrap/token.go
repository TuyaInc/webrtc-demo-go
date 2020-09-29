package bootstrap

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/webrtc-demo-go/config"
	"log"
	"time"
)

// InitToken 根据授权码获取token
func InitToken() (err error) {
	var url string

	switch config.App.AuthMode {
	case "easy":
		url = fmt.Sprintf("https://%s/v1.0/token?grant_type=1", config.App.OpenAPIURL)
	case "auth":
		url = fmt.Sprintf("https://%s/v1.0/token?grant_type=2&code=%s", config.App.OpenAPIURL, config.App.Auth.Code)
	default:
		return fmt.Errorf("unsupported auth mode %s", config.App.AuthMode)
	}

	body, err := Rest("GET", url, nil)
	if err != nil {
		log.Printf("GET token fail: %s, body: %s", err.Error(), string((body)))

		return
	}

	err = syncToConfig(body)
	if err != nil {
		log.Printf("sync OpenAPI ressponse to config fail: %s", err.Error())

		return
	}

	// 启动token维护更新协程
	go maintainToken()

	return
}

// 后续使用refrensh_token刷新token，并得到新的refresh_token
func refreshToken() (err error) {
	url := fmt.Sprintf("https://%s/v1.0/token/%s", config.App.OpenAPIURL, config.App.RefreshToken)

	body, err := Rest("GET", url, nil)
	if err != nil {
		log.Printf("GET token fail: %s, body: %s", err.Error(), string((body)))

		return
	}

	err = syncToConfig(body)
	if err != nil {
		log.Printf("sync OpenAPI ressponse to config fail: %s", err.Error())

		return
	}

	return
}

// 同步OpenAPI token服务接口的Response到Sample全局App配置
func syncToConfig(body []byte) error {
	uIdValue := gjson.GetBytes(body, "result.uid")
	if !uIdValue.Exists() {
		log.Printf("uid not exits in body: %s", string(body))

		return errors.New("uid not exist")
	}

	accessTokenValue := gjson.GetBytes(body, "result.access_token")
	if !accessTokenValue.Exists() {
		log.Printf("access_token not exits in body: %s", string(body))

		return errors.New("access_token not exist")
	}

	refreshTokenValue := gjson.GetBytes(body, "result.refresh_token")
	if !refreshTokenValue.Exists() {
		log.Printf("refresh_token not exist")

		return errors.New("refresh_token not exist")
	}

	expireTimeValue := gjson.GetBytes(body, "result.expire_time")
	if !expireTimeValue.Exists() {
		log.Printf("expire_time not exist")

		return errors.New("expire_time not exist")
	}

	switch config.App.AuthMode {
	case "easy":
		config.App.UID = config.App.Easy.UID
	case "auth":
		config.App.UID = uIdValue.String()
	default:
		return fmt.Errorf("unsupported auth mode %s", config.App.AuthMode)
	}

	config.App.AccessToken = accessTokenValue.String()
	config.App.RefreshToken = refreshTokenValue.String()
	config.App.ExpireTime = expireTimeValue.Int()

	log.Printf("UID: %s", config.App.UID)
	log.Printf("AccessToken: %s", config.App.AccessToken)
	log.Printf("RefreshToken: %s", config.App.RefreshToken)
	log.Printf("ExpireTime: %d", config.App.ExpireTime)

	return nil
}

// 第一次获取token成功后，需要定期维护更新token
// 如果更新失败，则每60 Sec再次更新
// 如果更新成功，则在token失效前300 Sec更新
func maintainToken() {
	interval := config.App.ExpireTime - 300

	for {
		timer := time.NewTimer(time.Duration(interval) * time.Second)

		select {
		case <-timer.C:
			if err := refreshToken(); err != nil {
				log.Printf("refresh token fail: %s", err.Error())

				interval = 60
			} else {
				interval = config.App.ExpireTime - 300
			}
		}
	}
}
