package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

const (
	constraintUniqueContactValue = "unique_contact_value"
	constraintContactFKPartner   = "fk_contact_information_partner"
)

var (
	constraintToErr = map[string]error{
		constraintUniqueContactValue: ErrContactAlreadyExists,
		constraintContactFKPartner:   ErrPartnerNotExist,
	}

	ErrContactAlreadyExists = errors.New("contact already exists")
	ErrEmptyID              = errors.New("empty id")
	ErrIsPrimary            = errors.New("is primary")
	ErrPartnerNotExist      = errors.New("partner does not exist")
	ErrRecordNotFound       = gorm.ErrRecordNotFound
	ErrTypeMismatch         = errors.New("type does not match")
	ErrUniquePrimary        = errors.New("primary contact must be unique")
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

// TableName returns the table name associated to partnersDB.
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
	err := db.conn.WithContext(context.TODO()).Create(&user).Error //no olvidar el context.TODO()

	return user.ID, handleError(err)
}

func (db *UsersDB) GetByEmail(email string) (user User, err error) {
	err = db.conn.WithContext(context.TODO()).Where("email = ?", email).Find(&user).Error
	return
}

func handleError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if constraintErr := constraintToErr[pgErr.ConstraintName]; constraintErr != nil {
			return constraintErr
		}
	}

	return err
}
