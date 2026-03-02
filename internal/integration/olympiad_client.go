package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TeamResponse struct {
	Exists     bool         `json:"exists"`
	Integrates []TeamMember `json:"integrantes"`
}

type TeamMember struct {
	Matricula string `json:"matricula"`
	Nome      string `json:"nome"`
}

func FetchTeam(codeRoute string, code string) (*TeamResponse, error) {
	client := http.Client{
		Timeout: 60 * time.Second,
	}

	url := fmt.Sprintf("%s/%s", codeRoute, code)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("olimpíada retornou %d", resp.StatusCode)
	}

	var data TeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
