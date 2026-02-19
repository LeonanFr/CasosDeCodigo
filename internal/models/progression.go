package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Progression struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	CaseID        string             `bson:"case_id" json:"case_id"`
	CurrentPuzzle int                `bson:"current_puzzle" json:"current_puzzle"`
	CurrentFocus  string             `bson:"current_focus" json:"current_focus"`
	SQLHistory    []SQLHistoryItem   `bson:"sql_history" json:"sql_history"`

	PuzzleCheckpoints map[string]int `bson:"puzzle_checkpoints,omitempty" json:"puzzle_checkpoints,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	Completed bool      `bson:"completed" json:"completed"`
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
