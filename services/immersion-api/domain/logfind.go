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
	session := commondomain.ParseUserIdentity(ctx)
	userID := uuid.Nil
	admin := false
	authenticated := session != nil && session.Subject != "guest"
	if authenticated {
		var err error
		userID, err = uuid.Parse(session.Subject)
		if err != nil {
			return nil, ErrUnauthorized
		}
		admin = isAdmin(ctx)
	}

	req.IncludeDeleted = admin

	log, err := s.repo.FindLogByID(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := hydrateLogActivity(log); err != nil {
		return nil, err
	}

	// Needed to prevent leaking private registrations, only show to admins and the owner of the log
	isOwner := authenticated && log.UserID == userID
	if !admin && !isOwner {
		log.Registrations = nil
	}

	return log, nil
}
