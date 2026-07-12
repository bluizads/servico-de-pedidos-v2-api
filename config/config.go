package config

import (
	"log" //registro de mensagens
	"os"  // biblioteca do sistema operacional

	"github.com/joho/godotenv"
)

// agrupar as config
type Config struct {
	DatabaseURL string
	Port        string
}

func Load() Config {
	err := godotenv.Load() // Lê o arquivo .env e cola os dados q não poderiam ser expostos
	if err != nil {
		log.Println("AVISO: .env não encontrado, usando")
	}

	databaseURL := os.Getenv("DATABASE_URL") // Lê os dados colados em .env"
	if databaseURL == "" {
		log.Fatal("DATABASE_URL nao configurada") // sem banco não pode continuar
	}

	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080" // porta padrão
	}
	return Config{DatabaseURL: databaseURL, Port: porta}
}
