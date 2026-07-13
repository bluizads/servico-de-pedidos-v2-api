package routes

// como um maître de restaurante

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"servico-de-pedidos-v2-api/controllers"
)

func Configurar(
	clienteController *controllers.ClienteController,
	produtoController *controllers.ProdutoController,
	pedidoController *controllers.PedidoController,
) *chi.Mux {
	r := chi.NewRouter()

	// middlewares: tipo um segurança na porta
	r.Use(middleware.Logger)    // anota tudo
	r.Use(middleware.Recoverer) // recupera de panics

	// api tá viva?
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Route("/clientes", func(r chi.Router) {
		r.Post("/", clienteController.Criar)
		r.Get("/", clienteController.Listar)
		r.Get("/{id}", clienteController.BuscarPorID)
	})

	r.Route("/produtos", func(r chi.Router) {
		r.Post("/", produtoController.Criar)
		r.Get("/", produtoController.Listar)
		r.Get("/{id}", produtoController.BuscarPorID)
	})

	r.Route("/pedidos", func(r chi.Router) {
		r.Post("/", pedidoController.Criar)
		r.Get("/", pedidoController.Listar)
		r.Get("/{id}", pedidoController.BuscarPorID)
		r.Post("/{id}/pagar", pedidoController.Pagar)
		r.Post("/{id}/cancelar", pedidoController.Cancelar)
	})

	return r
}
