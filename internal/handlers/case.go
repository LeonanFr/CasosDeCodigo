package handlers

import (
	"casos-de-codigo-api/internal/auth"
	"casos-de-codigo-api/internal/db"
	"casos-de-codigo-api/internal/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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
	cases, err := h.MongoManager.GetAllCases()
	if err != nil {
		http.Error(w, `{"error": "Erro ao buscar casos"}`, http.StatusInternalServerError)
		return
	}

	summaries := make([]models.CaseSummary, 0)
	for _, c := range cases {
		summaries = append(summaries, models.CaseSummary{
			ID:          c.ID,
			Title:       c.Title,
			Description: c.Description,
			Difficulty:  c.Difficulty,
		})
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Cases []models.CaseSummary `json:"cases"`
		}{Cases: summaries})
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
		Cases:        summaries,
		Progressions: progressions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(models.InitializeResponse{
			Case: caso,
		})
		return
	}

	progression, err := h.MongoManager.GetProgression(userID, caseID)
	if err != nil {
		http.Error(w, `{"error": "Erro ao buscar progresso"}`, http.StatusInternalServerError)
		return
	}

	response := models.InitializeResponse{
		Progression: progression,
		Case:        caso,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *CaseHandler) InitializeCase(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Não autorizado"}`, http.StatusUnauthorized)
		return
	}

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

		if err := h.MongoManager.UpsertProgression(progression); err != nil {
			http.Error(w, `{"error": "Erro ao inicializar progresso"}`, http.StatusInternalServerError)
			return
		}
	}

	response := models.InitializeResponse{
		Progression: progression,
		Case:        caso,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
