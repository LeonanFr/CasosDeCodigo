package db

import (
	"casos-de-codigo-api/internal/models"
	"database/sql"
	"log"
	"strings"
	"unicode"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteFactory struct{}

var accentMap = map[rune]rune{
	'Á': 'A', 'À': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A',
	'á': 'A', 'à': 'A', 'â': 'A', 'ã': 'A', 'ä': 'A',
	'É': 'E', 'È': 'E', 'Ê': 'E', 'Ë': 'E',
	'é': 'E', 'è': 'E', 'ê': 'E', 'ë': 'E',
	'Í': 'I', 'Ì': 'I', 'Î': 'I', 'Ï': 'I',
	'í': 'I', 'ì': 'I', 'î': 'I', 'ï': 'I',
	'Ó': 'O', 'Ò': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O',
	'ó': 'O', 'ò': 'O', 'ô': 'O', 'õ': 'O', 'ö': 'O',
	'Ú': 'U', 'Ù': 'U', 'Û': 'U', 'Ü': 'U',
	'ú': 'U', 'ù': 'U', 'û': 'U', 'ü': 'U',
	'Ç': 'C', 'ç': 'C',
	'Ñ': 'N', 'ñ': 'N',
}

func normalizeString(s string) string {
	var result strings.Builder
	result.Grow(len(s))
	for _, r := range s {
		if newR, ok := accentMap[unicode.ToUpper(r)]; ok {
			result.WriteRune(newR)
		} else {
			result.WriteRune(unicode.ToUpper(r))
		}
	}
	return result.String()
}

func init() {
	sql.Register("sqlite3_normalize", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			return conn.RegisterFunc("NORMALIZE", normalizeString, true)
		},
	})
}

func NewSQLiteFactory() *SQLiteFactory {
	return &SQLiteFactory{}
}

func (f *SQLiteFactory) CreateInMemoryDB(caso *models.Case, progression *models.Progression) (*sql.DB, error) {

	db, err := sql.Open("sqlite3_normalize", ":memory:")
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
