package handlers

import (
	"casos-de-codigo-api/internal/auth"
	"casos-de-codigo-api/internal/sse"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ReserveMemberRequest struct {
	TeamCode  string `json:"team_code"`
	Matricula string `json:"matricula"`
}

func (h *GameHandler) ReserveMember(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := auth.GetSessionIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Sessão não identificada"}`, http.StatusUnauthorized)
		return
	}

	var req ReserveMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Requisição inválida"}`, http.StatusBadRequest)
		return
	}

	err := h.MongoManager.ReserveMember(req.TeamCode, req.Matricula, sessionID)
	if err != nil {
		if err.Error() == "matrícula já está em uso por outra sessão" {
			http.Error(w, `{"error":"Esta matrícula já está em uso."}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"Erro ao reservar matrícula"}`, http.StatusInternalServerError)
		return
	}

	sse.NotifyMemberOccupied(req.TeamCode, req.Matricula)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *GameHandler) ReleaseMember(w http.ResponseWriter, r *http.Request) {
	sessionID, _ := auth.GetSessionIDFromContext(r.Context())
	var req ReserveMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Requisição inválida"}`, http.StatusBadRequest)
		return
	}
	err := h.MongoManager.ReleaseMember(req.TeamCode, req.Matricula, sessionID)
	if err != nil {
		http.Error(w, `{"error":"Erro ao liberar matrícula"}`, http.StatusInternalServerError)
		return
	}
	sse.NotifyMemberFree(req.TeamCode, req.Matricula)
	w.WriteHeader(http.StatusNoContent)
}

func (h *GameHandler) GetMyMatricula(w http.ResponseWriter, r *http.Request) {
	sessionID, _ := auth.GetSessionIDFromContext(r.Context())
	teamCode := r.URL.Query().Get("team_code")
	if teamCode == "" {
		http.Error(w, `{"error":"team_code obrigatório"}`, http.StatusBadRequest)
		return
	}

	ms, err := h.MongoManager.GetMemberSessionBySessionID(teamCode, sessionID)
	if err != nil {
		http.Error(w, `{"error":"Erro ao buscar reserva"}`, http.StatusInternalServerError)
		return
	}
	if ms == nil {
		http.Error(w, `{"error":"Nenhuma matrícula reservada"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"matricula": ms.Matricula,
	})
}

func (h *GameHandler) SubscribeMemberEvents(w http.ResponseWriter, r *http.Request) {
	teamCode := r.URL.Query().Get("team_code")
	if teamCode == "" {
		http.Error(w, "team_code obrigatório", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE não suportado", http.StatusInternalServerError)
		return
	}

	eventChan := sse.SubscribeMember(teamCode)
	defer sse.UnsubscribeMember(teamCode)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event := <-eventChan:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: member-status\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
