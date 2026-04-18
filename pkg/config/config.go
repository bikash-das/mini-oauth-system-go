package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// App config
type Config struct {
	ClientId           string
	ClientSecret       string
	RedirectURI        string
	AuthServerURL      string
	AuthServerTokenURL string
	Scope              string
}

func Load() *Config {
	if err := godotenv.Load(".env"); err == nil {
		log.Println(".env loaded")
	} else {
		p, _ := os.Getwd()
		log.Println("Current working directory: ", p)
		log.Println("No .env file, using system env")
	}
	return &Config{
		ClientId:           os.Getenv("CLIENT_ID"),
		ClientSecret:       os.Getenv("CLIENT_SECRET"),
		RedirectURI:        os.Getenv("REDIRECT_URI"),
		AuthServerURL:      os.Getenv("AUTH_SERVER_URL"),
		AuthServerTokenURL: os.Getenv("AUTH_SERVER_TOKEN_URL"),
		Scope:              os.Getenv("SCOPE"),
	}
}
