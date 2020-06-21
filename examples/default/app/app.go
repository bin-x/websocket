package main

import (
	. "github.com/bin-x/websocket"
)

type App struct {
}

// called when new websocket client connected. Only called once per connection.
// @params clientId will create by system. this is unique in the entire distributed system
func (app *App) OnConnect(clientId string) {
	Api.SendToAll([]byte("hello"))
}

// Called every time the service receives a message from the client
func (app *App) OnMessage(clientId string, message []byte) {
}

// called before close the connection. Only called once per connection.
func (app *App) OnClose(clientId string) {
}
