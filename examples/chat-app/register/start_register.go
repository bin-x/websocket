package main

import (
	"github.com/bin-x/websocket"
)

func main() {
	// 注册中心监听地址
	addr := ":8101"

	hub := websocket.NewRegisterHub()
	hub.Start(addr)
}
