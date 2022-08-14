package configs

import "errors"

// ServiceName is a const holding the service name.
const ServiceName = "partner-manager"

// Constants for environment setup
const (
	ServerHostKey = "SERVER_HOST"
	ServerPortKey = "SERVER_PORT"
	QueueUrl      = "QUEUE_URL"
	CacheUrl      = "CACHE_URL"
	DbHost        = "DB_HOST"
	DbPort        = "DB_PORT"
	DbUser        = "DB_USER"
	DbPwd         = "DB_PASSWORD"
	DbName        = "DB_NAME"
	DbSchema      = "DB_SCHEMA"
)

// Environment configurations struct
type Environment struct {
	ServerHost string
	ServerPort string
	QueueUrl   string
	CacheUrl   string
	DbHost     string
	DbPort     string
	DbUser     string
	DbPwd      string
	DbName     string
	DbSchema   string
}

// Check validates service configurations
func (e Environment) Check() (Environment, error) {
	switch {
	case e.ServerHost == "":
		return e, errMessage(ServerHostKey)
	case e.ServerPort == "":
		return e, errMessage(ServerPortKey)
	case e.QueueUrl == "":
		return e, errMessage(QueueUrl)
	case e.CacheUrl == "":
		return e, errMessage(CacheUrl)
	case e.DbHost == "":
		return e, errMessage(DbHost)
	case e.DbPort == "":
		return e, errMessage(DbPort)
	case e.DbUser == "":
		return e, errMessage(DbUser)
	case e.DbPwd == "":
		return e, errMessage(DbPwd)
	case e.DbName == "":
		return e, errMessage(DbName)
	case e.DbSchema == "":
		return e, errMessage(DbSchema)
	}
	return e, nil
}

func errMessage(key string) error {
	return errors.New("unable to get " + key + " configuration")
}
