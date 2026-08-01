package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"servico-de-pedidos-v2-api/model"
)

type pedidoRepoFake struct {
	pedido         model.Pedido
	pedidos        []model.Pedido
	err            error
	criarChamado   bool
	limitRecebido  int
	offsetRecebido int
}

func (f *pedidoRepoFake) Criar(ctx context.Context, req model.CriarPedidoRequest) (model.Pedido, error) {
	f.criarChamado = true
	return f.pedido, f.err
}
func (f *pedidoRepoFake) Listar(ctx context.Context, limit int, offset int) ([]model.Pedido, error) {
	f.limitRecebido = limit
	f.offsetRecebido = offset
	return f.pedidos, f.err
}
func (f *pedidoRepoFake) BuscarPorID(ctx context.Context, id string) (model.Pedido, error) {
	return f.pedido, f.err
}
func (f *pedidoRepoFake) Pagar(ctx context.Context, id string) (model.Pedido, error) {
	return f.pedido, f.err
}
func (f *pedidoRepoFake) Cancelar(ctx context.Context, id string) (model.Pedido, error) {
	return f.pedido, f.err
}

func TestResponderErroPedido(t *testing.T) {
	casos := []struct {
		nome     string
		erro     error
		esperado int
	}{
		{"sucesso", nil, http.StatusOK},
		{"pedido nao encontrado", model.ErrPedidoNaoEncontrado, http.StatusNotFound},
		{"cliente nao encontrado", model.ErrClienteNaoEncontrado, http.StatusNotFound},
		{"produto nao encontrado", model.ErrProdutoNaoEncontrado, http.StatusNotFound},
		{"estoque insuficiente", model.ErrEstoqueInsuficiente, http.StatusConflict},
		{"status invalido", model.ErrMudancasStatusInvalida, http.StatusConflict},
		{"quantidade invalida", model.ErrQuantidadeInvalida, http.StatusBadRequest},
		{"pedido vazio", model.ErrPedidoVazio, http.StatusBadRequest},
		{"cliente invalido", model.ErrClienteInvalido, http.StatusBadRequest},
		{"erro embrulhado", fmt.Errorf("ao buscar: %w", model.ErrPedidoNaoEncontrado), http.StatusNotFound},
		{"erro desconhecido", errors.New("banco caiu"), http.StatusInternalServerError},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			controller := NovoPedidoController(&pedidoRepoFake{})
			gravar := httptest.NewRecorder()

			controller.responderErroPedido(gravar, caso.erro, model.Pedido{}, http.StatusOK)

			if gravar.Code != caso.esperado {
				t.Errorf("erro %v: status = %d, esperado %d", caso.erro, gravar.Code, caso.esperado)
			}
		})
	}
}

func TestLerInteiro(t *testing.T) {
	casos := []struct {
		nome     string
		url      string
		esperado int
	}{
		{"ausente vira padrao", "/pedidos", 10},
		{"valor valido", "/pedidos?limit=5", 5},
		{"nao numero vira padrao", "/pedidos?limit=abc", 10},
		{"negativo vira padrao", "/pedidos?limit=-1", 10},
		{"zero e valido", "/pedidos?limit=0", 0},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, caso.url, nil)
			if got := lerInteiro(req, "limit", 10); got != caso.esperado {
				t.Errorf("%s: lerInteiro = %d, esperado %d", caso.url, got, caso.esperado)
			}
		})
	}
}

func TestCriarPedido_JSONInvalido(t *testing.T) {
	fake := &pedidoRepoFake{}
	controller := NovoPedidoController(fake)

	req := httptest.NewRequest(http.MethodPost, "/pedidos", strings.NewReader("{"))
	gravar := httptest.NewRecorder()
	controller.Criar(gravar, req)

	if gravar.Code != http.StatusBadRequest {
		t.Errorf("status = %d, esperado: 400", gravar.Code)
	}
	if fake.criarChamado {
		t.Error("repositorio NAO deveria ter sido chamado com JSON invalido")
	}
}

func TestCriarPedido_SemClienteID(t *testing.T) {
	controller := NovoPedidoController(&pedidoRepoFake{})
	req := httptest.NewRequest(http.MethodPost, "/pedidos",
		strings.NewReader(`{"itens":[{"produtoId":"p1","quantidade":2}]}`))
	gravar := httptest.NewRecorder()
	controller.Criar(gravar, req)

	if gravar.Code != http.StatusBadRequest {
		t.Errorf("status = %d, esperado: 400", gravar.Code)
	}
}

func TestCriarPedido_SemItens(t *testing.T) {
	controller := NovoPedidoController(&pedidoRepoFake{})
	req := httptest.NewRequest(http.MethodPost, "/pedidos",
		strings.NewReader(`{"clienteId":"c1","itens":[]}`))
	gravar := httptest.NewRecorder()
	controller.Criar(gravar, req)

	if gravar.Code != http.StatusBadRequest {
		t.Errorf("status = %d, esperado: 400", gravar.Code)
	}
}

func TestCriarPedido_Sucesso(t *testing.T) {
	fake := &pedidoRepoFake{pedido: model.Pedido{ID: "1", ClienteID: "c1"}}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodPost, "/pedidos",
		strings.NewReader(`{"clienteId":"c1","itens":[{"produtoId":"p1","quantidade":2}]}`))
	gravar := httptest.NewRecorder()
	controller.Criar(gravar, req)

	if gravar.Code != http.StatusCreated {
		t.Errorf("status = %d, esperado: 201", gravar.Code)
	}
}

func TestListarPedidos_Sucesso(t *testing.T) {
	fake := &pedidoRepoFake{pedidos: []model.Pedido{{ID: "1"}}}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodGet, "/pedidos", nil)
	gravar := httptest.NewRecorder()
	controller.Listar(gravar, req)

	if gravar.Code != http.StatusOK {
		t.Errorf("status = %d, esperado: 200", gravar.Code)
	}
}

func TestListarPedidos_Erro(t *testing.T) {
	fake := &pedidoRepoFake{err: errors.New("falha")}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodGet, "/pedidos", nil)
	gravar := httptest.NewRecorder()
	controller.Listar(gravar, req)

	if gravar.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, esperado: 500", gravar.Code)
	}
}

func TestListarPedidos_UsaLimitEOffset(t *testing.T) {
	fake := &pedidoRepoFake{}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodGet, "/pedidos?limit=5&offset=20", nil)
	gravar := httptest.NewRecorder()
	controller.Listar(gravar, req)

	if fake.limitRecebido != 5 {
		t.Errorf("limit recebido = %d, esperado 5", fake.limitRecebido)
	}
	if fake.offsetRecebido != 20 {
		t.Errorf("offset recebido = %d, esperado 20", fake.offsetRecebido)
	}
}

func TestBuscarPedido_Sucesso(t *testing.T) {
	fake := &pedidoRepoFake{pedido: model.Pedido{ID: "1"}}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodGet, "/pedidos/1", nil)
	gravar := httptest.NewRecorder()
	controller.BuscarPorID(gravar, req)

	if gravar.Code != http.StatusOK {
		t.Errorf("status = %d, esperado: 200", gravar.Code)
	}
}

func TestBuscarPedido_NaoEncontrado(t *testing.T) {
	fake := &pedidoRepoFake{err: model.ErrPedidoNaoEncontrado}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodGet, "/pedidos/xyz", nil)
	gravar := httptest.NewRecorder()
	controller.BuscarPorID(gravar, req)

	if gravar.Code != http.StatusNotFound {
		t.Errorf("status = %d, esperado: 404", gravar.Code)
	}
}

func TestPagarPedido_Sucesso(t *testing.T) {
	fake := &pedidoRepoFake{pedido: model.Pedido{ID: "1"}}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodPost, "/pedidos/1/pagar", nil)
	gravar := httptest.NewRecorder()
	controller.Pagar(gravar, req)

	if gravar.Code != http.StatusOK {
		t.Errorf("status = %d, esperado: 200", gravar.Code)
	}
}

func TestPagarPedido_StatusInvalido(t *testing.T) {
	fake := &pedidoRepoFake{err: model.ErrMudancasStatusInvalida}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodPost, "/pedidos/1/pagar", nil)
	gravar := httptest.NewRecorder()
	controller.Pagar(gravar, req)

	if gravar.Code != http.StatusConflict {
		t.Errorf("status = %d, esperado: 409", gravar.Code)
	}
}

func TestCancelarPedido_Sucesso(t *testing.T) {
	fake := &pedidoRepoFake{pedido: model.Pedido{ID: "1"}}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodPost, "/pedidos/1/cancelar", nil)
	gravar := httptest.NewRecorder()
	controller.Cancelar(gravar, req)

	if gravar.Code != http.StatusOK {
		t.Errorf("status = %d, esperado: 200", gravar.Code)
	}
}

func TestCancelarPedido_NaoEncontrado(t *testing.T) {
	fake := &pedidoRepoFake{err: model.ErrPedidoNaoEncontrado}
	controller := NovoPedidoController(fake)
	req := httptest.NewRequest(http.MethodPost, "/pedidos/1/cancelar", nil)
	gravar := httptest.NewRecorder()
	controller.Cancelar(gravar, req)

	if gravar.Code != http.StatusNotFound {
		t.Errorf("status = %d, esperado: 404", gravar.Code)
	}
}
