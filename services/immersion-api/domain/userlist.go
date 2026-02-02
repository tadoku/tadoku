package domain

import (
	"context"
	"strings"

	"github.com/sahilm/fuzzy"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type UserListCache interface {
	GetUsers() []UserCacheEntry
}

type UserListRoleRepository interface {
	GetAllUserRoles(ctx context.Context) (map[string]string, error)
}

type UserCacheEntry struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
}

// userSearchSource implements fuzzy.Source for fuzzy matching on users
type userSearchSource struct {
	users []UserCacheEntry
}

func (s userSearchSource) String(i int) string {
	u := s.users[i]
	return strings.ToLower(u.DisplayName + " " + u.Email)
}

func (s userSearchSource) Len() int {
	return len(s.users)
}

type UserListRequest struct {
	PerPage int64
	Page    int64
	Query   string
}

type UserListResponse struct {
	Users     []UserListEntry
	TotalSize int
}

type UserListEntry struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
	Role        string // "user", "admin", or "banned" - empty string means "user"
}

type UserList struct {
	userCache UserListCache
	roleRepo  UserListRoleRepository
}

func NewUserList(userCache UserListCache, roleRepo UserListRoleRepository) *UserList {
	return &UserList{userCache: userCache, roleRepo: roleRepo}
}

func (s *UserList) Execute(ctx context.Context, req *UserListRequest) (*UserListResponse, error) {
	session := commondomain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	if session.Role != commondomain.RoleAdmin {
		return nil, ErrForbidden
	}

	perPage := int(req.PerPage)
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	page := int(req.Page)
	if page < 0 {
		page = 0
	}

	offset := page * perPage
	allUsers := s.userCache.GetUsers()

	var matchedUsers []UserCacheEntry
	var totalSize int

	if req.Query == "" {
		// No search - paginate full list
		totalSize = len(allUsers)
		start := offset
		if start >= totalSize {
			matchedUsers = []UserCacheEntry{}
		} else {
			end := start + perPage
			if end > totalSize {
				end = totalSize
			}
			matchedUsers = make([]UserCacheEntry, end-start)
			copy(matchedUsers, allUsers[start:end])
		}
	} else {
		// Fuzzy search
		source := userSearchSource{users: allUsers}
		matches := fuzzy.FindFrom(strings.ToLower(req.Query), source)
		totalSize = len(matches)

		start := offset
		if start >= totalSize {
			matchedUsers = []UserCacheEntry{}
		} else {
			end := start + perPage
			if end > totalSize {
				end = totalSize
			}
			matchedUsers = make([]UserCacheEntry, 0, end-start)
			for _, match := range matches[start:end] {
				matchedUsers = append(matchedUsers, allUsers[match.Index])
			}
		}
	}

	// Get all user roles from the database
	roleMap := make(map[string]string)
	if s.roleRepo != nil {
		var err error
		roleMap, err = s.roleRepo.GetAllUserRoles(ctx)
		if err != nil {
			// Log error but continue - roles will just be empty
			roleMap = make(map[string]string)
		}
	}

	users := make([]UserListEntry, 0, len(matchedUsers))
	for _, u := range matchedUsers {
		role := roleMap[u.ID] // Will be empty string if not found (means "user")
		users = append(users, UserListEntry{
			ID:          u.ID,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			CreatedAt:   u.CreatedAt,
			Role:        role,
		})
	}

	return &UserListResponse{
		Users:     users,
		TotalSize: totalSize,
	}, nil
}
