package users

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"go-chat/api"
	"go-chat/db"
)

const (
	userNotExistMsg      = "user not exists"
	userAlreadyExistsMsg = "user already exists"
	invalidCredentialMsg = "invalid credentials"
)

type (
	UsersDB interface {
		Create(user db.User) (uuid.UUID, error)
		GetByEmail(email string) (user db.User, err error)
	}
	UsersMgr struct {
		UsersDB *db.UsersDB
	}
)

func NewUsersMgr(usersDB *db.UsersDB) *UsersMgr {
	return &UsersMgr{
		UsersDB: usersDB,
	}
}

func (m *UsersMgr) Create(body api.CreateUserRequest) (uuid.UUID, *api.APIError) {
	dbUser, err := m.UsersDB.GetByEmail(body.Email)
	if err != nil {
		return uuid.Nil, &api.APIError{HTTPStatusCode: http.StatusInternalServerError, Cause: err}
	}
	if dbUser != (db.User{}) {
		return uuid.Nil, &api.APIError{HTTPStatusCode: http.StatusBadRequest, Msg: userAlreadyExistsMsg}
	}
	encodedPWD, err := encrypt(body.Password)
	if err != nil {
		return uuid.Nil, &api.APIError{HTTPStatusCode: http.StatusInternalServerError, Cause: err}
	}
	user := db.User{
		ID:        uuid.New(),
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Nickname:  body.NickName,
		Password:  encodedPWD,
		Email:     body.Email,
	}
	insertID, err := m.UsersDB.Create(user)
	if err != nil {
		//return uuid.Nil, err
		return uuid.Nil, &api.APIError{HTTPStatusCode: http.StatusInternalServerError, Cause: err}
	}

	log.Info().Msg("succesfully added new user \n")

	return insertID, nil
}

func (m *UsersMgr) VerifyForLogin(body api.ValidateUserRequest) (api.ValidatedUserResponse, *api.APIError) {
	dbUser, err := m.UsersDB.GetByEmail(body.Email)
	if err != nil {
		return api.ValidatedUserResponse{}, &api.APIError{HTTPStatusCode: http.StatusInternalServerError, Cause: err}
	}
	if dbUser == (db.User{}) {
		return api.ValidatedUserResponse{}, &api.APIError{HTTPStatusCode: http.StatusBadRequest, Msg: userNotExistMsg}
	}

	if !checkPwd(dbUser.Password, body.Password) {
		return api.ValidatedUserResponse{}, &api.APIError{HTTPStatusCode: http.StatusBadRequest, Msg: invalidCredentialMsg}
	}

	return api.ValidatedUserResponse{
		ID:        dbUser.ID,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		NickName:  dbUser.Nickname,
		Email:     dbUser.Email,
	}, nil
}

//Encrypt - hides sensible user data
func encrypt(data string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

//checkPwd - reveals sensible user data
func checkPwd(hash, plain string) bool {
	byteHash := []byte(hash)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plain))
	if err != nil {
		return false
	}

	return true
}
