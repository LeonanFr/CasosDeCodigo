package handlers

import (
	"casos-de-codigo-api/internal/integration"
	"casos-de-codigo-api/internal/ws"
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

	var cases []map[string]interface{}
	if len(tournament.CaseIDs) > 0 {
		allCases, err := h.MongoManager.GetAllCases()
		if err == nil {
			for _, id := range tournament.CaseIDs {
				for _, c := range allCases {
					if c.ID == id {
						filter := bson.M{"team_code": code, "case_id": id}
						count, _ := h.MongoManager.ProgressionColl.CountDocuments(context.Background(), filter)
						occupied := count > 0
						cases = append(cases, map[string]interface{}{
							"id":          c.ID,
							"title":       c.Title,
							"description": c.Description,
							"difficulty":  c.Difficulty,
							"occupied":    occupied,
						})
						break
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"valid":     teamResp.Exists,
		"team_code": code,
		"members":   teamResp.Integrates,
		"cases":     cases,
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

func (h *GameHandler) LeaveCase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CaseID    string `json:"case_id"`
		TeamCode  string `json:"team_code"`
		Matricula string `json:"matricula"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Requisição inválida"}`, http.StatusBadRequest)
		return
	}

	progression, err := h.MongoManager.GetProgression(req.CaseID, nil, &req.TeamCode, &req.Matricula)
	if err != nil || progression == nil {
		http.Error(w, `{"error":"Progressão não encontrada"}`, http.StatusNotFound)
		return
	}

	progression.Active = false
	progression.SessionID = primitive.NilObjectID
	if err := h.MongoManager.UpsertProgression(progression); err != nil {
		http.Error(w, `{"error":"Erro ao atualizar"}`, http.StatusInternalServerError)
		return
	}

	filter := bson.M{"team_code": req.TeamCode, "case_id": req.CaseID}
	count, _ := h.MongoManager.ProgressionColl.CountDocuments(context.Background(), filter)
	if count == 0 {
		event := map[string]string{"case_id": req.CaseID, "status": "free"}
		data, _ := json.Marshal(event)
		ws.BroadcastToTeam(req.TeamCode, data)
	}

	w.WriteHeader(http.StatusNoContent)
}
