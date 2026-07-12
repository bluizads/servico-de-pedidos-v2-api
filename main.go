package main

import (
	"context"
	"fmt"
	"log"

	"servico-de-pedidos-v2-api/config"
	"servico-de-pedidos-v2-api/database"
)

func main() {
	// contexto ou base da API
	contexto := context.Background()

	// carregar do .env
	configuracao := config.Load()

	// abre a pool de conexao com o banco
	pool, err := database.Conectar(contexto, configuracao.DatabaseURL)
	if err != nil {
		log.Fatal(err) // sem banco, quebra
	}
	defer pool.Close() // só roda quando o main termina

	fmt.Println("Conectado ao banco de dados com sucesso!")
	fmt.Println("Porta: ", configuracao.Port)
}
