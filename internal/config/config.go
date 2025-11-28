package config

type ServerModeType string

const (
	ServerModeProd ServerModeType = "prod"
	ServerModeDev  ServerModeType = "dev"
)

//go:generate go run github.com/ecordell/optgen -output zz_generated.configuration.go . Configuration
type Configuration struct {
	Server  Server         `debugmap:"visible"`
	Agent   Agent          `debugmap:"visible"`
	Auth    Authentication `debugmap:"visible"`
	Console Console        `debugmap:"visible"`

	// Log
	LogFormat string `debugmap:"visible"`
	LogLevel  string `debugmap:"visible"`
}

type Server struct {
	Mode          string `debugmap:"visible" default:"dev"`
	HTTPPort      int    `debugmap:"visible" default:"8080"`
	StaticsFolder string `debugmap:"visible"`
}

type Agent struct {
	Mode              string `debugmap:"visible" default:"disconnected"`
	ID                string `debugmap:"visible"`
	SourceID          string `debugmap:"visible"`
	NumWorkers        int    `debugmap:"visible" default:"3"`
	DataFolder        string `debugmap:"visible"`
	OpaPoliciesFolder string `debugmap:"visible"`
}

type Console struct {
	URL            string `debugmap:"visible" default:"localhost:7443"`
	UpdateInterval string `debugmap:"visible" default:"5s"`
}

type Authentication struct {
	Enabled     bool   `debugmap:"visible" default:"true"`
	JWTFilePath string `debugmap:"visible"`
}
