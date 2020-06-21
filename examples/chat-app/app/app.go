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
