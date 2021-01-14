package config

import (
	"fmt"
	"os"

	toml "github.com/BurntSushi/toml"
)

// default values for config
const (
	defaultIP      = "127.0.0.1"
	defaultPort    = "6000"
	defaultLogfile = "/Logs/go-chat-app.log"
)

// Server configuration struct
type Server struct {
	IP      string `toml:"ip"`
	Port    string
	Logfile string
}

// Configuration struct to hold the complete configuration
type Configuration struct {
	Title  string
	Server Server `toml:"server"`
}

// NewDefaultConfig initiates default config
func NewDefaultConfig() Configuration {
	return Configuration{
		Title: "DefaultConfiguration",
		Server: Server{
			IP:      defaultIP,
			Port:    defaultPort,
			Logfile: defaultLogfile,
		},
	}
}

// LoadConfig creates new config from the config file path provided in the request
func LoadConfig(configFile string) Configuration {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("No configuration file exists in the path provided" + configFile)
		return NewDefaultConfig()
	} else if err != nil {
		fmt.Println("Error in reading the file" + configFile)
		return NewDefaultConfig()
	}

	var configuration Configuration
	if _, err := toml.DecodeFile(configFile, &configuration); err != nil {
		fmt.Println("Error in reading the toml file" + configFile)
		return NewDefaultConfig()
	}

	return configuration
}
