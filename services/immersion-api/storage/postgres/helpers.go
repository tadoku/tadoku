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
