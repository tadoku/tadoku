package postgres

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func NullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

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

func NullableFloat4(value *float32) pgtype.Float4 {
	if value == nil {
		return pgtype.Float4{}
	}
	return pgtype.Float4{Float32: *value, Valid: true}
}

func UUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: value, Valid: true}
}

func NullableUUID(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *value, Valid: true}
}

func TextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func Int4Pointer(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

func Float4Pointer(value pgtype.Float4) *float32 {
	if !value.Valid {
		return nil
	}
	return &value.Float32
}

func UUIDPointer(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	id := uuid.UUID(value.Bytes)
	return &id
}

func UUIDs(values []pgtype.UUID) []uuid.UUID {
	result := make([]uuid.UUID, len(values))
	for i, value := range values {
		result[i] = value.Bytes
	}
	return result
}
