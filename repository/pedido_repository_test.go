package repository

import (
	"context"
	"errors"
	"servico-de-pedidos-v2-api/model"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
)

func TestTraduzirErroEstoque(t *testing.T) {
	erroConstraint := &pgconn.PgError{Code: "23514", ConstraintName: "produtos_estoque_check"}

	resultado := traduzirErroEstoque(erroConstraint)

	if !errors.Is(resultado, model.ErrEstoqueInsuficiente) {
		t.Errorf("resultado = %v, wanted: %v", resultado, model.ErrEstoqueInsuficiente)
	}
}

func TestTraduzirErroEstoque_OutroErroPassaDireto(t *testing.T) {
	erroOriginal := errors.New("erro qualquer")

	resultado := traduzirErroEstoque(erroOriginal)

	if errors.Is(resultado, model.ErrEstoqueInsuficiente) {
		t.Errorf("resultado = %v, nao deveria ser ErrEstoqueInsuficiente", resultado)
	}
}

func TestAtualizarEstoqueComChecagem_Sucesso(t *testing.T) {
	mock := novoMock(t)
	repo := NovoPedidoRepository(mock)

	mock.ExpectExec("SET estoque = estoque - ").
		WithArgs(2, "1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	linhas, err := repo.AtualizarEstoqueComChecagem(context.Background(), "1", 2)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if linhas != 1 {
		t.Errorf("linhas = %d, wanted: 1", linhas)
	}
}

func TestAtualizarEstoqueComChecagem_EstoqueInsuficiente(t *testing.T) {
	mock := novoMock(t)
	repo := NovoPedidoRepository(mock)

	mock.ExpectExec("SET estoque = estoque - ").
		WithArgs(2, "1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	linhas, err := repo.AtualizarEstoqueComChecagem(context.Background(), "1", 2)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if linhas != 0 {
		t.Errorf("linhas = %d, wanted: 1", linhas)
	}
}

func TestAtualizarEstoqueComChecagem_DisparaRestricao(t *testing.T) {
	mock := novoMock(t)
	repo := NovoPedidoRepository(mock)

	mock.ExpectExec("SET estoque = estoque - ").
		WithArgs(2, "1").
		WillReturnError(&pgconn.PgError{Code: "23514"})

	_, err := repo.AtualizarEstoqueComChecagem(context.Background(), "1", 2)

	if !errors.Is(err, model.ErrEstoqueInsuficiente) {
		t.Errorf("err = %v, wanted: %v", err, model.ErrEstoqueInsuficiente)
	}

}

func TestPedidoRepositorioTx_ClienteExiste_Sucesso(t *testing.T) {
	mock := novoMock(t)
	tx := &pedidoRepositorioTx{db: mock}

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("1").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	existe, err := tx.ClienteExiste(context.Background(), "1")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !existe {
		t.Error("existe = false, wanted: true")
	}
}

func TestPedidoRepositorioTx_InserirPedido_Sucesso(t *testing.T) {
	mock := novoMock(t)
	tx := &pedidoRepositorioTx{db: mock}

	criadoEm := time.Now()
	mock.ExpectQuery("INSERT INTO pedidos").
		WithArgs("1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "cliente_id", "status", "created_at"}).
			AddRow("10", "1", model.StatusPendente, criadoEm))

	pedido, err := tx.InserirPedido(context.Background(), "1")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if pedido.ID != "10" {
		t.Errorf("pedido.ID = %q, wanted: %q", pedido.ID, "10")
	}
	if pedido.ClienteID != "1" {
		t.Errorf("pedido.ClienteID = %q, wanted: %q", pedido.ClienteID, "1")
	}
}

func TestPedidoRepositorioTx_BuscarProduto_Sucesso(t *testing.T) {
	mock := novoMock(t)
	tx := &pedidoRepositorioTx{db: mock}

	mock.ExpectQuery("SELECT id, nome, preco, estoque").
		WithArgs("1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "nome", "preco", "estoque"}).
			AddRow("1", "mouse", 99.99, 5))

	produto, err := tx.BuscarProduto(context.Background(), "1")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if produto.Nome != "mouse" {
		t.Errorf("produto.Nome = %q, wanted: %q", produto.Nome, "mouse")
	}
}

func TestPedidoRepositorioTx_BuscarProduto_NaoEncontrado(t *testing.T) {
	mock := novoMock(t)
	tx := &pedidoRepositorioTx{db: mock}

	mock.ExpectQuery("SELECT id, nome, preco, estoque").
		WithArgs("1").
		WillReturnError(pgx.ErrNoRows)

	_, err := tx.BuscarProduto(context.Background(), "1")

	if !errors.Is(err, model.ErrProdutoNaoEncontrado) {
		t.Errorf("err = %v, wanted: %v", err, model.ErrProdutoNaoEncontrado)
	}
}

func TestPedidoRepositorioTx_InserirItem_Sucesso(t *testing.T) {
	mock := novoMock(t)
	tx := &pedidoRepositorioTx{db: mock}

	mock.ExpectQuery("INSERT INTO itens_pedido").
		WithArgs("p1", "prod1", 10.5, 2).
		WillReturnRows(pgxmock.NewRows([]string{"id", "pedido_id", "produto_id", "preco_na_compra", "quantidade"}).
			AddRow("100", "p1", "prod1", 10.5, 2))

	item, err := tx.InserirItem(context.Background(), "p1", "prod1", 10.5, 2)

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if item.Quantidade != 2 {
		t.Errorf("item.Quantidade = %d, wanted: %d", item.Quantidade, 2)
	}
}

func TestExecutarEmTransacao_CommitEmSucesso(t *testing.T) {
	mock := novoMock(t)
	repo := NovoPedidoRepository(mock)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := repo.ExecutarEmTransacao(context.Background(), func(tx PedidoRepositorioTransacao) error {
		return nil
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
}

func TestExecutarEmTransacao_RollbackErro(t *testing.T) {
	mock := novoMock(t)
	repo := NovoPedidoRepository(mock)

	mock.ExpectBegin()
	mock.ExpectRollback()

	erroEsperado := errors.New("falhou")

	err := repo.ExecutarEmTransacao(context.Background(), func(tx PedidoRepositorioTransacao) error {
		return erroEsperado
	})

	if !errors.Is(err, erroEsperado) {
		t.Fatalf("err = %v, wanted: %v", err, erroEsperado)
	}
}
