package websocket

import "time"

const (
	registerActionConnect       = "connect"
	registerActionBroadcastAddr = "broadcast_addresses"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	registerPongWait = 65 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	registerPingPeriod = 30 * time.Second

	pongWait = 30 * time.Second

	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)
