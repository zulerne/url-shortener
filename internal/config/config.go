package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "production"
)

type Config struct {
	Env         string
	StoragePath string
	AliasLength int
	HttpConfig  HttpConfig
}

type HttpConfig struct {
	Address         string
	Timeout         time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	User            string
	Password        string
}

func MustLoad() *Config {
	cfg := &Config{
		Env:         fetchString("ENV", "local"),
		StoragePath: fetchStringRequired("STORAGE_PATH"),
		AliasLength: fetchInt("ALIAS_LENGTH", 6),
		HttpConfig: HttpConfig{
			Address:         fetchStringRequired("HTTP_ADDRESS"),
			Timeout:         fetchDuration("HTTP_TIMEOUT", 5*time.Second),
			IdleTimeout:     fetchDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: fetchDuration("HTTP_SHUTDOWN_TIMEOUT", 5*time.Second),
			User:            fetchString("HTTP_USER", ""),
			Password:        fetchString("HTTP_PASSWORD", ""),
		},
	}

	return cfg
}

func fetchString(key string, def string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return def
}

func fetchStringRequired(key string) string {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		log.Fatalf("%s is not set", key)
	}
	return val
}

func fetchDuration(key string, def time.Duration) time.Duration {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		return def
	}
	dur, err := time.ParseDuration(val)
	if err != nil {
		log.Fatalf("%s is not a valid duration", key)
	}
	return dur
}

func fetchInt(key string, def int) int {
	val, exists := os.LookupEnv(key)
	if !exists || val == "" {
		return def
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("%s is not a valid integer", key)
	}
	return i
}
