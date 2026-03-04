package chat

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
	chatClients   = make(map[string]map[*Client]bool)
	chatClientsMu sync.RWMutex
)

func RegisterChatClient(teamCode string, conn *websocket.Conn) *Client {
	client := &Client{
		TeamCode: teamCode,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}
	chatClientsMu.Lock()
	defer chatClientsMu.Unlock()
	if chatClients[teamCode] == nil {
		chatClients[teamCode] = make(map[*Client]bool)
	}
	chatClients[teamCode][client] = true
	return client
}

func UnregisterChatClient(client *Client) {
	chatClientsMu.Lock()
	defer chatClientsMu.Unlock()
	if set, ok := chatClients[client.TeamCode]; ok {
		delete(set, client)
		if len(set) == 0 {
			delete(chatClients, client.TeamCode)
		}
	}
	close(client.Send)
}

func BroadcastToChatTeam(teamCode string, message []byte) {
	chatClientsMu.RLock()
	defer chatClientsMu.RUnlock()
	for client := range chatClients[teamCode] {
		select {
		case client.Send <- message:
		default:
		}
	}
}
