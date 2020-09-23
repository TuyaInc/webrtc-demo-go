package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/webrtc-demo-go/bootstrap"
	"github.com/webrtc-demo-go/config"
	openmqtt "github.com/webrtc-demo-go/openapi/mqtt"
	"github.com/webrtc-demo-go/types"
	"log"
	"net/http"
)

// 此Sample下WebSocket传输的是json，Message Type为1(Text)
// 此Sample关闭了请求源地址检查，生产环境中应该开启
var upgrader = websocket.Upgrader{
	Subprotocols: []string{"json"},
	CheckOrigin:  checkOrigin,
}

// 关闭请求源地址检查
func checkOrigin(r *http.Request) bool {
	return true
}

// ListenAndServe 提供WebSocket服务入口/webrtc，由HTTP协议升级到WebSocket
func ListenAndServe() {
	http.HandleFunc("/webrtc", webrtc)

	log.Print("websocket server listen on :5555...")

	err := http.ListenAndServe(":5555", nil)
	if err != nil {
		log.Printf("websocket serve fail: %s", err.Error())
	}
}

// WebSocket的连接处理函数，在Golang中每个连接独享自己的协程（类似C++/Java中线程，更轻量化）
func webrtc(w http.ResponseWriter, r *http.Request) {
	// 升级连接协议到WebSocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade to websocket fail: %s", err.Error())

		return
	}
	defer c.Close()

	log.Printf("new ws client, addr: %s", r.RemoteAddr)

	// 保存当前WebSocket客户端的代理ID
	agentID := ""

	// 从WebSocket连接轮询消息
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("ws read fail: %s", err.Error())

			break
		}

		log.Printf("ws recv: %s", string(message))

		msg := &types.WsMessage{
			Conn: c,
		}

		err = json.Unmarshal(message, msg)
		if err != nil {
			log.Printf("unmarshal ws message fail: %s", err.Error())

			break
		}

		// 设置代理ID
		agentID = msg.AgentID

		// 增加会话id和WebSocket连接的映射关系
		bootstrap.AddLink(msg.AgentID, msg.SessionID, msg)

		dispatch(msg)
	}

	if agentID != "" {
		bootstrap.RemoveLinkByConnLost(agentID)
	}
}

func sendIceServers(c *websocket.Conn) {
	iceServers := &types.WsMessage{
		Type:    "webrtcConfigs",
		Payload: openmqtt.IceServers(),
		Success: true,
	}

	sendBytes, err := json.Marshal(iceServers)
	if err != nil {
		log.Printf("marshal iceServers fail: %s", err.Error())

		return
	}

	// iceServers send back to Javascript
	err = c.WriteMessage(1, sendBytes)
	if err != nil {
		log.Printf("ws write fail: %s", err.Error())
	}
}

func dispatch(msg *types.WsMessage) {
	switch config.App.OpenAPIMode {
	case "mqtt":
		// 浏览器网页每次点击Call时都会拉取webrtc configs
		if msg.Type == "webRTCConfigs" {
			if err := openmqtt.FetchWebRTCConfigs(); err != nil {
				log.Printf("%s fetch webrtc configs fail", msg.AgentID)
			} else {
				sendIceServers(msg.Conn)
			}
		} else {
			openmqtt.Post(msg)
		}
	default:
		log.Printf("OpenAPI webRTC only support [mqtt], mode: %s", config.App.OpenAPIMode)
	}
}
