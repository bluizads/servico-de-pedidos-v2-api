package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"servico-de-pedidos-v2-api/model"
	"servico-de-pedidos-v2-api/repository"
)

type PedidoController struct {
	repo *repository.PedidoRepository
}

func NovoPedidoController(repo *repository.PedidoRepository) *PedidoController {
	return &PedidoController{repo: repo}
}

// POST pedidos
func (c *PedidoController) Criar(w http.ResponseWriter, r *http.Request) {
	var req model.CriarPedidoRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		responderErro(w, http.StatusBadRequest, "JSON invalido")
		return
	}

	if req.ClienteID == "" {
		responderErro(w, http.StatusBadRequest, "clienteId e obrigatorio")
		return
	}
	if len(req.Itens) == 0 {
		responderErro(w, http.StatusBadRequest, "o pedido precisa ter pelo menos um item")
		return
	}

	pedido, err := c.repo.Criar(r.Context(), req)
	c.responderErroPedido(w, err, pedido, http.StatusCreated)
}

// GET pedidos (limit e offset)
func (c *PedidoController) Listar(w http.ResponseWriter, r *http.Request) {
	limit := lerInteiro(r, "limit", 10)
	offset := lerInteiro(r, "offset", 0)

	pedidos, err := c.repo.Listar(r.Context(), limit, offset)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, pedidos)
}

// GET pedidos id
func (c *PedidoController) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	pedido, err := c.repo.BuscarPorID(r.Context(), id)
	c.responderErroPedido(w, err, pedido, http.StatusOK)
}

// POST pedidos id -> pagar
func (c *PedidoController) Pagar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	pedido, err := c.repo.Pagar(r.Context(), id)
	c.responderErroPedido(w, err, pedido, http.StatusOK)
}

// POST /pedidos id -> cancelar
func (c *PedidoController) Cancelar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	pedido, err := c.repo.Cancelar(r.Context(), id)
	c.responderErroPedido(w, err, pedido, http.StatusOK)
}

// traducao dos erros
func (c *PedidoController) responderErroPedido(w http.ResponseWriter, err error, pedido model.Pedido, statusSucesso int) {
	switch {
	case err == nil:
		responderJSON(w, statusSucesso, pedido)
		// 200

	case errors.Is(err, model.ErrPedidoNaoEncontrado),
		errors.Is(err, model.ErrClienteNaoEncontrado),
		errors.Is(err, model.ErrProdutoNaoEncontrado):
		responderErro(w, http.StatusNotFound, err.Error())
		// 404

	case errors.Is(err, model.ErrEstoqueInsuficiente),
		errors.Is(err, model.ErrMudancasStatusInvalida):
		responderErro(w, http.StatusConflict, err.Error())
		//409

	case errors.Is(err, model.ErrQuantidadeInvalida),
		errors.Is(err, model.ErrPedidoVazio),
		errors.Is(err, model.ErrClienteInvalido):
		responderErro(w, http.StatusBadRequest, err.Error())
		// 400

	default:
		responderErro(w, http.StatusInternalServerError, err.Error())
		//500
	}
}

// ler da URL
func lerInteiro(r *http.Request, nome string, padrao int) int {
	valor := r.URL.Query().Get(nome)
	if valor == "" {
		return padrao
	}

	numero, err := strconv.Atoi(valor)
	if err != nil || numero < 0 {
		return padrao
	}

	return numero
}
