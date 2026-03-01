package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tournament struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	APIConfig APIConfig          `bson:"api_config" json:"api_config"`
	Active    bool               `bson:"active" json:"active"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type APIConfig struct {
	TeamValidateRoute string `bson:"team_validate_route" json:"team_validate_route"`
	PuzzleEventRoute  string `bson:"puzzle_event_route" json:"puzzle_event_route"`
	CaseEventRoute    string `bson:"case_event_route" json:"case_event_route"`
}
