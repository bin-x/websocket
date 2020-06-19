package app

type Application interface {
	OnConnect(string)
	OnMessage(string, []byte)
	OnClose(string)
}
