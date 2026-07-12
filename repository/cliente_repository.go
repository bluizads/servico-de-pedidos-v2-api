package repository

import (
	"context"
	"errors"
	"fmt"
	"servico-de-pedidos-v2-api/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type ClienteRepository struct {
	pool *pgxpool.Pool
}

func NovoClienteRepository(pool *pgxpool.Pool) *ClienteRepository {
	return &ClienteRepository{pool: pool}
}

func (repo *ClienteRepository) Criar(contexto context.Context, requisicao model.CriarClienteRequest) (model.Cliente, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(requisicao.Password), bcrypt.DefaultCost)
	// bcrypt.DefaultCost : quanto maior, mais lento pra quebrar
	if err != nil {
		return model.Cliente{}, fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}

	query :=
		`
		INSERT INTO clientes (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password_hash, created_at
		`

	var criado model.Cliente
	err = repo.pool.QueryRow(contexto, query, requisicao.Name, requisicao.Email, string(hash)).Scan(&criado.ID, &criado.Name, &criado.Email, &criado.PasswordHash, &criado.CreatedAt)

	// email duplicado: banco devolve 23505
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return model.Cliente{}, model.ErrEmailJaCadastrado
	}

	if err != nil {
		return model.Cliente{}, fmt.Errorf("erro ao criar cliente: %w", err)
	}

	return criado, nil
}

func (repo *ClienteRepository) BuscarPorID(contexto context.Context, id string) (model.Cliente, error) {
	query :=
		`
		SELECT id, name, email, password_hash, created_at
		FROM clientes
		WHERE id = $1
		`

	var cliente model.Cliente
	err := repo.pool.QueryRow(contexto, query, id).Scan(&cliente.ID, &cliente.Name, &cliente.Email, &cliente.PasswordHash, &cliente.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Cliente{}, model.ErrClienteNaoEncontrado
	}

	if err != nil {
		return model.Cliente{}, fmt.Errorf("erro ao buscar cliente: %w", err)
	}

	return cliente, nil
}

func (repo *ClienteRepository) Listar(contexto context.Context) ([]model.Cliente, error) {
	query :=
		`
		SELECT id, name, email, password_hash, created_at
		FROM clientes
		ORDER BY created_at
		`

	linhas, err := repo.pool.Query(contexto, query)
	if err != nil {
		return nil, fmt.Errorf("erros ao listar clientes: %w", err)
	}
	defer linhas.Close()

	lista := make([]model.Cliente, 0)

	for linhas.Next() {
		var cliente model.Cliente
		err := linhas.Scan(
			&cliente.ID, &cliente.Name, &cliente.Email, &cliente.PasswordHash, &cliente.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler cliente : %w", err)
		}

		lista = append(lista, cliente)
	}

	if linhas.Err() != nil {
		return nil, fmt.Errorf("erro ao percorrer a lista de clientes: %w", linhas.Err())
	}
	return lista, nil
}
