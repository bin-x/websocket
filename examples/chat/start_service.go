package main

import (
	//"fmt"
	"strconv"

	//"log"
	. "github.com/bin-x/websocket"
)

type App struct {
}

func (app *App) OnConnect(clientId string) {
	//fmt.Println("this app's Onconnect")
	message := clientId + " connected."
	for i := 0; i < 10000; i++ {
		if i == 300 {
			go Api.CloseClient("127.0.0.1:60776")
		}
		Api.SendToAll([]byte(message + strconv.Itoa(i)))
	}
}

func (app *App) OnMessage(ClientId string, message []byte) {
	//Api.CloseClient("127.0.0.1:60776")
}

func (app *App) OnClose(ClientId string) {
}

func main() {
	// 注册中心地址
	registerAddr := "localhost:8001"

	// 提供给集群中其他服务调用的rpc端口
	// 注意：安全起见，仅允许内网访问，请勿开放外网访问
	rpcPort := "8003"

	// 本地局域网ip地址，内网地址，可让其他机器访问到。
	lanIp := "127.0.0.1"

	// websocket的监听地址，供客户端访问。注意检查端口是否能够正常访问
	wsAddr := ":9003"

	hub := NewServiceHub(registerAddr, rpcPort, lanIp, &App{})
	hub.Start(wsAddr)
}
