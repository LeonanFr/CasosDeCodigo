package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Progression struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	UserID   *primitive.ObjectID `bson:"user_id,omitempty"`
	TeamCode *string             `bson:"team_code,omitempty"`

	CaseID string `bson:"case_id"`

	CurrentPuzzle int
	CurrentFocus  string

	SQLHistory []SQLHistoryItem

	PuzzleCheckpoints map[string]int

	Active    bool
	Completed bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

type SQLHistoryItem struct {
	Timestamp   time.Time `bson:"timestamp" json:"timestamp"`
	Query       string    `bson:"query" json:"query"`
	PuzzleState int       `bson:"puzzle_state" json:"puzzle_state"`
	FocusState  string    `bson:"focus_state" json:"focus_state"`
}

type GameState struct {
	CaseID        string   `json:"case_id"`
	CurrentPuzzle int      `json:"current_puzzle"`
	CurrentFocus  string   `json:"current_focus"`
	Tables        []string `json:"tables"`
	Commands      []string `json:"commands"`
	Narrative     string   `json:"narrative,omitempty"`
	ImageKey      string   `json:"image_key,omitempty"`
}
