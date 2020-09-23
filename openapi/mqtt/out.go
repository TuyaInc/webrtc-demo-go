package openmqtt

import (
	"encoding/json"
	"github.com/webrtc-demo-go/config"
	"github.com/webrtc-demo-go/types"
	"log"
	"time"
)

func Post(msg *types.WsMessage) {
	switch msg.Type {
	case "offer":
		sendOffer(msg, msg.Payload)
	case "candidate":
		sendCandidate(msg, msg.Payload)
	case "disconnect":
		sendDisconnect(msg)
	default:
		log.Printf("unsupported ws message, type: %s", msg.Type)
	}
}

func sendOffer(msg *types.WsMessage, sdp string) {
	offerFrame := struct {
		Mode       string `json:"mode"`        // offer的模式，默认为webrtc
		Sdp        string `json:"sdp"`         // 浏览器生成的offer
		StreamType uint32 `json:"stream_type"` // 码流类型，默认为1
		Auth       string `json:"auth"`        // webRTC认证需要的授权码，从开放平台获取
	}{
		Mode:       "webrtc",
		Sdp:        sdp,
		StreamType: 1,
		Auth:       auth,
	}

	offerMqtt := &MqttMessage{
		Protocol: 302,
		Pv:       "2.2",
		T:        time.Now().Unix(),
		Data: MqttFrame{
			Header: MqttFrameHeader{
				Type:      "offer",
				From:      config.App.MQTTUID,
				To:        config.App.DeviceID,
				SubDevID:  "",
				SessionID: msg.SessionID,
				MotoID:    motoID,
			},
			Message: offerFrame,
		},
	}

	sendBytes, err := json.Marshal(offerMqtt)
	if err != nil {
		log.Printf("marshal offer mqtt to bytes fail: %s", err.Error())

		return
	}

	publish(sendBytes)
}

func sendCandidate(msg *types.WsMessage, candidate string) {
	candidateFrame := struct {
		Mode      string `json:"mode"`      // candidate的模式，默认为webrtc
		Candidate string `json:"candidate"` // 候选地址，a=candidate:1922393870 1 UDP 2130706431 192.168.1.171 51532 typ host
	}{
		Mode:      "webrtc",
		Candidate: candidate,
	}

	candidateMqtt := &MqttMessage{
		Protocol: 302,
		Pv:       "2.2",
		T:        time.Now().Unix(),
		Data: MqttFrame{
			Header: MqttFrameHeader{
				Type:      "candidate",
				From:      config.App.MQTTUID,
				To:        config.App.DeviceID,
				SubDevID:  "",
				SessionID: msg.SessionID,
				MotoID:    motoID,
			},
			Message: candidateFrame,
		},
	}

	sendBytes, err := json.Marshal(candidateMqtt)
	if err != nil {
		log.Printf("marshal candidate mqtt to bytes fail: %s", err.Error())

		return
	}

	publish(sendBytes)
}

func sendDisconnect(msg *types.WsMessage) {
	disconnectFrame := struct {
		Mode string `json:"mode"`
	}{
		Mode: "webrtc", // disconnect的模式，默认为webrtc
	}

	disconnectMqtt := &MqttMessage{
		Protocol: 302,
		Pv:       "2.2",
		T:        time.Now().Unix(),
		Data: MqttFrame{
			Header: MqttFrameHeader{
				Type:      "disconnect",
				From:      config.App.MQTTUID,
				To:        config.App.DeviceID,
				SubDevID:  "",
				SessionID: msg.SessionID,
				MotoID:    motoID,
			},
			Message: disconnectFrame,
		},
	}

	sendBytes, err := json.Marshal(disconnectMqtt)
	if err != nil {
		log.Printf("marshal candidate mqtt to bytes fail: %s", err.Error())

		return
	}

	publish(sendBytes)
}

// 发布mqtt消息
func publish(payload []byte) {
	token := client.Publish(publishTopic, 1, false, payload)
	if token.Error() != nil {
		log.Printf("mqtt publish fail: %s, topic: %s", token.Error().Error(),
			publishTopic)
	}
}
