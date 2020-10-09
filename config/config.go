package config

import (
	"encoding/json"
	"io/ioutil"
)

// Easy 简单模式，需要手动填入想要访问的IPC所属的uId
type Easy struct {
	UID string `json:"uId"`
}

// Auth 授权码模式，需要在涂鸦开放平台授权页面上输入用户的账号密码进行授权，并在此填入返回的授权码
type Auth struct {
	Code string `json:"code"`
}

// APPConfig Sample运行的应用配置
type APPConfig struct {
	OpenAPIMode string `json:"openAPIMode"` // 连接涂鸦开放平台的模式，暂时只支持mqtt
	OpenAPIURL  string `json:"openAPIUrl"`  // 涂鸦开放平台URL

	ClientID string `json:"clientId"` // 涂鸦开放平台申请的appId
	Secret   string `json:"secret"`   // 涂鸦开放平台申请的secretKey

	AuthMode string `json:"authMode"` // 授权模式，为"easy" / "auth"

	Easy Easy `json:"easy"`
	Auth Auth `json:"auth"`

	DeviceID string `json:"deviceId"` // Sample要连接的设备id

	UID          string `json:"-"` // 设备ID所属的用户ID，easy模式下人工填写，auth模式通过授权码获取access_token会返回用户ID
	MQTTUID      string `json:"-"` // 与涂鸦MQTT通信时Web端的Topic标识ID
	AccessToken  string `json:"-"` // 涂鸦开放平台授权码模式返回的access_token
	RefreshToken string `json:"-"` // 涂鸦开放平台授权码模式返回的refresh_token
	ExpireTime   int64  `json:"-"` // 涂鸦开放平台授权码模式返回的token有效期
}

var App = APPConfig{
	OpenAPIMode: "mqtt",
	OpenAPIURL:  "openapi.tuyacn.com",
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
