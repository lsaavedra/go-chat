package messages

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"go-chat/api"
	"go-chat/chatrooms"
	"go-chat/db"
)

type (
	messagesDB interface {
		Create(user db.User) (uuid.UUID, error)
	}
	MessagesMgr struct {
		MessagesDB *db.MessagesDB
	}
)

func NewMessagesMgr(messagesDB *db.MessagesDB) *MessagesMgr {
	return &MessagesMgr{
		MessagesDB: messagesDB,
	}
}

func (m *MessagesMgr) Create(body api.CreateMessageRequest) (uuid.UUID, error) {
	log.Print("creating new message \n")
	message := db.Message{
		ID:       uuid.New(),
		UserID:   body.UserID,
		Body:     body.Body,
		Chatroom: body.Chatroom,
	}
	insertID, err := m.MessagesDB.Create(message)
	if err != nil {
		return uuid.Nil, err
	}

	log.Print("succesfully added new message \n")

	return insertID, nil
}

func (m *MessagesMgr) CreateFromEvent(body chatrooms.ChatMessage) (uuid.UUID, error) {
	log.Print("creating new message \n")
	message := db.Message{
		ID:       uuid.New(),
		UserID:   uuid.Nil,
		Body:     body.Text,
		Chatroom: string(body.Room),
	}
	insertID, err := m.MessagesDB.Create(message)
	if err != nil {
		return uuid.Nil, err
	}

	log.Print("succesfully added new message \n")

	return insertID, nil
}
