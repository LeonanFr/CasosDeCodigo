package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	TeamCode string
	Conn     *websocket.Conn
	Send     chan []byte
}

var (
	clients   = make(map[string]map[*Client]bool)
	clientsMu sync.RWMutex
)

func RegisterClient(teamCode string, conn *websocket.Conn) *Client {
	client := &Client{
		TeamCode: teamCode,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if clients[teamCode] == nil {
		clients[teamCode] = make(map[*Client]bool)
	}
	clients[teamCode][client] = true
	return client
}

func UnregisterClient(client *Client) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if set, ok := clients[client.TeamCode]; ok {
		delete(set, client)
		if len(set) == 0 {
			delete(clients, client.TeamCode)
		}
	}
	close(client.Send)
}

func BroadcastToTeam(teamCode string, message []byte) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	for client := range clients[teamCode] {
		select {
		case client.Send <- message:
		default:
		}
	}
}
