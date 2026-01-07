package engine

import (
	"casos-de-codigo-api/internal/models"
	"strings"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateSQLCommand(caso *models.Case, progression *models.Progression, command string) error {
	upper := strings.ToUpper(command)

	for _, req := range caso.FocusRequirements {
		if req.Puzzle == progression.CurrentPuzzle {
			for _, cmdType := range req.CommandTypes {
				if strings.Contains(upper, cmdType) && progression.CurrentFocus != req.RequiredFocus {
					return models.APIError{
						Message: req.ErrorMessage,
						Code:    models.ErrFocusRequired,
					}
				}
			}
		}
	}

	return nil
}

func (v *Validator) IsValidCase(caso *models.Case) bool {
	if caso.ID == "" || caso.Title == "" {
		return false
	}

	if len(caso.Puzzles) == 0 {
		return false
	}

	for _, puzzle := range caso.Puzzles {
		if puzzle.Number < 1 {
			return false
		}
	}

	return true
}
