package router

import (
	"github.com/labstack/echo"

	"go-chat/chatrooms"
	"go-chat/messages"
	"go-chat/users"
)

type (
	APIHandlers struct {
		UsersHandler     *users.Handler
		MessagesHandler  *messages.Handler
		ChatroomsHandler *chatrooms.Handler
	}
)

func NewAPIHandlers(usersHandler *users.Handler, messagesHandler *messages.Handler, chatroomsHandler *chatrooms.Handler) *APIHandlers {
	return &APIHandlers{
		UsersHandler:     usersHandler,
		MessagesHandler:  messagesHandler,
		ChatroomsHandler: chatroomsHandler,
	}
}

func Router(h *APIHandlers) *echo.Echo {
	router := echo.New()

	// for the login
	router.File("/login", "public/login_chat.html")
	// for the chatroom websocket
	router.File("/chatrooms/:id", "public/chatroom.html")
	//router.GET("/websocket", h.ChatroomsHandler.Hello)
	router.GET("/websocket", h.ChatroomsHandler.HandleConnections)

	router.POST("/api/v1/users", h.UsersHandler.Create)
	router.POST("/api/v1/users/login", h.UsersHandler.VerifyForLogin)
	router.POST("/api/v1/messages", h.MessagesHandler.Create)

	return router
}