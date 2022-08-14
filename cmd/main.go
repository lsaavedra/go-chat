package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-redis/redis"
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
	"go-chat/events"
	"go-chat/messages"
	"go-chat/router"
	"go-chat/users"
)

const serviceName = "go-chat"

func main() {
	log.Info().Msg(fmt.Sprintf("starting %s service \n", serviceName))

	env, err := configs.Environment{
		ServerHost: os.Getenv(configs.ServerHostKey),
		ServerPort: os.Getenv(configs.ServerPortKey),
		QueueUrl:   os.Getenv(configs.QueueUrl),
		CacheUrl:   os.Getenv(configs.CacheUrl),
		DbHost:     os.Getenv(configs.DbHost),
		DbPort:     os.Getenv(configs.DbPort),
		DbUser:     os.Getenv(configs.DbUser),
		DbPwd:      os.Getenv(configs.DbPwd),
		DbName:     os.Getenv(configs.DbName),
		DbSchema:   os.Getenv(configs.DbSchema),
	}.Check()
	if err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", env.DbHost, env.DbPort, env.DbUser, env.DbPwd, env.DbName)
	conn, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Error().Err(err)
	}

	queueCon, err := rabbit.Dial(env.QueueUrl)
	failOnError(err, "Failed to connect to RabbitMQ")
	queueClient := events.NewQueueClient(serviceName, queueCon)
	defer queueClient.CloseConnection()

	redisClient := redis.NewClient(&redis.Options{
		Addr: env.CacheUrl,
	})
	defer redisClient.Close()

	usersDB := db.NewUsersDB(conn)
	messagesDB := db.NewMessagesDB(conn)

	usersMgr := users.NewUsersMgr(usersDB)
	messagesMgr := messages.NewMessagesMgr(messagesDB, usersDB)

	usersHandler := users.Handler{
		UsersMgr: usersMgr,
	}

	botClient := bot.StocksClient{
		Getter: &http.Client{},
	}

	botMgr := bot.NewBotMgr(botClient, queueClient, nil)

	chatroomsHandler := chatrooms.Handler{
		BotManager:  botMgr,
		RedisClient: redisClient,
		Publisher:   queueClient,
	}

	msgProcessor := messages.NewProcessor(messagesMgr, botMgr, queueClient)

	apiHandlers := router.NewAPIHandlers(&usersHandler, &chatroomsHandler)
	r := router.Router(apiHandlers)

	go chatroomsHandler.HandleMessages()
	go msgProcessor.WaitForQueueMsgs()
	go chatroomsHandler.WaitingForQueueMsgs()

	log.Info().Msg(fmt.Sprintf("successfully started %s service \n", serviceName))
	r.Start(":" + env.ServerPort)

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panic().Err(err).Msg(msg)
	}
}
