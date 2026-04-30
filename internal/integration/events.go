package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"casos-de-codigo-api/internal/models"
)

type EventPayload struct {
	TeamCode  string `json:"team_code"`
	Type      string `json:"type"`
	Matricula string `json:"matricula,omitempty"`
}

func sendEvent(route string, payload EventPayload) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", route, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("evento retornou status %d", resp.StatusCode)
	}

	return nil
}

func SendPuzzleEvent(
	tournament *models.Tournament,
	teamCode string,
	matricula string,
) error {
	return sendEvent(
		tournament.APIConfig.PuzzleEventRoute,
		EventPayload{
			TeamCode:  teamCode,
			Type:      "puzzle",
			Matricula: matricula,
		},
	)
}

func SendCaseEvent(
	tournament *models.Tournament,
	teamCode string,
) error {
	return sendEvent(
		tournament.APIConfig.CaseEventRoute,
		EventPayload{
			TeamCode: teamCode,
			Type:     "case",
		},
	)
}
