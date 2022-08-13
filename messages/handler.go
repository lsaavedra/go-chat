package messages

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
	MessagesMgr interface {
		Create(body api.CreateMessageRequest) (uuid.UUID, error)
	}
}

// Create - creates a user
func (h Handler) Create(c echo.Context) error {
	var createMessageRequest api.CreateMessageRequest
	if err := c.Bind(&createMessageRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	userID, err := h.MessagesMgr.Create(createMessageRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, response{
		ID:      userID,
		Message: "Message created successfully",
	})

	return nil
}
