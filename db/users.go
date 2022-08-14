package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()"`
	FirstName string
	LastName  string
	Nickname  string
	Password  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// TableName returns the table name associated to UsersDB.
func (*User) TableName() string {
	return "chatrooms.users"
}

type UsersDB struct {
	conn *gorm.DB
}

func NewUsersDB(conn *gorm.DB) *UsersDB {
	return &UsersDB{conn: conn}
}

func (db *UsersDB) Create(user User) (uuid.UUID, error) {
	err := db.conn.WithContext(context.TODO()).Create(&user).Error

	return user.ID, err
}

func (db *UsersDB) GetByEmail(email string) (user User, err error) {
	err = db.conn.WithContext(context.TODO()).Where("email = ?", email).Find(&user).Error
	return
}

func (db *UsersDB) GetByNickName(nickname string) (user User, err error) {
	err = db.conn.WithContext(context.TODO()).Where("nickname = ?", nickname).Find(&user).Error
	return
}
