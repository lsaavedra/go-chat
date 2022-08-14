package api

import (
	"errors"

	"github.com/google/uuid"
)

type (
	CreateUserRequest struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		NickName  string `json:"nick_name"`
		Password  string `json:"password"`
	}

	CreateUserResponse struct {
		ID uuid.UUID `json:"id"`
	}

	ValidateUserRequest struct {
		Email    string `json:"email" form:"userName" query:"userName"`
		Password string `json:"password" form:"password" query:"password"`
	}

	ValidatedUserResponse struct {
		ID        uuid.UUID `json:"id"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		Email     string    `json:"email"`
		NickName  string    `json:"nick_name"`
	}
)

func (c *CreateUserRequest) Check() error {
	switch {
	case c.FirstName == "":
		return errors.New("first_name is required")
	case c.LastName == "":
		return errors.New("last_name is required")
	case c.Email == "":
		return errors.New("email is required")
	case c.Password == "":
		return errors.New("password is required")
	}
	return nil
}

func (c *ValidateUserRequest) Check() error {
	switch {
	case c.Email == "":
		return errors.New("email is required")
	case c.Password == "":
		return errors.New("password is required")
	}
	return nil
}
