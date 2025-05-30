package auth

import (
	"errors"
	"fluxend/internal/domain/auth"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Auth struct {
	ctx echo.Context
}

func NewAuth(ctx echo.Context) *Auth {
	return &Auth{ctx: ctx}
}

func (a *Auth) User() (auth.User, error) {
	user := a.ctx.Get("user")

	// Ensure the value is a map as expected from JWT claims
	userData, ok := user.(auth.User)
	if !ok {
		return auth.User{}, errors.New("user.error.invalid_claim_structure")
	}

	return userData, nil
}

func (a *Auth) Uuid() (uuid.UUID, error) {
	user, err := a.User()
	if err != nil {
		return uuid.Nil, err
	}

	return user.Uuid, nil
}

func (a *Auth) RoleID() (int, error) {
	user, err := a.User()
	if err != nil {
		return 0, err
	}

	return user.RoleID, nil
}
