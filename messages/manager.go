package messages

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"go-chat/api"
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
	log.Print("creating new user \n")
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

	log.Print("succesfully added new user \n")

	return insertID, nil
}
