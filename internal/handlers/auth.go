package handlers

import (
	"casos-de-codigo-api/internal/auth"
	"casos-de-codigo-api/internal/db"
	"casos-de-codigo-api/internal/models"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	MongoManager *db.MongoManager
	Validator    *validator.Validate
}

func NewAuthHandler(mongo *db.MongoManager) *AuthHandler {
	return &AuthHandler{
		MongoManager: mongo,
		Validator:    validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Requisição inválida"}`, http.StatusBadRequest)
		return
	}

	if err := h.Validator.Struct(req); err != nil {
		http.Error(w, `{"error": "Dados de registro inválidos"}`, http.StatusBadRequest)
		return
	}

	existingUser, _ := h.MongoManager.FindUserByUsername(req.Username)
	if existingUser != nil {
		http.Error(w, `{"error": "Usuário já existe"}`, http.StatusConflict)
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error": "Erro ao processar senha"}`, http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
	}

	if err := h.MongoManager.CreateUser(user); err != nil {
		http.Error(w, `{"error": "Erro ao criar usuário"}`, http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(user.ID.Hex(), user.Username)
	if err != nil {
		http.Error(w, `{"error": "Erro ao gerar token"}`, http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User: models.UserResponse{
			ID:       user.ID.Hex(),
			Username: user.Username,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Requisição inválida"}`, http.StatusBadRequest)
		return
	}

	if err := h.Validator.Struct(req); err != nil {
		http.Error(w, `{"error": "Dados de login inválidos"}`, http.StatusBadRequest)
		return
	}

	user, err := h.MongoManager.FindUserByUsername(req.Username)
	if err != nil {
		http.Error(w, `{"error": "Usuário ou senha inválidos"}`, http.StatusUnauthorized)
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, `{"error": "Usuário ou senha inválidos"}`, http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID.Hex(), user.Username)
	if err != nil {
		http.Error(w, `{"error": "Erro ao gerar token"}`, http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User: models.UserResponse{
			ID:       user.ID.Hex(),
			Username: user.Username,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Não autorizado"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.MongoManager.FindUserByID(userID)
	if err != nil {
		http.Error(w, `{"error": "Usuário não encontrado"}`, http.StatusNotFound)
		return
	}

	response := models.UserResponse{
		ID:       user.ID.Hex(),
		Username: user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
