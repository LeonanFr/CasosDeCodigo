package handlers

import (
	"casos-de-codigo-api/internal/chat"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var chatUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type ChatMessage struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	TeamCode  string             `json:"team_code" bson:"team_code"`
	User      string             `json:"user" bson:"user"`
	Matricula string             `json:"matricula" bson:"matricula"`
	Message   string             `json:"message" bson:"message"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
}

func (h *GameHandler) ChatWebSocket(w http.ResponseWriter, r *http.Request) {
	teamCode := r.URL.Query().Get("team_code")
	if teamCode == "" {
		http.Error(w, "team_code obrigatório", http.StatusBadRequest)
		return
	}

	conn, err := chatUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro ao fazer upgrade para websocket (chat): %v", err)
		return
	}
	defer conn.Close()

	client := chat.RegisterChatClient(teamCode, conn)
	defer chat.UnregisterChatClient(client)

	go func() {
		for {
			select {
			case message, ok := <-client.Send:
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				conn.WriteMessage(websocket.TextMessage, message)
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming struct {
			User      string `json:"user"`
			Matricula string `json:"matricula"`
			Message   string `json:"message"`
		}
		if err := json.Unmarshal(msg, &incoming); err != nil {
			continue
		}

		chatMsg := ChatMessage{
			TeamCode:  teamCode,
			User:      incoming.User,
			Matricula: incoming.Matricula,
			Message:   incoming.Message,
			Timestamp: time.Now(),
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = h.MongoManager.ChatMessagesColl.InsertOne(ctx, chatMsg)
		}()

		data, _ := json.Marshal(chatMsg)
		chat.BroadcastToChatTeam(teamCode, data)
	}
}
