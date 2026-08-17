package main

import (
	"context"
	"log"
	"net/http"

	"servico-de-pedidos-v2-api/config"
	"servico-de-pedidos-v2-api/controllers"
	"servico-de-pedidos-v2-api/database"
	"servico-de-pedidos-v2-api/repository"
	"servico-de-pedidos-v2-api/routes"
	"servico-de-pedidos-v2-api/service"
)

func main() {
	// contexto ou base da API
	contexto := context.Background()

	// carregar do .env
	configuracao, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// abre a pool de conexao com o banco
	pool, err := database.Conectar(contexto, configuracao.DatabaseURL)
	if err != nil {
		log.Fatal(err) // sem banco, quebra
	}
	defer pool.Close() // só roda quando o main termina

	// sequencia: pool → repositórios → controllers → rotas → servidor

	// repositorios recebem o pool
	clienteRepo := repository.NovoClienteRepository(pool)
	produtoRepo := repository.NovoProdutoRepository(pool)
	pedidoRepo := repository.NovoPedidoRepository(pool)

	//adicionando o service
	produtoService := service.NovoProdutoService(produtoRepo)
	clienteService := service.NovoClienteService(clienteRepo)
	pedidoService := service.NovoPedidoService(pedidoRepo)

	// controller recebe os repositorios
	clienteController := controllers.NovoClienteController(clienteService)
	produtoController := controllers.NovoProdutoController(produtoService)
	pedidoController := controllers.NovoPedidoController(pedidoService)

	// rotas recebem os controllers
	router := routes.Configurar(clienteController, produtoController, pedidoController)

	// servidor
	endereco := ":" + configuracao.Port
	log.Println("servidor rodando em http://localhost" + endereco)

	err = http.ListenAndServe(endereco, router) // fica escutando pra sempre)
	if err != nil {
		log.Fatal(err)
	}
}
