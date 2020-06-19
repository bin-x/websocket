// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"log"

	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type RegisterClient struct {
	hub *RegisterHub

	// The websocket connection.
	conn *websocket.Conn

	rpcAddr string
	wsAddr  string

	// Buffered channel of outbound messages.
	send chan []byte
}

type RegisterMessage struct {
	Action    string   `json:"action"`
	Data      string   `json:"data"`
	RpcAddr   string   `json:"rpc_addr"`
	Addresses []string `json:"addresses"`
}

func (c *RegisterClient) read() {
	defer func() {
		c.hub.close <- c
		c.conn.Close()

	}()
	// 设置超时时间，如果收到pong消息，则自动延长时间
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(registerPongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(registerPongWait))
		log.Println("receive pong from ", c.conn.RemoteAddr().String())
		return nil
	})
	for {
		var message RegisterMessage
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		switch message.Action {
		//after the service send ip, rpc port, websocket port,
		//the register hub will save this service,
		//than other services can find this one
		case registerActionConnect:
			if !HostAddrCheck(message.RpcAddr) {
				log.Println("error: rpc address error")
				break
			}

			c.rpcAddr = message.RpcAddr
			c.hub.connect <- c
		}
	}
}

func (c *RegisterClient) write() {
	ticker := time.NewTicker(registerPingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		close(c.send)
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}

		case <-ticker.C:
			//定时发送ping消息给客户端,client默认pinghander为返回pong消息
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("can't send ping to ", c.conn.RemoteAddr().String())
				return
			}
			log.Println("ping to ", c.conn.RemoteAddr().String())
		}
	}
}

func (c *RegisterClient) close() {
	c.hub.close <- c
	c.conn.Close()
	close(c.send)
}

// serveWs handles websocket requests from the peer.
func registerServeWs(hub *RegisterHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("new client: ", conn.RemoteAddr().String())
	client := &RegisterClient{hub: hub, conn: conn, send: make(chan []byte, 256)}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.write()
	go client.read()
}
