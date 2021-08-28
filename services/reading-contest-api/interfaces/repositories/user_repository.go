package repositories

import (
	"github.com/tadoku/tadoku/services/reading-contest-api/domain"
	"github.com/tadoku/tadoku/services/reading-contest-api/interfaces/rdb"
	"github.com/tadoku/tadoku/services/reading-contest-api/usecases"
)

// NewUserRepository instantiates a new user repository
func NewUserRepository(sqlHandler rdb.SQLHandler) usecases.UserRepository {
	return &userRepository{sqlHandler: sqlHandler}
}

type userRepository struct {
	sqlHandler rdb.SQLHandler
}

// TODO: Refactor this for contest-api context user
func (r *userRepository) Store(user *domain.User) error {
	if user.ID == 0 {
		return r.create(user)
	}

	return r.update(user)
}

func (r *userRepository) create(user *domain.User) error {
	query := `
		insert into users
		(email, display_name, password, role, preferences)
		values ($1, $2, $3, $4, $5)
		returning id
	`

	row := r.sqlHandler.QueryRow(query, user.Email, user.DisplayName, "foobar", user.Role, user.Preferences)
	err := row.Scan(&user.ID)
	if err != nil {
		return domain.WrapError(err)
	}

	return nil
}

func (r *userRepository) update(user *domain.User) error {
	query := `
		update users
		set
			display_name = :display_name,
			preferences = :preferences
		where id = :id
	`
	_, err := r.sqlHandler.NamedExecute(query, user)
	return domain.WrapError(err)
}
