package cmd

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/lsaavedra/go-chat/configs"
	"github.com/lsaavedra/go-chat/db"
	"github.com/lsaavedra/go-chat/router"
	"github.com/lsaavedra/go-chat/users"
)

func main() {
	log.Print("starting benefits service \n")

	environment, err := configs.Environment{
		ServerHost: os.Getenv(configs.ServerHostKey),
		ServerPort: os.Getenv(configs.ServerPortKey),
		DbURL:      os.Getenv(configs.DbURLKey),
	}.Check()

	if err != nil {
		log.Fatal(err)
	}

	conn, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  environment.DbURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		log.Fatal(err)
	}

	initialMigration(conn)

	usersDB := db.NewUsersDB(conn)

	mgr := users.UsersMgr{
		UsersDB: usersDB,
	}

	handler := users.Handler{
		UsersMgr: mgr,
	}

	r := router.Router(handler)

	log.Print("successfully started go-chat service \n")
	r.Start(":" + environment.ServerPort)

}

func initialMigration(db *gorm.DB) {
	log.Print("starting db migration \n")
	err := db.AutoMigrate()

	if err != nil {
		log.Fatal("unable to run migrations")
	}
	log.Print("finished migrations \n")
}
