package models

type GameResponse struct {
	Success         bool        `json:"success"`
	Narrative       string      `json:"narrative,omitempty"`
	Data            interface{} `json:"data,omitempty"`
	Error           string      `json:"error,omitempty"`
	State           GameState   `json:"state,omitempty"`
	IsReset         bool        `json:"is_reset,omitempty"`
	IsDebug         bool        `json:"is_debug,omitempty"`
	ImageKey        string      `json:"image_key"`
	SuccessImageKey string      `json:"success_image_key,omitempty"`
	FailureImageKey string      `json:"failure_image_key,omitempty"`
}

type ExecuteRequest struct {
	CaseID string `json:"case_id" validate:"required"`
	SQL    string `json:"sql" validate:"required"`
}

type InitializeRequest struct {
	CaseID string `json:"case_id" validate:"required"`
}

type InitializeResponse struct {
	Progression *Progression `json:"progression"`
	Case        *Case        `json:"case"`
}

type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e APIError) Error() string {
	return e.Message
}

const (
	ErrInvalidSQL       = "INVALID_SQL"
	ErrFocusRequired    = "FOCUS_REQUIRED"
	ErrCaseNotFound     = "CASE_NOT_FOUND"
	ErrPlayerNotFound   = "PLAYER_NOT_FOUND"
	ErrValidationFailed = "VALIDATION_FAILED"
	ErrInternalError    = "INTERNAL_ERROR"
	ErrUnauthorized     = "UNAUTHORIZED"
	ErrInvalidToken     = "INVALID_TOKEN"
)
