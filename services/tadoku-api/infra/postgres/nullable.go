package postgres

import "github.com/jackc/pgx/v5/pgtype"

func NullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

// NullableNonEmptyText maps nil and empty strings to SQL NULL.
func NullableNonEmptyText(value *string) pgtype.Text {
	if value == nil || *value == "" {
		return pgtype.Text{}
	}
	return NullableText(value)
}

func NullableInt4(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}
