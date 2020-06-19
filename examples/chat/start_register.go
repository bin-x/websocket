package main

import (
	"websocket"
)

func main() {
	// 注册中心监听地址
	addr := ":8001"

	hub := websocket.NewRegisterHub()
	hub.Start(addr)
}
