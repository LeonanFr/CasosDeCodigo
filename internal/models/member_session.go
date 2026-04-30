package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MemberSession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TeamCode  string             `bson:"team_code" json:"team_code"`
	Matricula string             `bson:"matricula" json:"matricula"`
	SessionID primitive.ObjectID `bson:"session_id" json:"session_id"`
	Active    bool               `bson:"active" json:"active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
