package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    uuid.UUID
	Body      string
	Chatroom  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// TableName returns the table name associated to MessagesDB.
func (*Message) TableName() string {
	return "chatrooms.messages"
}

type MessagesDB struct {
	conn *gorm.DB
}

func NewMessagesDB(conn *gorm.DB) *MessagesDB {
	return &MessagesDB{conn: conn}
}

func (db *MessagesDB) Create(message Message) (uuid.UUID, error) {
	err := db.conn.WithContext(context.TODO()).Create(&message).Error

	return message.ID, err
}
