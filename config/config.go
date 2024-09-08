package config

import (
	"log"

	"github.com/joho/godotenv"
)

func InitConfig() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	log.Print("Env loaded")
	return nil
}
