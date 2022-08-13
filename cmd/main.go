package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go-chat/bot"
	"go-chat/chatrooms"
	"go-chat/configs"
	"go-chat/db"
	"go-chat/messages"
	"go-chat/router"
	"go-chat/users"
)

const (
	DefaultMigrationPath = "file://db/migrations"
	connectionError      = "Error setting up DB connection"
)

func main() {
	log.Print("starting go-chat service \n")

	environment, err := configs.Environment{
		ServerHost: os.Getenv(configs.ServerHostKey),
		ServerPort: os.Getenv(configs.ServerPortKey),
		DbURL:      os.Getenv(configs.DbURLKey),
	}.Check()

	if err != nil {
		log.Fatal().Err(err)
	}

	conn, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  environment.DbURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		log.Fatal().Err(err)
	}

	initialMigration(conn)
	//migrateSchemaWithPath(conn)

	queueCon, err := rabbit.Dial("amqp://guest:guest@localhost:5672/") // change to env vars
	failOnError(err, "Failed to connect to RabbitMQ")
	defer queueCon.Close()

	usersDB := db.NewUsersDB(conn)
	messagesDB := db.NewMessagesDB(conn)

	usersMgr := users.NewUsersMgr(usersDB)
	messagesMgr := messages.NewMessagesMgr(messagesDB)

	usersHandler := users.Handler{
		UsersMgr: usersMgr,
	}
	messagesHandler := messages.Handler{
		MessagesMgr: messagesMgr,
	}
	botClient := bot.StocksClient{
		Getter: &http.Client{},
	}
	botMgr := bot.NewBotMgr(botClient, queueCon)

	chatroomsHandler := chatrooms.Handler{
		BotManager: botMgr,
		QueueCon:   queueCon,
	}

	msgProcessor := messages.NewProcessor(messagesMgr, botMgr, queueCon)

	apiHandlers := router.NewAPIHandlers(&usersHandler, &messagesHandler, &chatroomsHandler)
	r := router.Router(apiHandlers)

	go chatroomsHandler.HandleMessages()
	go msgProcessor.ReadAndProcess()
	go chatroomsHandler.ReadAndProcess()

	log.Print("successfully started go-chat service \n")
	r.Start(":" + environment.ServerPort)

}

func initialMigration(conn *gorm.DB) {
	log.Print("starting db migration \n")
	err := conn.AutoMigrate(
		&db.User{},
		&db.Message{},
	)

	if err != nil {
		log.Fatal().Err(err).Msg("unable to run migrations")
	}
	log.Print("finished migrations \n")
}

// MigrateSchemaWithPath runs new upward data migrations.
func migrateSchemaWithPath(conn *gorm.DB) {
	user := "postgres"
	pwd := "postgres"
	host := "localhost"
	port := "7004"
	dbName := "chatrooms"
	forceMigrations := true

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=go-chat,public", user, pwd, host, port, dbName)

	migrations, err := migrate.New(DefaultMigrationPath, databaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to the database during migration")
	}

	if forceMigrations {
		if err := migrations.Force(-1); err != nil {
			log.Fatal().Err(err).Msg("error running migration force")
		}
	}

	if err = migrations.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Fatal().Err(err).Msg("error running migration up")
		}

		log.Error().Msgf("migration: %v", err)
	}

	defer func() {
		sourceErr, databaseErr := migrations.Close()
		if sourceErr != nil {
			log.Error().Err(sourceErr).Send()
			//log.Err(sourceErr).Send()
		}
		if databaseErr != nil {
			log.Error().Err(databaseErr).Send()
			//log.Err(databaseErr).Send()
		}
	}()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panic().Err(err).Msg(msg)
	}
}
