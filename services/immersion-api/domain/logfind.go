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
	// Check authorization before making DB call
	session := commondomain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}

	userID, err := uuid.Parse(session.Subject)
	if err != nil {
		return nil, ErrUnauthorized
	}

	req.IncludeDeleted = commondomain.IsRole(ctx, commondomain.RoleAdmin)

	log, err := s.repo.FindLogByID(ctx, req)
	if err != nil {
		return nil, err
	}

	// Needed to prevent leaking private registrations, only show to admins and the owner of the log
	isAdmin := commondomain.IsRole(ctx, commondomain.RoleAdmin)
	isOwner := log.UserID == userID
	if !isAdmin && !isOwner {
		log.Registrations = nil
	}

	return log, nil
}
