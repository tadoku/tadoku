package repositories

import (
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/rdb"
	"github.com/tadoku/api/usecases"
)

func NewUserRepository(sqlHandler rdb.SQLHandler) usecases.UserRepository {
	return &userRepository{sqlHandler: sqlHandler}
}

type userRepository struct {
	sqlHandler rdb.SQLHandler
}

func (r *userRepository) Store(user domain.User) error {
	return nil
}

func (r *userRepository) FindByID(id int) domain.User {
	return nil
}
