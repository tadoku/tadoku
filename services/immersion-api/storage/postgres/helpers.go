package postgres

import (
	"database/sql"
	"time"
)

func NewQueries(psql *sql.DB) *Queries {
	return &Queries{psql}
}

func NewNullTime(val *time.Time) sql.NullTime {
	if val == nil {
		return sql.NullTime{
			Valid: false,
		}
	}

	return sql.NullTime{
		Valid: true,
		Time:  *val,
	}
}

func NewTimeFromNullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}

	return &t.Time
}

func NewNullString(val *string) sql.NullString {
	if val == nil || *val == "" {
		return sql.NullString{
			Valid: false,
		}
	}

	return sql.NullString{
		Valid:  true,
		String: *val,
	}
}

func NewStringFromNullString(val sql.NullString) *string {
	if !val.Valid {
		return nil
	}

	return &val.String
}

func NewNullInt32(val *int32) sql.NullInt32 {
	if val == nil {
		return sql.NullInt32{
			Valid: false,
		}
	}

	return sql.NullInt32{
		Valid: true,
		Int32: *val,
	}
}

func NewNullInt16(val *int16) sql.NullInt16 {
	if val == nil {
		return sql.NullInt16{
			Valid: false,
		}
	}

	return sql.NullInt16{
		Valid: true,
		Int16: *val,
	}
}

// TagsFromInterface converts an interface{} (from postgres array_agg) to []string
func TagsFromInterface(val interface{}) []string {
	if val == nil {
		return []string{}
	}

	switch v := val.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				result[i] = s
			}
		}
		return result
	case []byte:
		// Handle postgres array format: {tag1,tag2}
		return parsePostgresArray(string(v))
	case string:
		return parsePostgresArray(v)
	default:
		return []string{}
	}
}

func parsePostgresArray(s string) []string {
	if s == "" || s == "{}" {
		return []string{}
	}
	// Remove surrounding braces
	if len(s) >= 2 && s[0] == '{' && s[len(s)-1] == '}' {
		s = s[1 : len(s)-1]
	}
	if s == "" {
		return []string{}
	}
	// Simple split by comma - this handles basic cases
	// For more complex cases with escaped characters, a proper parser would be needed
	result := []string{}
	for _, part := range splitPostgresArray(s) {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitPostgresArray(s string) []string {
	var result []string
	var current string
	inQuotes := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuotes = !inQuotes
		} else if c == ',' && !inQuotes {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
