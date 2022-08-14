package messages

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"go-chat/chatrooms"
	"go-chat/db"
)

type (
	messagesDB interface {
		Create(user db.User) (uuid.UUID, error)
	}
	MessagesMgr struct {
		MessagesDB *db.MessagesDB
		UsersDB    *db.UsersDB
	}
)

func NewMessagesMgr(messagesDB *db.MessagesDB, usersDB *db.UsersDB) *MessagesMgr {
	return &MessagesMgr{
		MessagesDB: messagesDB,
		UsersDB:    usersDB,
	}
}

func (m *MessagesMgr) SaveMsg(body chatrooms.ChatMessage) (uuid.UUID, error) {
	user, err := m.UsersDB.GetByNickName(body.Username)
	if err != nil {
		return uuid.Nil, err
	}
	message := db.Message{
		ID:       uuid.New(),
		UserID:   user.ID,
		Body:     body.Text,
		Chatroom: string(body.Room),
	}
	insertID, err := m.MessagesDB.Create(message)
	if err != nil {
		return uuid.Nil, err
	}

	log.Info().Msg("message save ok\n")
	return insertID, nil
}
