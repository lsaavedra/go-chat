package env

// ServiceName is a const holding the service name.
const ServiceName = "partner-manager"

// structs holding service configuration.
type (
	Configuration struct {
		Hostname string `required:"true"`
		Env      string `required:"false"`
		Server   Server `required:"true"`
	}

	Server struct {
		Port string `required:"true"`
	}
)
