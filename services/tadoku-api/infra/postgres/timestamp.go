package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func Timestamp(value time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: value.UTC(), Valid: true}
}

func Timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}
