package handlers

import (
	"casos-de-codigo-api/internal/integration"
	"encoding/json"
	"net/http"
)

type TeamValidateRequest struct {
	Code string `json:"code"`
}

func (h *GameHandler) ValidateTeam(w http.ResponseWriter, r *http.Request) {
	var req TeamValidateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"payload inválido"}`, http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		http.Error(w, `{"error":"código obrigatório"}`, http.StatusBadRequest)
		return
	}

	tournament, err := h.MongoManager.GetActiveTournament()
	if err != nil || tournament == nil {
		http.Error(w, `{"error":"torneio indisponível"}`, http.StatusServiceUnavailable)
		return
	}

	teamResp, err := integration.FetchTeam(tournament.CodeRoute, req.Code)
	if err != nil {
		http.Error(w, `{"error":"falha ao validar time"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]any{
		"valid":     teamResp.Exists,
		"team_code": req.Code,
		"members":   teamResp.Integrates,
	})
	if err != nil {
		return
	}
}
