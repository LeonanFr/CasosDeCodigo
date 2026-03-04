package handlers

import (
	"casos-de-codigo-api/internal/auth"
	"casos-de-codigo-api/internal/db"
	"casos-de-codigo-api/internal/models"
	"casos-de-codigo-api/internal/ws"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CaseHandler struct {
	MongoManager *db.MongoManager
}

func NewCaseHandler(mongo *db.MongoManager) *CaseHandler {
	return &CaseHandler{
		MongoManager: mongo,
	}
}

func (h *CaseHandler) GetAllCases(w http.ResponseWriter, r *http.Request) {
	allCases, err := h.MongoManager.GetAllCases()
	if err != nil {
		http.Error(w, `{"error": "Erro ao buscar casos"}`, http.StatusInternalServerError)
		return
	}

	cases := make([]models.CaseSummary, 0)
	for _, c := range allCases {
		if c.TournamentID == "" {
			cases = append(cases, models.CaseSummary{
				ID:          c.ID,
				Title:       c.Title,
				Description: c.Description,
				Difficulty:  c.Difficulty,
			})
		}
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Cases []models.CaseSummary `json:"cases"`
		}{Cases: cases})
		return
	}

	progressions, err := h.MongoManager.GetUserProgressions(userID)
	if err != nil {
		progressions = []models.Progression{}
	}

	response := struct {
		Cases        []models.CaseSummary `json:"cases"`
		Progressions []models.Progression `json:"progressions,omitempty"`
	}{
		Cases:        cases,
		Progressions: progressions,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *CaseHandler) GetCase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	caseID := vars["id"]

	caso, err := h.MongoManager.GetCase(caseID)
	if err != nil {
		http.Error(w, `{"error": "Caso não encontrado"}`, http.StatusNotFound)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.InitializeResponse{
			Case: caso,
		})
		return
	}

	progression, err := h.MongoManager.GetProgression(caseID, &userID, nil, nil)
	if err != nil {
		http.Error(w, `{"error": "Erro ao buscar progresso"}`, http.StatusInternalServerError)
		return
	}

	response := models.InitializeResponse{
		Progression: progression,
		Case:        caso,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *CaseHandler) InitializeCase(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	sessionID, _ := auth.GetSessionIDFromContext(r.Context())

	var req models.InitializeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Requisição inválida"}`, http.StatusBadRequest)
		return
	}

	caso, err := h.MongoManager.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, `{"error": "Caso não encontrado"}`, http.StatusNotFound)
		return
	}

	var userPtr *primitive.ObjectID
	var teamPtr *string
	var matriculaPtr *string
	var sessionPtr *primitive.ObjectID

	if req.TeamCode != nil && *req.TeamCode != "" {
		if req.Matricula == "" {
			http.Error(w, `{"error": "Matrícula é obrigatória para modo torneio"}`, http.StatusBadRequest)
			return
		}
		teamPtr = req.TeamCode
		matriculaPtr = &req.Matricula
		sessionPtr = &sessionID
	} else if ok {
		userPtr = &userID
	} else {
		http.Error(w, `{"error": "Identificação necessária"}`, http.StatusBadRequest)
		return
	}

	if teamPtr != nil {
		ms, err := h.MongoManager.GetMemberSessionBySessionID(*teamPtr, sessionID)
		if err != nil {
			http.Error(w, `{"error": "Erro ao verificar reserva de matrícula"}`, http.StatusInternalServerError)
			return
		}
		if ms == nil || ms.Matricula != req.Matricula {
			http.Error(w, `{"error": "Matrícula não reservada para esta sessão"}`, http.StatusForbidden)
			return
		}
	}

	progression, err := h.MongoManager.GetProgression(req.CaseID, userPtr, teamPtr, matriculaPtr)
	if err != nil {
		http.Error(w, `{"error": "Erro ao buscar progresso"}`, http.StatusInternalServerError)
		return
	}

	if progression != nil {
		if teamPtr != nil {
			if progression.Active {
				if progression.SessionID != primitive.NilObjectID && progression.SessionID != sessionID {
					http.Error(w, `{"error": "Esta conta já está em uso em outra sessão."}`, http.StatusConflict)
					return
				}
			} else {
				progression.Active = true
				progression.SessionID = sessionID
				if err := h.MongoManager.UpsertProgression(progression); err != nil {
					http.Error(w, `{"error": "Erro ao reativar progresso"}`, http.StatusInternalServerError)
					return
				}
			}
		}
	} else {

		if teamPtr != nil {
			ms, err := h.MongoManager.GetMemberSessionBySessionID(*teamPtr, sessionID)
			if err != nil {
				http.Error(w, `{"error": "Erro ao verificar reserva de matrícula"}`, http.StatusInternalServerError)
				return
			}
			if ms == nil || ms.Matricula != req.Matricula {
				http.Error(w, `{"error": "Matrícula não reservada para esta sessão"}`, http.StatusForbidden)
				return
			}

			count, err := h.MongoManager.CountAllProgressionsByMatricula(*teamPtr, req.Matricula)
			if err != nil {
				http.Error(w, `{"error": "Erro ao verificar disponibilidade da matrícula"}`, http.StatusInternalServerError)
				return
			}
			if count > 0 {
				http.Error(w, `{"error": "Esta matrícula já está vinculada a outro caso."}`, http.StatusConflict)
				return
			}

			filter := bson.M{"team_code": *teamPtr, "case_id": req.CaseID}
			count, err = h.MongoManager.ProgressionColl.CountDocuments(r.Context(), filter)
			if err != nil {
				http.Error(w, `{"error": "Erro ao verificar disponibilidade do caso"}`, http.StatusInternalServerError)
				return
			}
			if count > 0 {
				http.Error(w, `{"error": "Esta linha narrativa já foi escolhida por outro membro do time."}`, http.StatusConflict)
				return
			}

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
				http.Error(w, `{"error": "Erro ao inicializar progresso"}`, http.StatusInternalServerError)
				return
			}
			event := map[string]string{"case_id": req.CaseID, "status": "occupied"}
			data, _ := json.Marshal(event)
			ws.BroadcastToTeam(*teamPtr, data)
		} else {
			progression = &models.Progression{
				UserID:        userPtr,
				CaseID:        req.CaseID,
				CurrentPuzzle: caso.Config.StartingPuzzle,
				CurrentFocus:  "none",
				SQLHistory:    []models.SQLHistoryItem{},
				Active:        true,
				Completed:     false,
			}
			if err := h.MongoManager.UpsertProgression(progression); err != nil {
				http.Error(w, `{"error": "Erro ao inicializar progresso"}`, http.StatusInternalServerError)
				return
			}
		}
	}

	response := models.InitializeResponse{
		Progression: progression,
		Case:        caso,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
