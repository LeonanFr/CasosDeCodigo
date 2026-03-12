package handlers

import (
	"casos-de-codigo-api/internal/auth"
	"casos-de-codigo-api/internal/db"
	"casos-de-codigo-api/internal/engine"
	"casos-de-codigo-api/internal/integration"
	"casos-de-codigo-api/internal/models"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
	sessionID, _ := auth.GetSessionIDFromContext(ctx)

	var req models.ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Requisição inválida"}`, http.StatusBadRequest)
		return
	}

	var userPtr *primitive.ObjectID
	var teamPtr *string
	var matriculaPtr *string
	isTournament := false

	if req.TeamCode != nil && *req.TeamCode != "" {
		if req.Matricula == "" {
			http.Error(w, `{"error":"Matrícula é obrigatória para modo torneio"}`, http.StatusBadRequest)
			return
		}
		teamPtr = req.TeamCode
		matriculaPtr = &req.Matricula
		isTournament = true
	} else {
		userPtr = &userID
	}

	caso, err := h.MongoManager.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, `{"error":"Caso não encontrado"}`, http.StatusNotFound)
		return
	}

	progression, err := h.MongoManager.GetProgression(req.CaseID, userPtr, teamPtr, matriculaPtr)
	if err != nil {
		http.Error(w, `{"error":"Erro ao buscar progresso"}`, http.StatusInternalServerError)
		return
	}

	if progression == nil {
		progression = &models.Progression{
			UserID:        userPtr,
			TeamCode:      teamPtr,
			Matricula:     req.Matricula,
			SessionID:     sessionID,
			CaseID:        req.CaseID,
			CurrentPuzzle: caso.Config.StartingPuzzle,
			CurrentFocus:  "none",
			SQLHistory:    []models.SQLHistoryItem{},
			Active:        true,
			Completed:     false,
		}
		if err := h.MongoManager.UpsertProgression(progression); err != nil {
			http.Error(w, `{"error":"Erro ao criar progresso"}`, http.StatusInternalServerError)
			return
		}
	} else {
		if isTournament {
			if progression.SessionID != primitive.NilObjectID && progression.SessionID != sessionID {
				http.Error(w, `{"error":"Esta conta já está em uso em outra sessão."}`, http.StatusConflict)
				return
			}
			if !progression.Active {
				http.Error(w, `{"error":"Seu acesso foi bloqueado pelo administrador."}`, http.StatusForbidden)
				return
			}
		}
	}

	var tournament *models.Tournament
	if isTournament {
		tournament, err = h.MongoManager.GetActiveTournament()
		if err != nil || tournament == nil {
			http.Error(w, `{"error":"torneio indisponível"}`, http.StatusServiceUnavailable)
			return
		}

		allStarted := true
		for _, caseID := range tournament.CaseIDs {
			filter := bson.M{"team_code": *teamPtr, "case_id": caseID}
			count, err := h.MongoManager.ProgressionColl.CountDocuments(ctx, filter)
			if err != nil || count == 0 {
				allStarted = false
				break
			}
		}

		if !allStarted {
			cmd := strings.ToUpper(strings.TrimSpace(req.SQL))
			if cmd != "STATUS" && cmd != "CLS" && cmd != "AJUDA" {
				response := models.GameResponse{
					Success:   true,
					Narrative: "Aguardando seu time se conectar. Digite CLS quando todos estiverem prontos.",
					State:     h.GameProcessor.GetCurrentState(caso, progression),
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	cleanSQL := strings.ToUpper(strings.TrimSpace(req.SQL))
	if cleanSQL == "RESET" {
		err = h.MongoManager.ResetProgression(
			req.CaseID,
			userPtr,
			teamPtr,
			matriculaPtr,
			caso.Config.StartingPuzzle,
		)
		if err != nil {
			http.Error(w, `{"error":"Erro ao resetar progresso"}`, http.StatusInternalServerError)
			return
		}

		err := h.GameProcessor.ResetSession(progression)
		if err != nil {
			return
		}

		response := models.GameResponse{
			Success:   true,
			IsReset:   true,
			Narrative: "Progresso resetado.",
			State: models.GameState{
				CaseID:        req.CaseID,
				CurrentPuzzle: caso.Config.StartingPuzzle,
				CurrentFocus:  "none",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		return
	}

	outcome, err := h.GameProcessor.ProcessCommand(caso, progression, req.SQL)
	if err != nil {
		http.Error(w, `{"error":"Erro interno"}`, http.StatusInternalServerError)
		return
	}

	response := outcome.Response
	historyItem := outcome.HistoryItem

	if response.Success {

		if historyItem != nil && historyItem.Query != "" && !response.IsDebug {
			progression.SQLHistory = append(progression.SQLHistory, *historyItem)
		}

		if err := h.MongoManager.UpsertProgression(progression); err != nil {
			http.Error(w, `{"error":"Erro ao salvar progresso"}`, http.StatusInternalServerError)
			return
		}

		if isTournament && tournament != nil {

			if progression.PuzzlesEventSent == nil {
				progression.PuzzlesEventSent = make(map[int]bool)
			}

			for p := outcome.OldPuzzle; p < outcome.NewPuzzle; p++ {
				if !progression.PuzzlesEventSent[p] {
					_ = integration.SendPuzzleEvent(tournament, *teamPtr, req.Matricula)
					progression.PuzzlesEventSent[p] = true
				}
			}

			_ = h.MongoManager.UpsertProgression(progression)

			if outcome.CaseCompleted {

				caseBelongs := false
				for _, id := range tournament.CaseIDs {
					if id == caso.ID {
						caseBelongs = true
						break
					}
				}

				if caseBelongs {

					allCompleted, err := h.MongoManager.
						HasTeamCompletedAllTournamentCases(*teamPtr, tournament.CaseIDs)

					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					if allCompleted {
						_ = integration.SendCaseEvent(tournament, *teamPtr)
					}
				}
			}
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
	json.NewEncoder(w).Encode(response)
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

	if progressions == nil {
		progressions = []models.Progression{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(progressions)
}
