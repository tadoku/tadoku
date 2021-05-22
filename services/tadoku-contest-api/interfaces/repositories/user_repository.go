package repositories

import (
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

	row := r.sqlHandler.QueryRow(query, user.Email, user.DisplayName, user.Password, user.Role, user.Preferences)
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

func (r *userRepository) UpdatePassword(user *domain.User) error {
	query := `
		update users
		set
			password = :password
		where id = :id
	`
	_, err := r.sqlHandler.NamedExecute(query, user)
	return domain.WrapError(err)
}

func (r *userRepository) FindByID(id uint64) (domain.User, error) {
	u := domain.User{}

	query := `
		select id, email, display_name, role, preferences
		from users
		where id = $1
	`
	err := r.sqlHandler.QueryRow(query, id).StructScan(&u)
	if err != nil {
		return u, domain.WrapError(err)
	}

	return u, nil
}

func (r *userRepository) FindByEmail(email string) (domain.User, error) {
	u := domain.User{}

	query := `
		select id, email, password, display_name, role, preferences
		from users
		where email = $1
	`
	err := r.sqlHandler.QueryRow(query, email).StructScan(&u)
	if err != nil {
		return u, domain.WrapError(err)
	}

	return u, nil
}
