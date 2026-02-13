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

// StringArrayFromInterface converts the any result from array_agg to []string.
// PostgreSQL returns array data as []byte in text format like "{foo,bar}" which we parse.
func StringArrayFromInterface(val any) []string {
	if val == nil {
		return []string{}
	}

	// pq driver returns []byte for text[] arrays
	bytes, ok := val.([]byte)
	if !ok || len(bytes) == 0 {
		return []string{}
	}

	s := string(bytes)

	// Empty array case: "{}"
	if s == "{}" {
		return []string{}
	}

	// Remove surrounding braces
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return []string{}
	}
	s = s[1 : len(s)-1]

	// Split by comma and return
	var result []string
	for _, part := range splitArrayElements(s) {
		result = append(result, part)
	}
	return result
}

// splitArrayElements handles PostgreSQL array text format, including quoted elements
func splitArrayElements(s string) []string {
	if s == "" {
		return nil
	}

	var result []string
	var current []byte
	inQuote := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '"' && !inQuote:
			inQuote = true
		case ch == '"' && inQuote:
			inQuote = false
		case ch == ',' && !inQuote:
			result = append(result, string(current))
			current = nil
		default:
			current = append(current, ch)
		}
	}
	result = append(result, string(current))
	return result
}
