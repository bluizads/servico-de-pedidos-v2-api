package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"servico-de-pedidos-v2-api/model"
)

type ProdutoRepository struct {
	pool *pgxpool.Pool
}

func NovoProdutoRepository(pool *pgxpool.Pool) *ProdutoRepository {
	return &ProdutoRepository{pool: pool}
}

func (repo *ProdutoRepository) Criar(contexto context.Context, produto model.Produto) (model.Produto, error) {
	query :=
		`
		INSERT INTO produtos (nome, preco, estoque)
		VALUES ($1, $2, $3)
		RETURNING id, nome, preco, estoque, created_at
		`

	var criado model.Produto
	err := repo.pool.QueryRow(contexto, query, produto.Nome, produto.Preco, produto.Estoque).Scan(&criado.ID, &criado.Nome, &criado.Preco, &criado.Estoque, &criado.CreatedAt)

	if err != nil {
		return model.Produto{}, fmt.Errorf("erro ao criar produto: %w", err)
	}

	return criado, nil
}

func (repo *ProdutoRepository) BuscarPorID(contexto context.Context, id string) (model.Produto, error) {
	query :=
		`
		SELECT id, nome, preco, estoque, created_at
		FROM produtos
		WHERE id = $1
		`

	var produto model.Produto
	err := repo.pool.QueryRow(contexto, query, id).Scan(&produto.ID, &produto.Nome, &produto.Preco, &produto.Estoque, &produto.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Produto{}, model.ErrProdutoNaoEncontrado
	}

	if err != nil {
		return model.Produto{}, fmt.Errorf("erro ao buscar produto: %w", err)
	}

	return produto, nil
}

func (repo *ProdutoRepository) Listar(contexto context.Context) ([]model.Produto, error) {
	query :=
		`
		SELECT id, nome, preco, estoque, created_at
		FROM produtos
		ORDER BY created_at
		`

	linhas, err := repo.pool.Query(contexto, query)
	if err != nil {
		return nil, fmt.Errorf("erros ao listar produtos: %w", err)
	}
	defer linhas.Close()

	lista := make([]model.Produto, 0)

	for linhas.Next() {
		var produto model.Produto
		err := linhas.Scan(
			&produto.ID, &produto.Nome, &produto.Preco, &produto.Estoque, &produto.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler produto: %w", err)
		}

		lista = append(lista, produto)
	}

	if linhas.Err() != nil {
		return nil, fmt.Errorf("erro ao percorrer a lista de produtos: %w", linhas.Err())
	}
	return lista, nil
}

func (repo *ProdutoRepository) ReduzirEstoque(contexto context.Context, produtoID string, quantidade int) error {
	query :=
		`
		UPDATE produtos 
		SET estoque = estoque - $1 
		WHERE id = $2
		`

	_, err := repo.pool.Exec(contexto, query, quantidade, produtoID)
	if err != nil {
		return fmt.Errorf("erro reduzir estoque: %w", err)
	}
	return nil
}

func (repo *ProdutoRepository) DevolverEstoque(contexto context.Context, produtoID string, quantidade int) error {
	query :=
		`
		UPDATE produtos 
		SET estoque = estoque + $1 
		WHERE id = $2
		`

	_, err := repo.pool.Exec(contexto, query, quantidade, produtoID)
	if err != nil {
		return fmt.Errorf("erro devolver estoque: %w", err)
	}
	return nil
}
