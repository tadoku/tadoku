package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogFindRepository interface {
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
}

type LogFindRequest struct {
	ID             uuid.UUID
	IncludeDeleted bool
}

type LogFind struct {
	repo LogFindRepository
}

func NewLogFind(repo LogFindRepository) *LogFind {
	return &LogFind{repo: repo}
}

func (s *LogFind) Execute(ctx context.Context, req *LogFindRequest) (*Log, error) {
	req.IncludeDeleted = commondomain.IsRole(ctx, commondomain.RoleAdmin)

	log, err := s.repo.FindLogByID(ctx, req)
	if err != nil {
		return nil, err
	}

	session := commondomain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(session.Subject)

	// Needed to prevent leaking private registrations, only show to admins and the owner of the log
	if err != nil || !commondomain.IsRole(ctx, commondomain.RoleAdmin) && log.UserID != userID {
		log.Registrations = nil
	}

	return log, nil
}
