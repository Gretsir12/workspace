package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APIPort    uint16
	DBHost     string
	DBPort     uint16
	DBUser     string
	DBPassword string
	DBName     string

	ProjectPath string
}

// Load читает .env (если он есть) и переменные окружения платформы.
func Load() (Config, error) {
	_ = godotenv.Load()

	var (
		c   Config
		err error
	)
	if c.APIPort, err = Port("API_PORT"); err != nil {
		return Config{}, err
	}
	if c.DBPort, err = Port("DATABASE_PORT"); err != nil {
		return Config{}, err
	}
	for _, f := range []struct {
		dst *string
		key string
	}{
		{&c.DBHost, "DATABASE_HOST"},
		{&c.DBUser, "DATABASE_USER"},
		{&c.DBPassword, "DATABASE_PASSWORD"},
		{&c.DBName, "DATABASE_DB"},
		{&c.ProjectPath, "PROJECT_PATH"},
	} {
		if *f.dst, err = Env(f.key); err != nil {
			return Config{}, err
		}
	}
	return c, nil
}

// Env возвращает обязательную переменную окружения.
func Env(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("%s is not set", key)
	}
	return v, nil
}

func Port(key string) (uint16, error) {
	v, err := Env(key)
	if err != nil {
		return 0, err
	}
	p, err := strconv.ParseUint(v, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return uint16(p), nil
}
