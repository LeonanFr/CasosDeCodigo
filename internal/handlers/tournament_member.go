package handlers

import (
	"casos-de-codigo-api/internal/auth"
	"casos-de-codigo-api/internal/ws"
	"encoding/json"
	"net/http"
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

	event := map[string]string{"matricula": req.Matricula, "status": "occupied", "sessionId": sessionID.Hex()}
	data, _ := json.Marshal(event)
	ws.BroadcastToTeam(req.TeamCode, data)

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
	event := map[string]string{"matricula": req.Matricula, "status": "free"}
	data, _ := json.Marshal(event)
	ws.BroadcastToTeam(req.TeamCode, data)

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
