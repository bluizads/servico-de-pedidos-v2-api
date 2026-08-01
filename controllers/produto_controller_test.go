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

type produtoRepoFake struct {
	produto      model.Produto
	produtos     []model.Produto
	err          error
	criarChamado bool
}

func (f *produtoRepoFake) Criar(ctx context.Context, produto model.Produto) (model.Produto, error) {
	f.criarChamado = true
	return f.produto, f.err
}

func (f *produtoRepoFake) Listar(ctx context.Context) ([]model.Produto, error) {
	return f.produtos, f.err
}

func (f *produtoRepoFake) BuscarPorID(contexto context.Context, id string) (model.Produto, error) {
	return f.produto, f.err
}

func TestCriarProduto_JSONInvalido(t *testing.T) {
	fake := &produtoRepoFake{}
	controller := NovoProdutoController(fake)

	req := httptest.NewRequest(http.MethodPost, "/produtos", strings.NewReader("{"))
	gravar := httptest.NewRecorder()

	controller.Criar(gravar, req)

	if gravar.Code != http.StatusBadRequest {
		t.Errorf("status = %d, wanted: 400", gravar.Code)
	}

	if fake.criarChamado {
		t.Error("repositorio NAO deveria ter sido chamado com JSON invalido")
	}

}

func TestCriarProduto_NomeVazio(t *testing.T) {
	fake := &produtoRepoFake{}
	controller := NovoProdutoController(fake)

	req := httptest.NewRequest(http.MethodPost, "/produtos", strings.NewReader(`{"nome": ,"preco":10,"estoque":5}`))
	gravar := httptest.NewRecorder()

	controller.Criar(gravar, req)

	if gravar.Code != http.StatusBadRequest {
		t.Errorf("status = %d, wanted: 400", gravar.Code)
	}
}

func TestCriarProduto_ValoresNegativos(t *testing.T) {
	fake := &produtoRepoFake{}
	controller := NovoProdutoController(fake)

	req := httptest.NewRequest(http.MethodPost, "/produtos", strings.NewReader(`{"nome": numerosNegativos, "preco":-1,"estoque":-5}`))
	gravar := httptest.NewRecorder()

	controller.Criar(gravar, req)

	if gravar.Code != http.StatusBadRequest {
		t.Errorf("status = %d, wanted: 400", gravar.Code)
	}
}

func TestCriarProduto_ErroRepo(t *testing.T) {
	fake := &produtoRepoFake{err: errors.New("falha no banco")}
	controller := NovoProdutoController(fake)

	req := httptest.NewRequest(http.MethodPost, "/produtos", strings.NewReader(`{"nome": "mouse", "preco":99.99,"estoque":5}`))
	gravar := httptest.NewRecorder()

	controller.Criar(gravar, req)

	if gravar.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, wanted: 500", gravar.Code)
	}
}

func TestCriarProduto_Sucesso(t *testing.T) {
	fake := &produtoRepoFake{
		produto: model.Produto{
			ID: "1", Nome: "mouse", Preco: 99.99, Estoque: 5,
		}}
	controller := NovoProdutoController(fake)

	req := httptest.NewRequest(http.MethodPost, "/produtos", strings.NewReader(`{"nome": "mouse", "preco":99.99,"estoque":5}`))
	gravar := httptest.NewRecorder()

	controller.Criar(gravar, req)

	if gravar.Code != http.StatusCreated {
		t.Errorf("status = %d, wanted: 201", gravar.Code)
	}
}

func TestListarProdutos_Sucesso(t *testing.T) {
	fake := &produtoRepoFake{produtos: []model.Produto{
		{ID: "1", Nome: "mouse", Preco: 99.99, Estoque: 5},
		{ID: "2", Nome: "teclado", Preco: 159.99, Estoque: 3},
	}}
	controller := NovoProdutoController(fake)

	req := httptest.NewRequest(http.MethodGet, "/produtos", nil)
	gravar := httptest.NewRecorder()

	controller.Listar(gravar, req)

	if gravar.Code != http.StatusOK {
		t.Errorf("status = %d, wanted: 200", gravar.Code)
	}
}

func TestBuscarProduto_Sucesso(t *testing.T) {
	fake := &produtoRepoFake{
		produto: model.Produto{
			ID: "1", Nome: "mouse", Preco: 99.99, Estoque: 5,
		}}
	controller := NovoProdutoController(fake)

	req := httptest.NewRequest(http.MethodGet, "/produtos/1", nil)
	gravar := httptest.NewRecorder()

	controller.BuscarPorID(gravar, req)

	if gravar.Code != http.StatusOK {
		t.Errorf("status = %d, wanted: 200", gravar.Code)
	}
}

func TestBuscarProduto_NaoEncontrado(t *testing.T) {
	fake := &produtoRepoFake{err: model.ErrProdutoNaoEncontrado}
	controller := NovoProdutoController(fake)

	req := httptest.NewRequest(http.MethodGet, "/produtos/xyz", nil)
	gravar := httptest.NewRecorder()

	controller.BuscarPorID(gravar, req)

	if gravar.Code != http.StatusNotFound {
		t.Errorf("status = %d, wanted: 404", gravar.Code)
	}
}

func TestBuscarProduto_ErroGenerico(t *testing.T) {
	fake := &produtoRepoFake{err: errors.New("falha")}
	controller := NovoProdutoController(fake)

	req := httptest.NewRequest(http.MethodGet, "/produtos/1", nil)
	gravar := httptest.NewRecorder()

	controller.BuscarPorID(gravar, req)

	if gravar.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, wanted: 500", gravar.Code)
	}
}
