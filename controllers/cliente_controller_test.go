package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"servico-de-pedidos-v2-api/model"
	"strings"
	"testing"
)

// repositório sem banco de dados
type clienteRepoFake struct {
	cliente      model.Cliente
	err          error
	criarChamado bool
	clientes     []model.Cliente
}

func (f *clienteRepoFake) Criar(ctx context.Context, req model.CriarClienteRequest) (model.Cliente, error) {
	f.criarChamado = true
	return f.cliente, f.err
}

func (f *clienteRepoFake) Listar(ctx context.Context) ([]model.Cliente, error) {
	return f.clientes, f.err
}

func (f *clienteRepoFake) BuscarPorID(ctx context.Context, id string) (model.Cliente, error) {
	return f.cliente, f.err
}

func TestCriarCliente_JSONInvalido(t *testing.T) {
	fake := &clienteRepoFake{}
	novoCliente := NovoClienteController(fake)

	req := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader("{"))
	gravar := httptest.NewRecorder()
	novoCliente.Criar(gravar, req)

	if gravar.Code != http.StatusBadRequest {
		t.Errorf("status = %d, wanted: 400", gravar.Code)
	}
	if fake.criarChamado {
		t.Error("repositorio NAO deveria ter sido chamado com JSON invalido")
	}
}

func TestCriarCliente_CamposFaltando(t *testing.T) {
	fake := &clienteRepoFake{}
	novoCliente := NovoClienteController(fake)

	req := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(`{"name": "Ana"}`))
	gravar := httptest.NewRecorder()
	novoCliente.Criar(gravar, req)

	if gravar.Code != http.StatusBadRequest {
		t.Errorf("status = %d, wanted: 400", gravar.Code)
	}
}

func TestCriarCliente_EmailDuplicado(t *testing.T) {
	fake := &clienteRepoFake{err: model.ErrEmailJaCadastrado}
	novoCliente := NovoClienteController(fake)

	req := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(`{"name": "Ana", "email": "ana@ana.com", "password": "123"}`))
	gravar := httptest.NewRecorder()
	novoCliente.Criar(gravar, req)

	if gravar.Code != http.StatusConflict {
		t.Errorf("status = %d, wanted: 409", gravar.Code)
	}
}

func TestCriarCliente_Sucesso_NaoVazaHash(t *testing.T) {
	fake := &clienteRepoFake{cliente: model.Cliente{
		ID: "1", Name: "Ana", Email: "ana@ana.com", PasswordHash: "SEGREDO",
	}}
	novoCliente := NovoClienteController(fake)

	req := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(`{"name": "Ana", "email": "ana@ana.com", "password": "123"}`))
	gravar := httptest.NewRecorder()
	novoCliente.Criar(gravar, req)

	if gravar.Code != http.StatusCreated {
		t.Errorf("status = %d, wanted: 201", gravar.Code)
	}
	if strings.Contains(gravar.Body.String(), "SEGREDO") {
		t.Errorf("o hash da senha VAZOU no JSON!")
	}
}

func TestListarClientes_Sucesso(t *testing.T) {
	fake := &clienteRepoFake{clientes: []model.Cliente{{ID: "1"}, {ID: "2"}}}
	controller := NovoClienteController(fake)
	req := httptest.NewRequest(http.MethodGet, "/clientes", nil)
	gravar := httptest.NewRecorder()
	controller.Listar(gravar, req)

	if gravar.Code != http.StatusOK {
		t.Errorf("status = %d, esperado: 200", gravar.Code)
	}
}

func TestListarClientes_Erro(t *testing.T) {
	fake := &clienteRepoFake{err: errors.New("falha")}
	controller := NovoClienteController(fake)
	req := httptest.NewRequest(http.MethodGet, "/clientes", nil)
	gravar := httptest.NewRecorder()
	controller.Listar(gravar, req)

	if gravar.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, esperado: 500", gravar.Code)
	}
}

func TestBuscarCliente_NaoEncontrado(t *testing.T) {
	fake := &clienteRepoFake{err: model.ErrClienteNaoEncontrado}
	controller := NovoClienteController(fake)
	req := httptest.NewRequest(http.MethodGet, "/clientes/xyz", nil)
	gravar := httptest.NewRecorder()
	controller.BuscarPorID(gravar, req)

	if gravar.Code != http.StatusNotFound {
		t.Errorf("status = %d, esperado: 404", gravar.Code)
	}
}

func TestBuscarCliente_Erro(t *testing.T) {
	fake := &clienteRepoFake{err: errors.New("falha")}
	controller := NovoClienteController(fake)
	req := httptest.NewRequest(http.MethodGet, "/clientes/1", nil)
	gravar := httptest.NewRecorder()
	controller.BuscarPorID(gravar, req)

	if gravar.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, esperado: 500", gravar.Code)
	}
}

func TestBuscarCliente_Sucesso(t *testing.T) {
	fake := &clienteRepoFake{cliente: model.Cliente{ID: "1", PasswordHash: "SEGREDO"}}
	controller := NovoClienteController(fake)
	req := httptest.NewRequest(http.MethodGet, "/clientes/1", nil)
	gravar := httptest.NewRecorder()
	controller.BuscarPorID(gravar, req)

	if gravar.Code != http.StatusOK {
		t.Errorf("status = %d, esperado: 200", gravar.Code)
	}
	if strings.Contains(gravar.Body.String(), "SEGREDO") {
		t.Error("o hash da senha VAZOU no JSON!")
	}
}
