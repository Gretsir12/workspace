package util

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type EnvKey string
type Port uint16

type Config struct {
	API_PORT          Port
	DATABASE_PORT     Port
	DATABASE_HOST     string
	DATABASE_USER     string
	DATABASE_PASSWORD string
	PROJECTS_PATH     string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	return &Config{
		API_PORT:          mustPort("API_PORT"),
		DATABASE_PORT:     mustPort("DATABASE_PORT"),
		DATABASE_HOST:     mustEnv("DATABASE_HOST"),
		DATABASE_USER:     mustEnv("DATABASE_USER"),
		DATABASE_PASSWORD: mustEnv("DATABASE_PASSWORD"),
		PROJECTS_PATH:     mustEnv("PROJECTS_PATH"),
	}
}

func mustEnv(key EnvKey) string {
	value := os.Getenv(string(key))

	if value == "" {
		log.Fatalf("%s is not set", key)
	}

	return value
}

func mustPort(key EnvKey) Port {
	value := mustEnv(key)

	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		log.Fatalf("invalid %s: %v", key, err)
	}

	return Port(port)
}
