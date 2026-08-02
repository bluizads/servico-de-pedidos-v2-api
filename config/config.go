package config

import (
	"errors"
	"log" //registro de mensagens
	"os"  // biblioteca do sistema operacional

	"github.com/joho/godotenv"
)

// agrupar as config
type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (Config, error) {
	err := godotenv.Load() // Lê o arquivo .env e cola os dados q não poderiam ser expostos
	if err != nil {
		log.Println("AVISO: .env não encontrado, usando variáveis de ambiente do sistema")
	}

	databaseURL := os.Getenv("DATABASE_URL") // Lê os dados colados em .env"
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL nao configurada")
	}

	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080" // porta padrão
	}
	return Config{DatabaseURL: databaseURL, Port: porta}, nil
}
