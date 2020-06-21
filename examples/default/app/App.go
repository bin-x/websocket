package main

import (
	. "github.com/bin-x/websocket"
)

type App struct {
}

func (app *App) OnConnect(clientId string) {
	Api.SendToAll([]byte("hello"))
}

func (app *App) OnMessage(ClientId string, message []byte) {
}

func (app *App) OnClose(ClientId string) {
}
