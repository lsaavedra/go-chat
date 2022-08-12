package api

import (
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
