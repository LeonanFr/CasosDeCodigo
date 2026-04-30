package models

type CommandOutcome struct {
	Response      *GameResponse
	HistoryItem   *SQLHistoryItem
	OldPuzzle     int
	NewPuzzle     int
	CaseCompleted bool
}
