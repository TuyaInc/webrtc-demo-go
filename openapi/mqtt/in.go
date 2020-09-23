package openmqtt

import (
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	"github.com/webrtc-demo-go/bootstrap"
	"github.com/webrtc-demo-go/types"
	"log"
	"strings"
)

// mqtt消息接受回调函数
func consume(client mqtt.Client, msg mqtt.Message) {
	// 接受的mqtt消息体可为多种类型，先用反序列化到临时对象，json.RawMessage([]byte)存储消息体，在对应的handler反序列化到对应对象
	tmp := struct {
		Protocol int    `json:"protocol"`
		Pv       string `json:"pv"`
		T        int64  `json:"t"`
		Data     struct {
			Header  MqttFrameHeader `json:"header"`
			Message json.RawMessage `json:"msg"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(msg.Payload(), &tmp); err != nil {
		log.Printf("unmarshal received mqtt fail: %s, payload: %s", err.Error(), string(msg.Payload()))

		return
	}

	rmqtt := &MqttMessage{
		Protocol: tmp.Protocol,
		Pv:       tmp.Pv,
		T:        tmp.T,
		Data: MqttFrame{
			Header:  tmp.Data.Header,
			Message: tmp.Data.Message,
		},
	}

	log.Printf("mqtt recv message, session: %s, type: %s, from: %s, to: %s",
		rmqtt.Data.Header.SessionID,
		rmqtt.Data.Header.Type,
		rmqtt.Data.Header.From,
		rmqtt.Data.Header.To)

	dispatch(rmqtt)
}

// 分发从mqtt服务器接受到的消息
func dispatch(msg *MqttMessage) {
	link, err := bootstrap.GetLink(msg.Data.Header.SessionID)
	if err != nil {
		log.Printf("no link: %s, session: %s", err.Error(), msg.Data.Header.SessionID)

		return
	}

	switch msg.Data.Header.Type {
	case "answer":
		handleAnswer(msg, link)
	case "candidate":
		handleCandidate(msg, link)
	case "disconnect":
		handleDisconnect(msg, link)
	}
}

func handleAnswer(msg *MqttMessage, link *types.WsMessage) {
	frame, ok := msg.Data.Message.(json.RawMessage)
	if !ok {
		log.Printf("convert interface{} to []byte fail, session: %s", msg.Data.Header.SessionID)

		return
	}

	answerFrame := struct {
		Mode string `json:"mode"`
		Sdp  string `json:"sdp"`
	}{}

	if err := json.Unmarshal(frame, &answerFrame); err != nil {
		log.Printf("unmarshal mqtt answer frame fail: %s, session: %s, frame: %s",
			msg.Data.Header.SessionID,
			string(msg.Data.Message.([]byte)))

		return
	}

	response := &types.WsMessage{
		AgentID:   link.AgentID,
		Type:      "answer",
		SessionID: msg.Data.Header.SessionID,
		Payload:   answerFrame.Sdp,
		Success:   true,
	}

	sendBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("marshal http response fail: %s", err.Error())

		return
	}

	log.Printf("ws write: %s", string(sendBytes))

	// body send back to Javascript
	err = link.Conn.WriteMessage(websocket.TextMessage, sendBytes)
	if err != nil {
		log.Printf("ws write fail: %s", err.Error())
	}
}

func handleCandidate(msg *MqttMessage, link *types.WsMessage) {
	frame, ok := msg.Data.Message.(json.RawMessage)
	if !ok {
		log.Printf("convert interface{} to []byte fail, session: %s", msg.Data.Header.SessionID)

		return
	}

	candidateFrame := struct {
		Mode      string `json:"mode"`
		Candidate string `json:"candidate"`
	}{}

	if err := json.Unmarshal(frame, &candidateFrame); err != nil {
		log.Printf("unmarshal mqtt candidate frame fail: %s, session: %s, frame: %s",
			msg.Data.Header.SessionID,
			string(msg.Data.Message.([]byte)))

		return
	}

	// candidate from device start with "a=", end with "\r\n", which are not needed by Chrome webRTC
	candidateFrame.Candidate = strings.TrimPrefix(candidateFrame.Candidate, "a=")
	candidateFrame.Candidate = strings.TrimSuffix(candidateFrame.Candidate, "\r\n")

	response := &types.WsMessage{
		AgentID:   link.AgentID,
		Type:      "candidate",
		SessionID: msg.Data.Header.SessionID,
		Payload:   candidateFrame.Candidate,
		Success:   true,
	}

	sendBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("marshal http response fail: %s", err.Error())

		return
	}

	log.Printf("ws write: %s", string(sendBytes))

	// body send back to Javascript
	err = link.Conn.WriteMessage(websocket.TextMessage, sendBytes)
	if err != nil {
		log.Printf("ws write fail: %s", err.Error())
	}
}

func handleDisconnect(msg *MqttMessage, link *types.WsMessage) {
	frame, ok := msg.Data.Message.(json.RawMessage)
	if !ok {
		log.Printf("convert interface{} to []byte fail, session: %s", msg.Data.Header.SessionID)

		return
	}

	disconnectFrame := struct {
		Mode string `json:"mode"`
	}{}

	if err := json.Unmarshal(frame, &disconnectFrame); err != nil {
		log.Printf("unmarshal mqtt disconnect frame fail: %s, session: %s, frame: %s", err.Error(),
			msg.Data.Header.SessionID,
			string(msg.Data.Message.([]byte)))

		return
	}

	response := &types.WsMessage{
		AgentID:   link.AgentID,
		Type:      "disconnect",
		SessionID: msg.Data.Header.SessionID,
		Payload:   "",
		Success:   true,
	}

	sendBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("marshal http response fail: %s", err.Error())

		return
	}

	log.Printf("ws write: %s", string(sendBytes))

	// body send back to Javascript
	err = link.Conn.WriteMessage(websocket.TextMessage, sendBytes)
	if err != nil {
		log.Printf("ws write fail: %s", err.Error())
	}

	// disconnect session, then link should be removed
	bootstrap.RemoveLink(msg.Data.Header.SessionID)
}
