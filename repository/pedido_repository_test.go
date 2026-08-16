package repository

import (
	"context"
	"errors"
	"servico-de-pedidos-v2-api/model"
	"testing"

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
