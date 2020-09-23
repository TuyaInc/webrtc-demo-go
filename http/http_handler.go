package http

import (
	"log"
	"net/http"
)

// ListenAndServe 托管Web资源，供浏览器访问，最好使用Chrome浏览器
func ListenAndServe() {
	fs := http.FileServer(http.Dir("./static"))

	http.Handle("/", fs)

	log.Print("web server listen on :3333...")

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Printf("web serve fail: %s", err.Error())
	}
}
