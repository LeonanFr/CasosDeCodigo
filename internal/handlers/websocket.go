package handlers

import (
	"casos-de-codigo-api/internal/ws"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *GameHandler) TeamWebSocket(w http.ResponseWriter, r *http.Request) {
	teamCode := r.URL.Query().Get("team_code")
	if teamCode == "" {
		http.Error(w, "team_code obrigatório", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro ao fazer upgrade para websocket: %v", err)
		return
	}
	defer conn.Close()

	client := ws.RegisterClient(teamCode, conn)
	defer ws.UnregisterClient(client)

	reservas, err := h.MongoManager.GetActiveReservations(teamCode)
	if err != nil {
		log.Printf("Erro ao buscar reservas ativas: %v", err)
	} else {
		for _, r := range reservas {
			event := map[string]string{
				"matricula": r.Matricula,
				"status":    "occupied",
				"sessionId": r.SessionID.Hex(),
			}
			data, _ := json.Marshal(event)
			client.Send <- data
		}
	}

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
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
