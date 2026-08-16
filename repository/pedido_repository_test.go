package repository

import (
	"errors"
	"servico-de-pedidos-v2-api/model"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
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
