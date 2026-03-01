package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tournament struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	CodeRoute string             `bson:"code_route" json:"code_route"`
	Active    bool               `bson:"active" json:"active"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
