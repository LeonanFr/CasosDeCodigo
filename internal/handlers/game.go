package handlers

import (
	"casos-de-codigo-api/internal/auth"
	"casos-de-codigo-api/internal/db"
	"casos-de-codigo-api/internal/engine"
	"casos-de-codigo-api/internal/models"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ctx := r.Context()

	userID, _ := auth.GetUserIDFromContext(ctx)

	var req models.ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Requisição inválida"}`, http.StatusBadRequest)
		return
	}

	var userPtr *primitive.ObjectID
	var teamPtr *string
	isTournament := false

	if req.TeamCode != nil && *req.TeamCode != "" {
		teamPtr = req.TeamCode
		isTournament = true
	} else {
		userPtr = &userID
	}

	caso, err := h.MongoManager.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, `{"error":"Caso não encontrado"}`, http.StatusNotFound)
		return
	}

	progression, err := h.MongoManager.GetProgression(
		req.CaseID,
		userPtr,
		teamPtr,
	)
	if err != nil {
		http.Error(w, `{"error":"Erro ao buscar progresso"}`, http.StatusInternalServerError)
		return
	}

	if progression == nil {
		progression = &models.Progression{
			UserID:        userPtr,
			TeamCode:      teamPtr,
			CaseID:        req.CaseID,
			CurrentPuzzle: caso.Config.StartingPuzzle,
			CurrentFocus:  "none",
			SQLHistory:    []models.SQLHistoryItem{},
			Active:        true,
		}

		if err := h.MongoManager.UpsertProgression(progression); err != nil {
			http.Error(w, `{"error":"Erro ao criar progresso"}`, http.StatusInternalServerError)
			return
		}
	}

	cleanSQL := strings.ToUpper(strings.TrimSpace(req.SQL))
	if cleanSQL == "RESET" {
		_ = h.MongoManager.ResetProgression(
			req.CaseID,
			userPtr,
			teamPtr,
			caso.Config.StartingPuzzle,
		)

		err = json.NewEncoder(w).Encode(models.GameResponse{
			Success:   true,
			IsReset:   true,
			Narrative: "Progresso resetado.",
			State: models.GameState{
				CaseID:        req.CaseID,
				CurrentPuzzle: caso.Config.StartingPuzzle,
				CurrentFocus:  "none",
			},
		})
		if err != nil {
			return
		}
		return
	}

	response, historyItem, err := h.GameProcessor.ProcessCommand(
		caso,
		progression,
		req.SQL,
	)

	if err != nil {
		http.Error(w, `{"error":"Erro interno"}`, http.StatusInternalServerError)
		return
	}

	if response.Success {
		progression.CurrentPuzzle = response.State.CurrentPuzzle
		progression.CurrentFocus = response.State.CurrentFocus

		if progression.CurrentPuzzle >= len(caso.Puzzles) {
			progression.Completed = true
		}

		_ = h.MongoManager.UpsertProgression(progression)

		if historyItem != nil && historyItem.Query != "RESET_CASE" && !response.IsDebug {
			progression.SQLHistory = append(progression.SQLHistory, *historyItem)
			_ = h.MongoManager.UpsertProgression(progression)
		}
	}

	if !isTournament {
		event := &models.TelemetryEvent{
			UserID:    userID,
			CaseID:    req.CaseID,
			Timestamp: time.Now(),
			InputType: "game_command",
		}
		_ = h.MongoManager.SaveTelemetry(event)
	}

	w.Header().Set("Content-Type", "application/json")
	if !response.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *GameHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Não autorizado"}`, http.StatusUnauthorized)
		return
	}

	teamCode := r.URL.Query().Get("team_code")
	var progressions []models.Progression
	var err error

	if teamCode != "" {
		progressions, err = h.MongoManager.GetTournamentProgressions(teamCode)
	} else {
		progressions, err = h.MongoManager.GetUserProgressions(userID)
	}

	if err != nil {
		http.Error(w, `{"error":"Erro ao buscar progresso"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(progressions)
}
