package db

import (
	"casos-de-codigo-api/internal/models"
	"database/sql"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteFactory struct{}

func NewSQLiteFactory() *SQLiteFactory {
	return &SQLiteFactory{}
}

func (f *SQLiteFactory) CreateInMemoryDB(caso *models.Case, progression *models.Progression) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	for _, schema := range caso.Schemas {
		if schema.Puzzle <= progression.CurrentPuzzle {
			_, err = db.Exec(schema.CreateSQL)
			if err != nil {
				db.Close()
				return nil, err
			}

			if schema.InsertSQL != "" {
				_, err = db.Exec(schema.InsertSQL)
				if err != nil {
					db.Close()
					return nil, err
				}
			}
		}
	}

	for _, item := range progression.SQLHistory {
		if item.Query != "" && !f.isDangerousSQL(item.Query) {
			_, err = db.Exec(item.Query)
			if err != nil {
				log.Printf("Aviso: Falha ao reexecutar query do histórico: %v", err)
			}
		}
	}

	return db, nil
}

func (f *SQLiteFactory) isDangerousSQL(query string) bool {
	upperQuery := strings.ToUpper(query)
	dangerousKeywords := []string{"DROP", "TRUNCATE", "ALTER", "ATTACH", "DETACH", "VACUUM"}

	for _, keyword := range dangerousKeywords {
		if strings.Contains(upperQuery, keyword) {
			return true
		}
	}
	return false
}

func (f *SQLiteFactory) ValidateSQL(query string) bool {
	upperQuery := strings.ToUpper(query)

	if f.isDangerousSQL(query) {
		return false
	}

	allowedTables := []string{"pistas_logicas", "suspeitos", "livros", "funcionarios"}
	for _, table := range allowedTables {
		if strings.Contains(upperQuery, table) {
			return true
		}
	}

	return strings.HasPrefix(upperQuery, "SELECT") ||
		strings.HasPrefix(upperQuery, "INSERT") ||
		strings.HasPrefix(upperQuery, "UPDATE") ||
		strings.HasPrefix(upperQuery, "DELETE")
}
