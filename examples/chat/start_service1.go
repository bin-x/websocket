package main

import (
	"fmt"
	"log"

	//"log"

	//"log"
	. "github.com/bin-x/websocket"
)

type App struct {
}

func (app *App) OnConnect(clientId string) {
	fmt.Println("this app's OnConnect")
	message := clientId + " connected."
	Api.SendToAll([]byte(message))
}

func (app *App) OnMessage(ClientId string, message []byte) {
	log.Println("OnMessage:", ClientId)
	Api.CloseClient(ClientId)
}

func (app *App) OnClose(ClientId string) {
}

func main() {
	hub := NewServiceHub("localhost:8001", "8004", "127.0.0.1", &App{})
	hub.Start(":9004")
}
