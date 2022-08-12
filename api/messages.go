package api

import "github.com/google/uuid"

type (
	CreateMessageRequest struct {
		UserID    uuid.UUID `json:"user_id"`
		Body      string    `json:"body"`
		Chatroom  string    `json:"chatroom"`
		CreatedAt string    `json:"created_at"`
	}

	CreateMessageResponse struct {
		ID uuid.UUID `json:"id"`
	}
)
