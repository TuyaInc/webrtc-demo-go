package bootstrap

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/webrtc-demo-go/config"
	"log"
)

// GetCloudStorageTimeline 请求指定时间范围内所有录像片段的起始时间信息列表
func GetCloudStorageTimeline(timeGT, timeLT int64) (timeline string, err error) {
	url := fmt.Sprintf("https://%s/v1.0/users/%s/devices/%s/storage/stream/timeline?timeGT=%d&timeLT=%d",
		config.App.OpenAPIURL, config.App.UID, config.App.DeviceID, timeGT, timeLT)

	body, err := Rest("GET", url, nil)
	if err != nil {
		log.Printf("GET cloud storage timeline fail: %s, body: %s", err.Error(), string((body)))

		return
	}

	keyValue := gjson.GetBytes(body, "result")
	if !keyValue.Exists() {
		log.Printf("cloud storage timeline not exist, body: %s", string(body))

		return "", errors.New("timeline not exist")
	}

	timeline = keyValue.String()

	return
}

// GetCloudStorageHls 请求指定时间范围内所有云存储录像的播放资源
func GetCloudStorageHls(timeGT, timeLT int64) (hls string, err error) {
	url := fmt.Sprintf("https://%s/v1.0/users/%s/devices/%s/storage/stream/hls?timeGT=%d&timeLT=%d&callback=%s",
		config.App.OpenAPIURL, config.App.UID, config.App.DeviceID, timeGT, timeLT, "http://localhost:3333")

	body, err := Rest("GET", url, nil)
	if err != nil {
		log.Printf("GET cloud storage hls fail: %s, body: %s", err.Error(), string((body)))

		return
	}

	keyValue := gjson.GetBytes(body, "result.list")
	if !keyValue.Exists() {
		log.Printf("cloud storage hls not exist, body: %s", string(body))

		return "", errors.New("hls not exist")
	}

	hls = keyValue.String()

	return
}

// GetCloudStorageKey 根据用户id和设备id从开放平台获取信令服务moto的id、webRTC认证需要的授权码
func GetCloudStorageKey(magic string) (key string, err error) {
	url := fmt.Sprintf("https://%s/v1.0/users/%s/devices/%s/storage/stream/key?magic=%s",
		config.App.OpenAPIURL, config.App.UID, config.App.DeviceID, magic)

	body, err := Rest("GET", url, nil)
	if err != nil {
		log.Printf("GET cloud storage key fail: %s, body: %s", err.Error(), string((body)))

		return
	}

	keyValue := gjson.GetBytes(body, "result.key")
	if !keyValue.Exists() {
		log.Printf("key not exist in cloud storage key, body: %s", string(body))

		return "", errors.New("key not exist")
	}

	key = keyValue.String()

	return
}
