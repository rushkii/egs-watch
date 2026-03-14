package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	SessionString string
	TargetGroupId string
	PGHost        string
	PGPort        string
	PGUser        string
	PGPwd         string
	PGDb          string
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found, using default values\n%v\n", err)
	}

	SessionString = getEnv("SESSION_STRING", "file:sessions/bot.db?_foreign_keys=on")
	TargetGroupId = getEnv("TARGET_GROUP_ID", "") + "@g.us"

	PGHost = getEnv("PG_HOST", "localhost")
	PGPort = getEnv("PG_PORT", "5432")
	PGUser = getEnv("PG_USER", "postgres")
	PGPwd = getEnv("PG_PWD", "postgres")
	PGDb = getEnv("PG_DATABASE", "postgres")

}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
