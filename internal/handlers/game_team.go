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
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error":"código obrigatório"}`, http.StatusBadRequest)
		return
	}

	tournament, err := h.MongoManager.GetActiveTournament()
	if err != nil || tournament == nil {
		http.Error(w, `{"error":"torneio indisponível"}`, http.StatusServiceUnavailable)
		return
	}

	route := tournament.APIConfig.TeamValidateRoute
	if route == "" {
		http.Error(w, `{"error":"rota de validação indisponível"}`, http.StatusServiceUnavailable)
		return
	}

	teamResp, err := integration.FetchTeam(route, code)
	if err != nil {
		http.Error(w, `{"error":"falha ao validar time"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"valid":     teamResp.Exists,
		"team_code": code,
		"members":   teamResp.Integrates,
	})
}
