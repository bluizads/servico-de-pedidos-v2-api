package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"servico-de-pedidos-v2-api/model"
	//"servico-de-pedidos-v2-api/repository"
)

type ProdutoRepo interface {
	Criar(ctx context.Context, produto model.Produto) (model.Produto, error)
	Listar(ctx context.Context) ([]model.Produto, error)
	BuscarPorID(ctx context.Context, id string) (model.Produto, error)
}

type ProdutoController struct {
	//repo *repository.ProdutoRepository
	repo ProdutoRepo // trocando repositorio para interface
}

func NovoProdutoController(repo ProdutoRepo) *ProdutoController {
	return &ProdutoController{repo: repo}
}

// POST produtos
func (c *ProdutoController) Criar(w http.ResponseWriter, r *http.Request) {
	var produto model.Produto

	// decodifica JSON da requisicao
	err := json.NewDecoder(r.Body).Decode(&produto)
	if err != nil {
		responderErro(w, http.StatusBadRequest, "JSON invalido")
		return
	}

	// validando
	if produto.Nome == "" {
		responderErro(w, http.StatusBadRequest, "nome e obrigatorio")
		return
	}
	if produto.Preco < 0 || produto.Estoque < 0 {
		responderErro(w, http.StatusBadRequest, "preco e estoque nao podem ser negativos")
		return
	}

	criado, err := c.repo.Criar(r.Context(), produto)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, criado)
}

// GET produtos
func (c *ProdutoController) Listar(w http.ResponseWriter, r *http.Request) {
	produtos, err := c.repo.Listar(r.Context())
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, produtos)
}

// GET produtos id
func (c *ProdutoController) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	produto, err := c.repo.BuscarPorID(r.Context(), id)
	if errors.Is(err, model.ErrProdutoNaoEncontrado) {
		responderErro(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, produto)
}
