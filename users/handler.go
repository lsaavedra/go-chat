package users

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo"

	"go-chat/api"
)

type response struct {
	ID        *uuid.UUID  `json:"id,omitempty"`
	Message   string      `json:"message,omitempty"`
	RawObject interface{} `json:"raw,omitempty"`
}

type Handler struct {
	UsersMgr interface {
		Create(body api.CreateUserRequest) (uuid.UUID, *api.APIError)
		VerifyForLogin(body api.ValidateUserRequest) (api.ValidatedUserResponse, *api.APIError)
	}
}

// Create - creates a user
func (h Handler) Create(c echo.Context) error {
	var createUserRequest api.CreateUserRequest
	if err := c.Bind(&createUserRequest); err != nil {
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}
	if err := createUserRequest.Check(); err != nil {
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	userID, err := h.UsersMgr.Create(createUserRequest)
	if err != nil {
		return c.JSON(err.HTTPStatusCode, err)
	}

	return c.JSON(http.StatusOK, response{
		ID: &userID,
	})
}

// VerifyForLogin - verifies user credentials
func (h Handler) VerifyForLogin(c echo.Context) error {
	var validateUserRequest api.ValidateUserRequest
	if err := c.Bind(&validateUserRequest); err != nil {
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}
	if err := validateUserRequest.Check(); err != nil {
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	b, err := h.UsersMgr.VerifyForLogin(validateUserRequest)
	if err != nil {
		return c.JSON(err.HTTPStatusCode, err)
	}

	return c.JSON(http.StatusOK, b)
}
