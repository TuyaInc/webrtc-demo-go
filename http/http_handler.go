package http

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/webrtc-demo-go/bootstrap"
)

// ListenAndServe 托管Web资源，供浏览器访问，最好使用Chrome浏览器
func ListenAndServe() {
	fs := http.FileServer(http.Dir("./static"))

	http.Handle("/", fs)

	http.HandleFunc("/api/stream/rtsp", rtsp)
	http.HandleFunc("/api/stream/hls", hls)

	http.HandleFunc("/api/cloud/timeline", storageTimeline)
	http.HandleFunc("/api/cloud/hls", storageHls)
	http.HandleFunc("/api/cloud/key", storageKey)

	log.Print("web server listen on :3333...")

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Printf("web serve fail: %s", err.Error())
	}
}

// 从涂鸦开放平台获取RTSP播放资源
func rtsp(w http.ResponseWriter, r *http.Request) {
	url, err := bootstrap.GetRTSP()
	if err != nil {
		log.Printf("get rtsp fail: %s", err.Error())
		return
	}

	fmt.Fprint(w, url)
}

// 从涂鸦开放平台获取HLS实时流播放资源
func hls(w http.ResponseWriter, r *http.Request) {
	url, err := bootstrap.GetHLS()
	if err != nil {
		log.Printf("get hls fail: %s", err.Error())
		return
	}

	fmt.Fprint(w, url)
}

// 从涂鸦开放平台获取指定时间范围（秒）内云存储录像片段的起始时间信息
func storageTimeline(w http.ResponseWriter, r *http.Request) {
	timeGTStr := r.URL.Query().Get("timeGT")
	timeLTStr := r.URL.Query().Get("timeLT")

	timeGT, err := strconv.ParseInt(timeGTStr, 10, 64)
	timeLT, err := strconv.ParseInt(timeLTStr, 10, 64)

	timeline, err := bootstrap.GetCloudStorageTimeline(timeGT, timeLT)
	if err != nil {
		log.Printf("get cloud storage timeline fail: %s", err.Error())
		return
	}

	fmt.Fprint(w, timeline)
}

// 从涂鸦开放平台获取指定时间范围（秒）内云存储录像片段的HLS播放资源
func storageHls(w http.ResponseWriter, r *http.Request) {
	timeGTStr := r.URL.Query().Get("timeGT")
	timeLTStr := r.URL.Query().Get("timeLT")

	timeGT, err := strconv.ParseInt(timeGTStr, 10, 64)
	timeLT, err := strconv.ParseInt(timeLTStr, 10, 64)

	hls, err := bootstrap.GetCloudStorageHls(timeGT, timeLT)
	if err != nil {
		log.Printf("get cloud storage hls fail: %s", err.Error())
		return
	}

	fmt.Fprint(w, hls)
}

// 从涂鸦开放平台根据M3U8文件中的magic获取用于解密M3U8文件里TS流的Key，Key需要base64解码为原始二进制
func storageKey(w http.ResponseWriter, r *http.Request) {
	devID := r.URL.Query().Get("devId")
	magic := r.URL.Query().Get("magic")

	key, err := bootstrap.GetCloudStorageKey(magic)
	if err != nil {
		log.Printf("get storage hls key fail: %s", err.Error())
		return
	}

	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		log.Printf("base64 decode key fail: %s", err.Error())
		return
	}

	log.Printf("key : %s : %s, %s : [%+v]", devID, magic, key, data)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "binary/octet-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(data)
}
