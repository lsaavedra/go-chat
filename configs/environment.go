package configs

import "errors"

// ServiceName is a const holding the service name.
const ServiceName = "partner-manager"

// Constants for environment setup
const (
	ServerHostKey = "SERVER_HOST"
	ServerPortKey = "SERVER_PORT"
	DbURLKey      = "POSTGRES_URL"
)

// Environment configurations struct
type Environment struct {
	ServerHost string
	ServerPort string
	DbURL      string
}

// Check validates service configurations
func (e Environment) Check() (Environment, error) {
	switch {
	case e.ServerHost == "":
		return e, errMessage(ServerHostKey)
	case e.ServerPort == "":
		return e, errMessage(ServerPortKey)
	case e.DbURL == "":
		return e, errMessage(DbURLKey)
	}

	return e, nil
}

func errMessage(key string) error {
	return errors.New("unable to get " + key + " configuration")
}
