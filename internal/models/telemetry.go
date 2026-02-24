package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TelemetryEvent struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	UserID primitive.ObjectID `bson:"user_id"`
	CaseID string             `bson:"case_id"`
	Puzzle int                `bson:"puzzle_id"`

	Timestamp time.Time `bson:"timestamp"`

	InputType string `bson:"input_type"`

	Query         string          `bson:"query,omitempty"`
	QueryFeatures map[string]bool `bson:"query_features,omitempty"`

	Command       string `bson:"command,omitempty"`
	CommandTarget string `bson:"command_target,omitempty"`

	Result TelemetryResult `bson:"result"`

	FocusState string `bson:"focus_state"`
}

type TelemetryResult struct {
	Status    string `bson:"status"`
	ErrorType string `bson:"error_type"`
	DBChanged bool   `bson:"db_changed"`
}
