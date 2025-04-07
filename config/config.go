package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	WindowManager string `json:"window_manager"`
	RedisAddr     string `json:"redis_addr"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		WindowManager: "bspwm",
		RedisAddr:     "localhost:6379",
	}
}

// LoadConfig loads configuration from a file
func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	log.Println("loading config.. ", homeDir)
	if err != nil {
		log.Println("Using default config!")
		return DefaultConfig(), nil
	}

	configPath := filepath.Join(homeDir, ".config", "startorswitch", "config.json")
	log.Println("config path: ", configPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("error reading config:  %s", err.Error())
		return DefaultConfig(), nil
	}
	log.Printf("loaded data:  %s", data)

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("error unmarshalling config:  %s", err.Error())
		return DefaultConfig(), nil
	}

	log.Printf("using config - redis: %s, wm:  %s", config.RedisAddr, config.WindowManager)

	return &config, nil
}
