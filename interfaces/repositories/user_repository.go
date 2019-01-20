package repositories

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/rdb"
	"github.com/tadoku/api/usecases"
)

// NewUserRepository instantiates a new user repository
func NewUserRepository(sqlHandler rdb.SQLHandler) usecases.UserRepository {
	return &userRepository{sqlHandler: sqlHandler}
}

type userRepository struct {
	sqlHandler rdb.SQLHandler
}

func (r *userRepository) Store(user domain.User) error {
	if user.ID == 0 {
		return r.create(user)
	}

	return r.update(user)
}

func (r *userRepository) create(user domain.User) error {
	query := `
		insert into users
		(email, display_name, password, role, preferences)
		values (:email, :display_name, :password, :role, :preferences)
	`
	_, err := r.sqlHandler.NamedExecute(query, user)
	return fail.Wrap(err)
}

func (r *userRepository) update(user domain.User) error {
	query := `
		update users
		set
			display_name = :display_name,
			preferences = :preferences
		where id = :id
	`
	_, err := r.sqlHandler.NamedExecute(query, user)
	return fail.Wrap(err)
}

func (r *userRepository) FindByID(id int) domain.User {
	return domain.User{}
}
