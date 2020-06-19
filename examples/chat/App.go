package main

import (
	"fmt"
	"log"
	. "websocket"
)

type App struct {
}

func (app *App) OnConnect(clientId string) {
	fmt.Println("this app's Onconnect")
	message := clientId + " connected."
	Api.SendToAll([]byte(message))
}

func (app *App) OnMessage(ClientId string, message []byte) {
	Api.JoinGroup(ClientId, "1")
	log.Println("group:", Api.GetAllGroups())
	log.Println("group 1:", Api.GetClientIdsByGroup("1"))
	Api.SendToAll(message)
}

func (app *App) OnClose(ClientId string) {
}
