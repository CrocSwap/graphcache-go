package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)


func GoDotEnvVariable(key string) (string) {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
	  log.Println("Error loading .env file: ", key)
	}
	return os.Getenv(key)
  }