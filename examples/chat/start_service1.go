package main

import (
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
	hub := NewServiceHub("localhost:8001", uint16(8004), "127.0.0.1", &App{})
	hub.Start(":9004")
}
