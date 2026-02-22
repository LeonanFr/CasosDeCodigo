package engine

import (
	"regexp"
)

var whereRegex = regexp.MustCompile(`(?is)(.*?\bwhere\b)(.*)`)

func NormalizeSQL(query string) string {
	matches := whereRegex.FindStringSubmatch(query)
	if len(matches) != 3 {
		return query // não tem WHERE → não mexe
	}

	head := matches[1]
	whereClause := matches[2]

	whereClause = normalizeWhere(whereClause)

	return head + whereClause
}

func normalizeWhere(where string) string {

	eq := regexp.MustCompile(`(?i)\b(\w+)\s*=\s*'([^']*)'`)
	where = eq.ReplaceAllString(where, "LOWER($1) = LOWER('$2')")

	neq := regexp.MustCompile(`(?i)\b(\w+)\s*!=\s*'([^']*)'`)
	where = neq.ReplaceAllString(where, "LOWER($1) != LOWER('$2')")

	like := regexp.MustCompile(`(?i)\b(\w+)\s+LIKE\s+'([^']*)'`)
	where = like.ReplaceAllString(where, "LOWER($1) LIKE LOWER('$2')")

	return where
}
