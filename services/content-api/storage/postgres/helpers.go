package postgres

import (
	"database/sql"
	"time"
)

func NewNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{
			Valid: false,
		}
	}

	return sql.NullTime{
		Valid: true,
		Time:  *t,
	}
}

func NewTimeFromNullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}

	return &t.Time
}

func NewNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{
			Valid: false,
		}
	}

	return sql.NullString{
		Valid:  true,
		String: *s,
	}
}

func NewStringFromNullString(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}

	return &s.String
}
