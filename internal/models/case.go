package models

import "time"

type Case struct {
	ID                string             `bson:"_id" json:"id"`
	Title             string             `bson:"title" json:"title"`
	Description       string             `bson:"description" json:"description"`
	Difficulty        string             `bson:"difficulty" json:"difficulty"`
	Order             int                `bson:"order" json:"order"`
	Version           int                `bson:"version" json:"version"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
	Config            CaseConfig         `bson:"config" json:"config"`
	Puzzles           []Puzzle           `bson:"puzzles" json:"puzzles"`
	Schemas           []Schema           `bson:"schemas" json:"schemas"`
	CommandResponses  []CommandResponse  `bson:"command_responses" json:"command_responses"`
	Validations       []Validation       `bson:"validations" json:"validations"`
	FocusRequirements []FocusRequirement `bson:"focus_requirements" json:"focus_requirements"`
	SQLFunctions      []SQLFunction      `bson:"sql_functions" json:"sql_functions"`
	HelpTexts         []HelpText         `bson:"help_texts" json:"help_texts"`
}

type CaseSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
}

type CaseConfig struct {
	StartingPuzzle int      `bson:"starting_puzzle" json:"starting_puzzle"`
	Interactables  []string `bson:"interactables" json:"interactables"`
}

type Puzzle struct {
	Number    int      `bson:"number" json:"number"`
	Narrative string   `bson:"narrative" json:"narrative"`
	ImageKey  string   `bson:"image_key,omitempty" json:"image_key,omitempty"`
	Tables    []string `bson:"tables" json:"tables"`
	Commands  []string `bson:"commands" json:"commands"`
}

type Schema struct {
	Puzzle    int    `bson:"puzzle" json:"puzzle"`
	TableName string `bson:"table_name" json:"table_name"`
	CreateSQL string `bson:"create_sql" json:"create_sql"`
	InsertSQL string `bson:"insert_sql" json:"insert_sql"`
}

type CommandResponse struct {
	Command   string `bson:"command" json:"command"`
	Condition string `bson:"condition" json:"condition"`
	Value     string `bson:"value" json:"value"`
	Response  string `bson:"response" json:"response"`
	ImageKey  string `bson:"image_key,omitempty" json:"image_key,omitempty"`
}

type Validation struct {
	Puzzle           int    `json:"puzzle" bson:"puzzle"`
	Type             string `json:"type" bson:"type"`
	CheckSQL         string `json:"check_sql" bson:"check_sql"`
	ExpectValue      string `json:"expect_value" bson:"expect_value"`
	SuccessNarrative string `json:"success_narrative" bson:"success_narrative"`
	SuccessImageKey  string `json:"success_image_key,omitempty" bson:"success_image_key,omitempty"`
	FailureNarrative string `json:"failure_narrative,omitempty" bson:"failure_narrative,omitempty"`
	FailureImageKey  string `json:"failure_image_key,omitempty" bson:"failure_image_key,omitempty"`
	UnlocksNext      bool   `json:"unlocks_next" bson:"unlocks_next"`
	NextPuzzle       int    `json:"next_puzzle" bson:"next_puzzle"`
}

type FocusRequirement struct {
	Puzzle        int      `bson:"puzzle" json:"puzzle"`
	CommandTypes  []string `bson:"command_types" json:"command_types"`
	RequiredFocus string   `bson:"required_focus" json:"required_focus"`
	ErrorMessage  string   `bson:"error_message" json:"error_message"`
}

type SQLFunction struct {
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	Example     string `bson:"example" json:"example"`
}

type HelpText struct {
	Puzzle  int    `bson:"puzzle" json:"puzzle"`
	Topic   string `bson:"topic" json:"topic"`
	Content string `bson:"content" json:"content"`
}
