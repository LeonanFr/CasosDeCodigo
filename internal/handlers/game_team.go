package handlers

import (
	"casos-de-codigo-api/internal/integration"
	"casos-de-codigo-api/internal/models"
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
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

	var cases []models.CaseSummary
	if len(tournament.CaseIDs) > 0 {
		allCases, err := h.MongoManager.GetAllCases()
		if err == nil {
			for _, c := range allCases {
				for _, id := range tournament.CaseIDs {
					if c.ID == id {
						cases = append(cases, models.CaseSummary{
							ID:          c.ID,
							Title:       c.Title,
							Description: c.Description,
							Difficulty:  c.Difficulty,
						})
						break
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"valid":     teamResp.Exists,
		"team_code": code,
		"members":   teamResp.Integrates,
	})
}

func (h *GameHandler) TournamentStatus(w http.ResponseWriter, r *http.Request) {
	teamCode := r.URL.Query().Get("team_code")
	if teamCode == "" {
		http.Error(w, `{"error":"team_code obrigatório"}`, http.StatusBadRequest)
		return
	}

	tournament, err := h.MongoManager.GetActiveTournament()
	if err != nil || tournament == nil {
		http.Error(w, `{"error":"torneio indisponível"}`, http.StatusServiceUnavailable)
		return
	}

	var casesStatus []map[string]interface{}
	allStarted := true

	for _, caseID := range tournament.CaseIDs {
		filter := bson.M{"team_code": teamCode, "case_id": caseID}
		count, err := h.MongoManager.ProgressionColl.CountDocuments(context.Background(), filter)
		started := err == nil && count > 0
		if !started {
			allStarted = false
		}
		casesStatus = append(casesStatus, map[string]interface{}{
			"case_id": caseID,
			"started": started,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ready": allStarted,
		"cases": casesStatus,
	})
}
