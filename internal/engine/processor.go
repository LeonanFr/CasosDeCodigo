package engine

import (
	"casos-de-codigo-api/internal/db"
	"casos-de-codigo-api/internal/models"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type GameProcessor struct {
	SQLiteFactory *db.SQLiteFactory
	Validator     *Validator
}

func NewGameProcessor(factory *db.SQLiteFactory) *GameProcessor {
	return &GameProcessor{
		SQLiteFactory: factory,
		Validator:     NewValidator(),
	}
}

func (p *GameProcessor) ProcessCommand(caso *models.Case, progression *models.Progression, command string) (*models.GameResponse, *models.SQLHistoryItem, error) {
	upperCommand := strings.ToUpper(strings.TrimSpace(command))

	if response := p.handleGameCommand(caso, progression, upperCommand); response != nil {
		return response, nil, nil
	}

	if err := p.Validator.ValidateSQLCommand(caso, progression, command); err != nil {
		if apiErr, ok := err.(models.APIError); ok {
			return &models.GameResponse{
				Success: false,
				Error:   apiErr.Message,
				State:   p.getCurrentState(caso, progression),
			}, nil, nil
		}
		return nil, nil, err
	}

	return p.executeSQL(caso, progression, command)
}

func (p *GameProcessor) handleGameCommand(caso *models.Case, progression *models.Progression, command string) *models.GameResponse {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	baseCmd := parts[0]

	if baseCmd == "AJUDA" || baseCmd == "HELP" || baseCmd == "/AJUDA" || baseCmd == "/HELP" {
		if len(parts) > 1 {
			topic := parts[1]
			for _, ht := range caso.HelpTexts {
				if strings.EqualFold(ht.Topic, topic) && (ht.Puzzle == 0 || ht.Puzzle == progression.CurrentPuzzle) {
					return &models.GameResponse{
						Success:   true,
						Narrative: ht.Content,
						State:     p.getCurrentState(caso, progression),
					}
				}
			}
		}
	}

	for _, resp := range caso.CommandResponses {
		isMatch := false
		if resp.Command == command {
			isMatch = true
		} else if strings.HasPrefix(command, "OLHAR") && strings.Contains(resp.Command, command) {
			isMatch = true
		} else if (baseCmd == "AJUDA" || baseCmd == "HELP") && resp.Command == baseCmd && len(parts) == 1 {
			isMatch = true
		}

		if isMatch && p.checkCondition(resp, progression) {
			newFocus := progression.CurrentFocus
			if strings.HasPrefix(command, "OLHAR") {
				if len(parts) > 1 {
					newFocus = strings.ToLower(parts[1])
				}
			}

			if command == "SAIR" || command == "FECHAR" || command == "PARAR" {
				newFocus = "none"
			}

			state := p.getCurrentState(caso, progression)
			state.CurrentFocus = newFocus

			return &models.GameResponse{
				Success:   true,
				Narrative: resp.Response,
				State:     state,
			}
		}
	}
	return nil
}

func (p *GameProcessor) checkCondition(resp models.CommandResponse, prog *models.Progression) bool {
	switch resp.Condition {
	case "always":
		return true
	case "puzzle_state":
		return fmt.Sprintf("%d", prog.CurrentPuzzle) == resp.Value
	case "puzzle_state_not":
		return fmt.Sprintf("%d", prog.CurrentPuzzle) != resp.Value
	case "puzzle_state_less":
		val := 0
		fmt.Sscanf(resp.Value, "%d", &val)
		return prog.CurrentPuzzle < val
	case "puzzle_state_greater":
		val := 0
		fmt.Sscanf(resp.Value, "%d", &val)
		return prog.CurrentPuzzle > val
	case "current_focus_none":
		return prog.CurrentFocus == "none"
	default:
		return false
	}
}

func (p *GameProcessor) executeSQL(caso *models.Case, progression *models.Progression, query string) (*models.GameResponse, *models.SQLHistoryItem, error) {
	dbInstance, err := p.SQLiteFactory.CreateInMemoryDB(caso, progression)
	if err != nil {
		return nil, nil, err
	}
	defer dbInstance.Close()

	upper := strings.ToUpper(strings.TrimSpace(query))
	var data interface{}
	var historyItem *models.SQLHistoryItem

	if strings.HasPrefix(upper, "SELECT") {
		rows, err := dbInstance.Query(query)
		if err != nil {
			return &models.GameResponse{Success: false, Error: err.Error(), State: p.getCurrentState(caso, progression)}, nil, nil
		}
		defer rows.Close()
		data = p.serializeRows(rows)
	} else {
		_, err := dbInstance.Exec(query)
		if err != nil {
			return &models.GameResponse{Success: false, Error: err.Error(), State: p.getCurrentState(caso, progression)}, nil, nil
		}
		historyItem = &models.SQLHistoryItem{
			Timestamp:   time.Now(),
			Query:       query,
			PuzzleState: progression.CurrentPuzzle,
			FocusState:  progression.CurrentFocus,
		}
	}

	validationResponse := p.runValidations(caso, progression, dbInstance, data)
	if validationResponse != nil {
		return validationResponse, historyItem, nil
	}

	return &models.GameResponse{
		Success: true,
		Data:    data,
		State:   p.getCurrentState(caso, progression),
	}, historyItem, nil
}

func (p *GameProcessor) runValidations(caso *models.Case, prog *models.Progression, dbInstance *sql.DB, lastData interface{}) *models.GameResponse {
	for _, v := range caso.Validations {
		if v.Puzzle == prog.CurrentPuzzle {
			var passed bool
			if v.Type == "result_check" || v.Type == "table_check" || v.Type == "table_complete_check" || v.Type == "sql_check" {
				var count interface{}
				err := dbInstance.QueryRow(v.CheckSQL).Scan(&count)
				if err == nil {
					if fmt.Sprintf("%v", count) == v.ExpectValue {
						passed = true
					}
				}
			}

			if passed {
				if v.UnlocksNext {
					prog.CurrentPuzzle = v.NextPuzzle
					prog.CurrentFocus = "none"
				}

				newState := p.getCurrentState(caso, prog)

				return &models.GameResponse{
					Success:   true,
					Narrative: v.SuccessNarrative,
					Data:      lastData,
					State:     newState,
				}
			}
		}
	}
	return nil
}

func (p *GameProcessor) serializeRows(rows *sql.Rows) []map[string]interface{} {
	cols, _ := rows.Columns()
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		rows.Scan(columnPointers...)
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		results = append(results, m)
	}
	return results
}

func (p *GameProcessor) getCurrentState(caso *models.Case, prog *models.Progression) models.GameState {
	state := models.GameState{
		CaseID:        prog.CaseID,
		CurrentPuzzle: prog.CurrentPuzzle,
		CurrentFocus:  prog.CurrentFocus,
	}
	for _, pz := range caso.Puzzles {
		if pz.Number == prog.CurrentPuzzle {
			state.Tables = pz.Tables
			state.Commands = pz.Commands
			state.Narrative = pz.Narrative
		}
	}
	return state
}
