package infra

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/srvc/fail"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// ErrEmptyUser when context contains no user when trying to get one
var ErrEmptyUser = fail.Errorf("user is empty")

type context struct {
	echo.Context
}

func (c context) User() (*domain.User, error) {
	claims := c.Claims()
	if claims.User != nil {
		return claims.User, nil
	}

	return nil, ErrEmptyUser
}

func (c context) Claims() *usecases.SessionClaims {
	if token, ok := c.Get("user").(*jwt.Token); ok {
		fmt.Println(token.Claims)
		return &token.Claims.(*jwtClaims).SessionClaims
	}
	return nil
}
