package handlers

import (
	"casos-de-codigo-api/internal/auth"
	"casos-de-codigo-api/internal/db"
	"casos-de-codigo-api/internal/engine"
	"casos-de-codigo-api/internal/models"
	"encoding/json"
	"net/http"
	"strings"
)

type GameHandler struct {
	MongoManager  *db.MongoManager
	SQLiteFactory *db.SQLiteFactory
	GameProcessor *engine.GameProcessor
}

func NewGameHandler(mongo *db.MongoManager, factory *db.SQLiteFactory) *GameHandler {
	return &GameHandler{
		MongoManager:  mongo,
		SQLiteFactory: factory,
		GameProcessor: engine.NewGameProcessor(factory),
	}
}

func (h *GameHandler) ExecuteCommand(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Não autorizado"}`, http.StatusUnauthorized)
		return
	}

	var req models.ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Requisição inválida"}`, http.StatusBadRequest)
		return
	}

	caso, err := h.MongoManager.GetCase(req.CaseID)
	if err != nil {
		if strings.Contains(err.Error(), "no documents") {
			http.Error(w, `{"error": "Caso não encontrado"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error": "Erro de conexão com o banco"}`, http.StatusInternalServerError)
		}
		return
	}

	progression, err := h.MongoManager.GetProgression(userID, req.CaseID)
	if err != nil {
		http.Error(w, `{"error": "Erro ao buscar progresso"}`, http.StatusInternalServerError)
		return
	}

	if progression == nil {
		progression = &models.Progression{
			UserID:        userID,
			CaseID:        req.CaseID,
			CurrentPuzzle: caso.Config.StartingPuzzle,
			CurrentFocus:  "none",
			SQLHistory:    []models.SQLHistoryItem{},
		}
		h.MongoManager.CreateProgression(progression)
	}

	cleanSQL := strings.ToUpper(strings.TrimSpace(req.SQL))
	if cleanSQL == "RESET" {
		h.MongoManager.ResetProgression(userID, req.CaseID, caso.Config.StartingPuzzle)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.GameResponse{
			Success:   true,
			Narrative: "Progresso resetado.",
			State: models.GameState{
				CaseID:        req.CaseID,
				CurrentPuzzle: caso.Config.StartingPuzzle,
				CurrentFocus:  "none",
			},
		})
		return
	}

	oldPuzzle := progression.CurrentPuzzle
	oldFocus := progression.CurrentFocus

	response, historyItem, err := h.GameProcessor.ProcessCommand(caso, progression, req.SQL)
	if err != nil {
		http.Error(w, `{"error": "Erro interno"}`, http.StatusInternalServerError)
		return
	}

	if response.Success {
		if oldFocus != response.State.CurrentFocus || oldPuzzle != response.State.CurrentPuzzle {
			h.MongoManager.UpdateUserProgress(userID, req.CaseID, response.State.CurrentPuzzle, response.State.CurrentFocus)
		}

		if historyItem != nil && historyItem.Query != "RESET_CASE" && !response.IsDebug {
			h.MongoManager.AddSQLHistory(userID, req.CaseID, *historyItem)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if !response.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(response)
}

func (h *GameHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Não autorizado"}`, http.StatusUnauthorized)
		return
	}
	progressions, err := h.MongoManager.GetUserProgressions(userID)
	if err != nil {
		http.Error(w, `{"error": "Erro ao buscar progresso"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progressions)
}
