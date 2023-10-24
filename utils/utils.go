package utils

import (
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/log"
	"github.com/joho/godotenv"
)


func GoDotEnvVariable(key string) (string) {
	godotenv.Load(".env")
	return os.Getenv(key)
  }

  func GetEnvVarIntFromString(envVar string, fallback int) int {
	var envVarInt, err = strconv.Atoi(GoDotEnvVariable(envVar))
	if(err != nil){
		log.Error("Failed to read env var: ", envVar)
		return fallback
	}
	return envVarInt
}
