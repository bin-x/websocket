package websocket

import (
	"encoding/json"
	"log"
	"net/http"
)

type RegisterHub struct {
	clients map[*RegisterClient]bool
	connect chan *RegisterClient
	close   chan *RegisterClient
}

func NewRegisterHub() *RegisterHub {
	return &RegisterHub{
		clients: make(map[*RegisterClient]bool),
		connect: make(chan *RegisterClient),
		close:   make(chan *RegisterClient),
	}
}

func (r *RegisterHub) run() {
	for {
		select {
		case client := <-r.connect:
			r.clients[client] = true
			r.broadcastServices()
		case client := <-r.close:
			delete(r.clients, client)

			r.broadcastServices()
		}
	}
}

func (r *RegisterHub) broadcastServices() {
	var addresses []string
	for client := range r.clients {
		addresses = append(addresses, client.rpcAddr)
	}
	message := RegisterMessage{
		Action:    registerActionBroadcastAddr,
		Addresses: addresses,
	}
	msg, err := json.Marshal(message)
	if err != nil {
		log.Println("error")
		return
	}

	log.Println("starting broadcast addresses")
	for client := range r.clients {
		select {
		case client.send <- msg:
		default:
			delete(r.clients, client)
		}
	}
	log.Println("broadcast addresses success")
}

func (r *RegisterHub) Start(addr string) {
	go r.run()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		registerServeWs(r, writer, request)
	})
	log.Println("starting register...")
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
