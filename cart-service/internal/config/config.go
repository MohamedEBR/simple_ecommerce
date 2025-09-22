package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL	string
	Port		string
}

func Load() Config {
	_ = godotenv.Load("../.env")

	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		host := getenv("POSTGRES_HOST", "localhost")
		port := getenv("POSTGRES_PORT", "5432")
		user := getenv("APP_DB_USER", getenv("POSTGRES_USER", "postgres"))
		pass := getenv("APP_DB_PASSWORD", getenv("POSTGRES_PASSWORD", "postgres"))
		name := getenv("POSTGRES_DB", "simple_ecommerce")
		dbURL = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + name + "?sslmode=disable"
	}

	port:= os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("config: using DB=%s, PORT%s", redacted(dbURL), port)
	return Config{DatabaseURL: dbURL, Port: port}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func redacted(dsn string) string {
	at := strings.Index(dsn, "@")
	if at == -1 {
		return dsn
	}

	parts := strings.SplitN(dsn, "@", 2)
	return "****:@" + parts[1]
}
