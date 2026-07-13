package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"servico-de-pedidos-v2-api/model"
	"servico-de-pedidos-v2-api/repository"
)

type ClienteController struct {
	repo *repository.ClienteRepository
}

func NovoClienteController(repo *repository.ClienteRepository) *ClienteController {
	return &ClienteController{repo: repo}
}

// POST clientes
func (c *ClienteController) Criar(w http.ResponseWriter, r *http.Request) {
	var req model.CriarClienteRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		responderErro(w, http.StatusBadRequest, "JSON invalido")
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		responderErro(w, http.StatusBadRequest, "name, email e password sao obrigatorios")
		return
	}

	criado, err := c.repo.Criar(r.Context(), req)

	// email duplicado -> 409
	if errors.Is(err, model.ErrEmailJaCadastrado) {
		responderErro(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, criado)
}

// GET clientes
func (c *ClienteController) Listar(w http.ResponseWriter, r *http.Request) {
	clientes, err := c.repo.Listar(r.Context())
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}
	responderJSON(w, http.StatusOK, clientes)
}

// GET clientes id
func (c *ClienteController) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cliente, err := c.repo.BuscarPorID(r.Context(), id)
	if errors.Is(err, model.ErrClienteNaoEncontrado) {
		responderErro(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, cliente)
}
