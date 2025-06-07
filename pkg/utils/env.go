package utils

import "github.com/joho/godotenv"

func LoadEnv() error {
	if err := godotenv.Load("config/.env"); err != nil {
		return err
	}
	return nil
}