package users

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo"

	"go-chat/api"
)

type response struct {
	ID        uuid.UUID   `json:"id,omitempty"`
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
		return c.JSON(http.StatusBadRequest, err)
	}

	userID, err := h.UsersMgr.Create(createUserRequest)
	if err != nil {
		return c.JSON(err.HTTPStatusCode, err)
	}

	return c.JSON(http.StatusOK, response{
		ID: userID,
	})
}

// VerifyForLogin - verifies user credentials
func (h Handler) VerifyForLogin(c echo.Context) error {
	var validateUserRequest api.ValidateUserRequest
	if err := c.Bind(&validateUserRequest); err != nil {
		//return c.JSON(http.StatusBadRequest, err)
		return &api.APIError{HTTPStatusCode: http.StatusBadRequest, Cause: err}
	}

	b, err := h.UsersMgr.VerifyForLogin(validateUserRequest)
	if err != nil {
		//return c.JSON(http.StatusInternalServerError, err.Error())
		return c.JSON(err.HTTPStatusCode, err)
	}

	//SetCookie(c, b.FirstName)

	return c.JSON(http.StatusOK, b)
}

/*
func SetCookie(c echo.Context, value string) {
	cookie := new(http.Cookie)
	cookie.Name = "username"
	cookie.Value = value
	cookie.Expires = time.Now().Add(2 * time.Hour)
	c.SetCookie(cookie)
}
*/
