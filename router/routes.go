package router

import (
	"github.com/labstack/echo"

	"go-chat/chatrooms"
	"go-chat/users"
)

type (
	APIHandlers struct {
		UsersHandler     *users.Handler
		ChatroomsHandler *chatrooms.Handler
	}
)

func NewAPIHandlers(usersHandler *users.Handler, chatroomsHandler *chatrooms.Handler) *APIHandlers {
	return &APIHandlers{
		UsersHandler:     usersHandler,
		ChatroomsHandler: chatroomsHandler,
	}
}

func Router(h *APIHandlers) *echo.Echo {
	router := echo.New()

	// for the login
	router.File("/login", "public/login_chat.html")
	// for the chatroom websocket
	router.File("/chatrooms/:id", "public/chatroom.html")
	router.GET("/websocket/:id", h.ChatroomsHandler.HandleConnections)

	router.POST("/api/v1/users", h.UsersHandler.Create)
	router.POST("/api/v1/users/login", h.UsersHandler.VerifyForLogin)

	return router
}
