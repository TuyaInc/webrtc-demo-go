package config

import (
	"encoding/json"
	"io/ioutil"
)

// APPConfig Sample运行的应用配置
type APPConfig struct {
	OpenAPIMode string `json:"openAPIMode"` // 连接涂鸦开放平台的模式，暂时只支持mqtt
	OpenAPIURL  string `json:"openAPIUrl"`  // 涂鸦开放平台URL

	ClientID string `json:"clientId"` // 涂鸦开放平台申请的appId
	Secret   string `json:"secret"`   // 涂鸦开放平台申请的secretKey

	Code string `json:"code"` // 涂鸦开放平台返回的授权码

	DeviceID string `json:"deviceId"` // Sample要连接的设备id

	UID          string `json:"-"` // 涂鸦开放平台授权码模式返回的用户uid
	MQTTUID      string `json:"-"` // 与涂鸦MQTT通信时Web端的Topic标识ID
	AccessToken  string `json:"-"` // 涂鸦开放平台授权码模式返回的access_token
	RefreshToken string `json:"-"` // 涂鸦开放平台授权码模式返回的refresh_token
	ExpireTime   int64  `json:"-"` // 涂鸦开放平台授权码模式返回的token有效期
}

var App = APPConfig{
	OpenAPIMode: "mqtt",
	OpenAPIURL:  "openapi.tuyacn.com",

	ClientID: "",
	Secret:   "",

	Code: "",

	DeviceID: "",
}

// LoadConfig 加载webrtc.json配置到运行时环境
func LoadConfig() error {
	return parseJSON("webrtc.json", &App)
}

func parseJSON(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, v)
	return err
}
