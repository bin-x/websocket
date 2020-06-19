package websocket

type Application interface {
	OnConnect(string)
	OnMessage(string, []byte)
	OnClose(string)
}
