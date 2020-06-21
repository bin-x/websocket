package main

import (
	"log"
	"net/http"
)

func main() {
	// 监听地址，供客户端访问。注意检查端口是否能够正常访问
	addr := ":8082"

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "home.html")
	})

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
