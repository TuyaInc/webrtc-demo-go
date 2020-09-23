package main

import (
	"github.com/webrtc-demo-go/bootstrap"
	"github.com/webrtc-demo-go/config"
	"github.com/webrtc-demo-go/http"
	openmqtt "github.com/webrtc-demo-go/openapi/mqtt"
	"github.com/webrtc-demo-go/websocket"
	"log"

	"sync"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var wg sync.WaitGroup

	wg.Add(1)

	startWebRTCSample()
	log.Print("start webRTC sample")

	wg.Wait()
}

func startWebRTCSample() {
	if err := config.LoadConfig(); err != nil {
		log.Printf("load webrtc.json to runtime fail: %s", err.Error())
		return
	}

	// 根据授权码获取开放平台服务Token，并定时更新Token
	if err := bootstrap.InitToken(); err != nil {
		log.Printf("init token fail: %s", err.Error())
		return
	}

	// mqtt接入开放平台前，需要先通过Restful接口获取相关的配置来启动mqtt客户端
	if config.App.OpenAPIMode == "mqtt" {
		if err := openmqtt.Start(); err != nil {
			log.Printf("start mqtt fail: %s", err.Error())

			return
		}
	}

	// 启动web server
	go http.ListenAndServe()

	// 启动websocket server
	go websocket.ListenAndServe()
}
