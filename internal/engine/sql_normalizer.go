package engine

import "regexp"

func NormalizeSQL(query string) string {
	q := query

	eq := regexp.MustCompile(`(?i)(\w+)\s*=\s*'([^']*)'`)
	q = eq.ReplaceAllString(q, "LOWER($1) = LOWER('$2')")

	neq := regexp.MustCompile(`(?i)(\w+)\s*!=\s*'([^']*)'`)
	q = neq.ReplaceAllString(q, "LOWER($1) != LOWER('$2')")

	like := regexp.MustCompile(`(?i)(\w+)\s+LIKE\s+'([^']*)'`)
	q = like.ReplaceAllString(q, "LOWER($1) LIKE LOWER('$2')")

	return q
}
