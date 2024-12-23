package store

import (
	"strings"
)

func sortOrder(val string) string {
	if lower := strings.ToLower(val); lower == "asc" || lower == "" {
		return "ASC"
	}

	return "DESC"
}
