package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"

	"servico-de-pedidos-v2-api/model"
)

// colunasProduto sao as colunas que o repositorio espera nos SELECTs/RETURNING
var colunasProduto = []string{"id", "nome", "preco", "estoque", "created_at"}

// novoMock cria o pool falso e ja registra o Close + a checagem final
// de que todas as expectativas foram cumpridas
func novoMock(t *testing.T) pgxmock.PgxPoolIface {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("erro ao criar o mock do pool: %v", err)
	}

	t.Cleanup(func() {
		mock.Close()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas nao cumpridas: %v", err)
		}
	})

	return mock
}

func TestProdutoCriar_Sucesso(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	criadoEm := time.Now()
	mock.ExpectQuery("INSERT INTO produtos").
		WithArgs("mouse", 99.99, 5).
		WillReturnRows(pgxmock.NewRows(colunasProduto).
			AddRow("1", "mouse", 99.99, 5, criadoEm))

	produto, err := repo.Criar(context.Background(), model.Produto{
		Nome: "mouse", Preco: 99.99, Estoque: 5,
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if produto.ID != "1" {
		t.Errorf("ID = %q, wanted: %q", produto.ID, "1")
	}

	if produto.Nome != "mouse" {
		t.Errorf("Nome = %q, wanted: %q", produto.Nome, "mouse")
	}

	if produto.Preco != 99.99 {
		t.Errorf("Preco = %v, wanted: 99.99", produto.Preco)
	}

	if produto.Estoque != 5 {
		t.Errorf("Estoque = %d, wanted: 5", produto.Estoque)
	}
}

func TestProdutoCriar_ErroNoBanco(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	mock.ExpectQuery("INSERT INTO produtos").
		WithArgs("mouse", 99.99, 5).
		WillReturnError(errors.New("conexao caiu"))

	_, err := repo.Criar(context.Background(), model.Produto{
		Nome: "mouse", Preco: 99.99, Estoque: 5,
	})

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
}

func TestProdutoBuscarPorID_Sucesso(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	mock.ExpectQuery("SELECT id, nome, preco, estoque, created_at").
		WithArgs("1").
		WillReturnRows(pgxmock.NewRows(colunasProduto).
			AddRow("1", "mouse", 99.99, 5, time.Now()))

	produto, err := repo.BuscarPorID(context.Background(), "1")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if produto.Nome != "mouse" {
		t.Errorf("Nome = %q, wanted: %q", produto.Nome, "mouse")
	}
}

// pgx.ErrNoRows tem que virar o erro de dominio, nao um erro generico
func TestProdutoBuscarPorID_NaoEncontrado(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	mock.ExpectQuery("SELECT id, nome, preco, estoque, created_at").
		WithArgs("xyz").
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.BuscarPorID(context.Background(), "xyz")

	if !errors.Is(err, model.ErrProdutoNaoEncontrado) {
		t.Errorf("err = %v, wanted: %v", err, model.ErrProdutoNaoEncontrado)
	}
}

func TestProdutoBuscarPorID_ErroGenerico(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	mock.ExpectQuery("SELECT id, nome, preco, estoque, created_at").
		WithArgs("1").
		WillReturnError(errors.New("timeout"))

	_, err := repo.BuscarPorID(context.Background(), "1")

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}

	if errors.Is(err, model.ErrProdutoNaoEncontrado) {
		t.Error("erro generico NAO deveria virar ErrProdutoNaoEncontrado")
	}
}

func TestProdutoListar_Sucesso(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	mock.ExpectQuery("SELECT id, nome, preco, estoque, created_at").
		WillReturnRows(pgxmock.NewRows(colunasProduto).
			AddRow("1", "mouse", 99.99, 5, time.Now()).
			AddRow("2", "teclado", 159.99, 3, time.Now()))

	lista, err := repo.Listar(context.Background())

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(lista) != 2 {
		t.Fatalf("len(lista) = %d, wanted: 2", len(lista))
	}

	if lista[1].Nome != "teclado" {
		t.Errorf("lista[1].Nome = %q, wanted: %q", lista[1].Nome, "teclado")
	}
}

// lista vazia tem que ser slice vazio (nao nil), pra virar [] no JSON
func TestProdutoListar_Vazio(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	mock.ExpectQuery("SELECT id, nome, preco, estoque, created_at").
		WillReturnRows(pgxmock.NewRows(colunasProduto))

	lista, err := repo.Listar(context.Background())

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if lista == nil {
		t.Fatal("lista = nil, wanted: slice vazio")
	}

	if len(lista) != 0 {
		t.Errorf("len(lista) = %d, wanted: 0", len(lista))
	}
}

func TestProdutoListar_ErroNaQuery(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	mock.ExpectQuery("SELECT id, nome, preco, estoque, created_at").
		WillReturnError(errors.New("conexao caiu"))

	_, err := repo.Listar(context.Background())

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
}

// linha com menos colunas do que o Scan espera: falha ao ler a linha
func TestProdutoListar_ErroNoScan(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	mock.ExpectQuery("SELECT id, nome, preco, estoque, created_at").
		WillReturnRows(pgxmock.NewRows([]string{"id", "nome"}).
			AddRow("1", "mouse"))

	_, err := repo.Listar(context.Background())

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
}

// erro que so aparece depois de percorrer as linhas: linhas.Err()
func TestProdutoListar_ErroAoPercorrer(t *testing.T) {
	mock := novoMock(t)
	repo := NovoProdutoRepository(mock)

	mock.ExpectQuery("SELECT id, nome, preco, estoque, created_at").
		WillReturnRows(pgxmock.NewRows(colunasProduto).
			AddRow("1", "mouse", 99.99, 5, time.Now()).
			RowError(0, errors.New("conexao caiu no meio")))

	_, err := repo.Listar(context.Background())

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
}
