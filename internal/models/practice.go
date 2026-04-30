package models

import "time"

type CoopDeck struct {
	ID          string   `bson:"_id" json:"id"`
	Name        string   `bson:"name" json:"name"`
	Description string   `bson:"description" json:"description"`
	CaseIDs     []string `bson:"case_ids" json:"case_ids"`
}

type PracticeRoom struct {
	ID        string    `bson:"_id" json:"id"`
	DeckID    string    `bson:"deck_id" json:"deck_id"`
	CaseIDs   []string  `bson:"case_ids" json:"case_ids"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Active    bool      `bson:"active" json:"active"`
}
