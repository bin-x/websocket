package websocket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var currentId uint32 = 0

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	id     string
	uid    string
	groups map[string]bool
	info   map[string]string

	hub *ServiceHub
	// The websocket connection.
	conn *websocket.Conn
	// Buffered channel of outbound messages.
	send       chan []byte
	joinGroup  chan string
	leaveGroup chan string
	setInfo    chan map[string]string
	updateInfo chan map[string]string
	done       chan bool
}

func NewServiceClient(hub *ServiceHub, conn *websocket.Conn) *Client {
	client := &Client{
		hub:        hub,
		conn:       conn,
		groups:     make(map[string]bool),
		info:       make(map[string]string),
		send:       make(chan []byte, 256),
		joinGroup:  make(chan string),
		leaveGroup: make(chan string),
		setInfo:    make(chan map[string]string),
		updateInfo: make(chan map[string]string),
		done:       make(chan bool),
	}
	client.generateId()
	return client
}

func (c *Client) generateId() {
	// 高并发时防止出现相同id
	id := atomic.AddUint32(&currentId, 1)
	c.id = AddressToClientId(c.hub.lanIp, c.hub.rpcPort, id)
}

func (c *Client) read() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
		log.Println("recover on read...")
	}()
	defer func() {
		c.done <- true
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.hub.application.OnMessage(c.id, message)
	}
}

func (c *Client) write() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)

		}
		log.Println("recover on write...")
	}()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.done <- true
	}()
	for {
		select {
		case message, ok := <-c.send:
			//log.Println("sending message: ", message)
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
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		}
	}
}

func (c *Client) run() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
		log.Println("recover on client run ...")
	}()
	defer c.close()
	defer c.hub.application.OnClose(c.id)

	for {
		select {
		case group := <-c.joinGroup:
			c.groups[group] = true
		case group := <-c.leaveGroup:
			delete(c.groups, group)
		case info := <-c.setInfo:
			c.info = info
		case infos := <-c.updateInfo:
			for k, v := range infos {
				c.info[k] = v
			}
		case <-c.done:
			return
		}
	}
}
func (c *Client) close() {
	c.hub.close <- c
	c.conn.Close()
	close(c.done)
	close(c.send)
	close(c.joinGroup)
	close(c.leaveGroup)
	close(c.setInfo)
	close(c.updateInfo)
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *ServiceHub, w http.ResponseWriter, r *http.Request) {

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// todo, 初始化client id
	client := NewServiceClient(hub, conn)
	hub.connect <- client

	log.Println("all client:", hub.clients)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.run()
	go client.write()
	go client.read()

	hub.application.OnConnect(client.id)
	log.Println("new client from " + conn.RemoteAddr().String())
}
