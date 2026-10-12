package postgres

import (
	"context"
	"errors"
	"net"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func IsUniqueViolation(err error) bool {
	var pgError *pgconn.PgError
	return errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation
}

func classify(err error) error {
	if err == nil || errx.KindOf(err) != errx.Unknown {
		return err
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || !unavailable(err) {
		return err
	}
	return errx.NewUnavailableError("", err)
}

func unavailable(err error) bool {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case pgerrcode.AdminShutdown, pgerrcode.CrashShutdown, pgerrcode.CannotConnectNow:
			return true
		case pgerrcode.ProtocolViolation, pgerrcode.ConfigurationLimitExceeded:
			return false
		}
		return pgerrcode.IsConnectionException(pgError.Code) || pgerrcode.IsInsufficientResources(pgError.Code)
	}

	var connectError *pgconn.ConnectError
	var netError net.Error
	return errors.As(err, &connectError) && errors.As(connectError, &netError)
}
