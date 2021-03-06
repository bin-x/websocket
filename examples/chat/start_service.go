package main

import (

	//"log"
	. "github.com/bin-x/websocket"
)

type App struct {
}

func (app *App) OnConnect(clientId string) {
	message := "new client: " + clientId
	Api.SendToAll([]byte(message))
}

func (app *App) OnMessage(clientId string, message []byte) {
	s := []byte(clientId + ": ")
	Api.SendToAll(append(s, message...))
}

func (app *App) OnClose(clientId string) {
	s := []byte(clientId + " leave.")
	Api.SendToAll(s)
}

func main() {
	// 注册中心地址
	registerAddr := "localhost:8001"

	// 提供给集群中其他服务调用的rpc端口
	// 注意：安全起见，仅允许内网访问，请勿开放外网访问
	rpcPort := uint16(8003)

	// 本地局域网ip地址，内网地址，可让其他机器访问到。
	lanIp := "127.0.0.1"

	// websocket的监听地址，供客户端访问。注意检查端口是否能够正常访问
	wsAddr := ":9003"

	hub := NewServiceHub(registerAddr, rpcPort, lanIp, &App{})
	hub.Start(wsAddr)
}
