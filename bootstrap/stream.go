package bootstrap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/webrtc-demo-go/config"
	"log"
)

func GetRTSP() (stream string, err error) {
	return getStream("rtsp")
}

func GetHLS() (stream string, err error) {
	return getStream("hls")
}

// getStream构建涂鸦开放平台请求，请求rtsp/hls实时流播放地址
func getStream(streamType string) (stream string, err error) {
	url := fmt.Sprintf("https://%s/v1.0/users/%s/devices/%s/stream/actions/allocate", config.App.OpenAPIURL, config.App.UID, config.App.DeviceID)

	request := struct {
		Type string `json:"type"` // stream的类型，rtsp/hls
	}{
		Type: streamType,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		log.Printf("marshal stream Request fail: %s", err.Error())

		return "", err
	}

	body, err := Rest("POST", url, bytes.NewReader(payload))
	if err != nil {
		log.Printf("GET stream fail: %s, body: %s", err.Error(), string((body)))

		return
	}

	keyValue := gjson.GetBytes(body, "result.url")
	if !keyValue.Exists() {
		log.Printf("stream url not exist, body: %s", string(body))

		return "", fmt.Errorf("stream url not exist")
	}

	stream = keyValue.String()

	return
}
