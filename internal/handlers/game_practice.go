package handlers

import (
	"casos-de-codigo-api/internal/models"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func generateRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
	rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (h *GameHandler) GetCoopDecks(w http.ResponseWriter, r *http.Request) {
	decks, err := h.MongoManager.GetCoopDecks()
	if err != nil {
		http.Error(w, `{"error": "Erro ao listar decks"}`, http.StatusInternalServerError)
		return
	}
	if decks == nil {
		decks = []models.CoopDeck{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decks)
}

func (h *GameHandler) CreatePracticeRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DeckID string `json:"deck_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Requisição inválida"}`, http.StatusBadRequest)
		return
	}
	if req.DeckID == "" {
		http.Error(w, `{"error":"Deck é obrigatório"}`, http.StatusBadRequest)
		return
	}

	decks, err := h.MongoManager.GetCoopDecks()
	if err != nil {
		http.Error(w, `{"error":"Erro ao validar deck"}`, http.StatusInternalServerError)
		return
	}
	var deck *models.CoopDeck
	for i := range decks {
		if decks[i].ID == req.DeckID {
			deck = &decks[i]
			break
		}
	}
	if deck == nil {
		http.Error(w, `{"error":"Deck não encontrado"}`, http.StatusNotFound)
		return
	}

	code := generateRoomCode()
	room := &models.PracticeRoom{
		ID:        code,
		DeckID:    deck.ID,
		CaseIDs:   deck.CaseIDs,
		CreatedAt: time.Now(),
		Active:    true,
	}
	if err := h.MongoManager.CreatePracticeRoom(room); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			room.ID = generateRoomCode()
			err = h.MongoManager.CreatePracticeRoom(room)
		}
		if err != nil {
			http.Error(w, `{"error":"Erro ao criar sala"}`, http.StatusInternalServerError)
			return
		}
	}

	resp := map[string]interface{}{
		"room_code": code,
		"cases":     deck.CaseIDs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *GameHandler) CheckPracticeRoom(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error":"Código obrigatório"}`, http.StatusBadRequest)
		return
	}

	room, err := h.MongoManager.GetPracticeRoom(code)
	if err != nil {
		http.Error(w, `{"error":"Sala não encontrada"}`, http.StatusNotFound)
		return
	}

	resp := map[string]interface{}{
		"room_code": room.ID,
		"case_ids":  room.CaseIDs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *GameHandler) PracticeRoomStatus(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error":"Código obrigatório"}`, http.StatusBadRequest)
		return
	}

	room, err := h.MongoManager.GetPracticeRoom(code)
	if err != nil {
		http.Error(w, `{"error":"Sala não encontrada"}`, http.StatusNotFound)
		return
	}

	allStarted := true
	for _, caseID := range room.CaseIDs {
		filter := bson.M{
			"team_code": code,
			"case_id":   caseID,
			"active":    true,
		}
		count, err := h.MongoManager.ProgressionColl.CountDocuments(r.Context(), filter)
		if err != nil || count == 0 {
			allStarted = false
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ready": allStarted})
}
