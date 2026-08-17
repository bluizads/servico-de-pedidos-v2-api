package repository

import (
	"context"
	"errors"
	"fmt"
	"servico-de-pedidos-v2-api/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PedidoRepository struct {
	//pool *pgxpool.Pool
	pool DB
}

func NovoPedidoRepository(pool DB) *PedidoRepository {
	return &PedidoRepository{pool: pool}
}

type PedidoRepositorioTransacao interface {
	ClienteExiste(contexto context.Context, clienteID string) (bool, error)
	InserirPedido(contexto context.Context, clienteID string) (model.Pedido, error)
	BuscarProduto(contexto context.Context, produtoID string) (model.Produto, error)
	InserirItem(contexto context.Context, pedidoID, produtoID string, precoNaCompra float64, quantidade int) (model.ItemPedido, error)
	AtualizarEstoqueComChecagem(contexto context.Context, produtoID string, quantidade int) (int64, error)
}

type pedidoRepositorioTx struct {
	db DB
}

func (repo *PedidoRepository) BuscarPorID(contexto context.Context, id string) (model.Pedido, error) {
	var pedido model.Pedido

	// busca o pedido
	err := repo.pool.QueryRow(contexto,
		`
		SELECT id, cliente_id, status, created_at
		FROM pedidos
		WHERE id = $1`,
		id,
	).Scan(&pedido.ID, &pedido.ClienteID, &pedido.Status, &pedido.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Pedido{}, model.ErrPedidoNaoEncontrado
	}

	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao buscar pedido: %w", err)
	}

	// busca itens
	itens, err := repo.buscarItens(contexto, pedido.ID)
	if err != nil {
		return model.Pedido{}, err
	}

	pedido.Itens = itens

	return pedido, nil
}

func (repo *PedidoRepository) buscarItens(contexto context.Context, pedidoID string) ([]model.ItemPedido, error) {
	linhas, err := repo.pool.Query(contexto,
		`
		SELECT id, pedido_id, produto_id, preco_na_compra, quantidade
		 FROM itens_pedido
		 WHERE pedido_id = $1`,
		pedidoID,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar itens: %w", err)
	}
	defer linhas.Close()

	itens := make([]model.ItemPedido, 0)
	for linhas.Next() {
		var item model.ItemPedido
		err := linhas.Scan(&item.ID, &item.PedidoID, &item.ProdutoID, &item.PrecoNaCompra, &item.Quantidade)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler item: %w", err)
		}
		itens = append(itens, item)
	}

	return itens, linhas.Err()
}

func (repo *PedidoRepository) Listar(contexto context.Context, limit int, offset int) ([]model.Pedido, error) {
	linhas, err := repo.pool.Query(contexto,
		`SELECT id, cliente_id, status, created_at
		FROM pedidos
		ORDER BY created_at
		LIMIT $1
		OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar pedidos: %w", err)
	}
	defer linhas.Close()

	pedidos := make([]model.Pedido, 0)
	for linhas.Next() {
		var pedido model.Pedido
		err := linhas.Scan(&pedido.ID, &pedido.ClienteID, &pedido.Status, &pedido.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler pedido: %w", err)
		}
		pedidos = append(pedidos, pedido)
	}
	if linhas.Err() != nil {
		return nil, fmt.Errorf("erro ao percorrer pedidos: %w", linhas.Err())
	}

	// busca os itens
	for i := range pedidos { //modifica mesmo
		itens, err := repo.buscarItens(contexto, pedidos[i].ID)
		if err != nil {
			return nil, err
		}
		pedidos[i].Itens = itens
	}

	return pedidos, nil
}

func (repo *PedidoRepository) Pagar(contexto context.Context, id string) (model.Pedido, error) {
	pedido, err := repo.BuscarPorID(contexto, id)
	if err != nil {
		return model.Pedido{}, err
	}

	if !pedido.PodeSerPago() {
		return model.Pedido{}, model.ErrMudancasStatusInvalida
	}

	_, err = repo.pool.Exec(contexto,
		`UPDATE pedidos
		SET status = $1
		WHERE id = $2`,
		model.StatusPago, id,
	)
	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao pagar pedido: %w", err)
	}

	pedido.Status = model.StatusPago
	return pedido, nil
}

func (repo *PedidoRepository) Cancelar(contexto context.Context, id string) (model.Pedido, error) {
	// busca pedido
	pedido, err := repo.BuscarPorID(contexto, id)
	if err != nil {
		return model.Pedido{}, err
	}

	// pode cancelar?
	if !pedido.PodeSerCancelado() {
		return model.Pedido{}, model.ErrMudancasStatusInvalida
	}

	// abre a transacao
	transacao, err := repo.pool.Begin(contexto)
	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao abrir transacao: %w", err)
	}
	defer transacao.Rollback(contexto)

	// mudando status
	_, err = transacao.Exec(contexto,
		`UPDATE pedidos
		SET status = $1
		WHERE id = $2
		`,
		model.StatusCancelado, id,
	)
	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao cancelar pedido: %w", err)
	}

	// devolve cada item
	for _, item := range pedido.Itens {
		_, err := transacao.Exec(contexto,
			`UPDATE produtos
			SET estoque = estoque + $1
			WHERE id = $2`,
			item.Quantidade, item.ProdutoID)

		if err != nil {
			return model.Pedido{}, fmt.Errorf("erro ao devolver estoque: %w", err)
		}
	}

	// confirma tudo
	err = transacao.Commit(contexto)
	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao confirmar transacao: %w", err)
	}

	pedido.Status = model.StatusCancelado
	return pedido, nil
}

func traduzirErroEstoque(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23514" {
		return model.ErrEstoqueInsuficiente
	}
	return err
}

func atualizarEstoqueComChecagem(contexto context.Context, db DB, produtoID string, quantidade int) (int64, error) {
	resultado, err := db.Exec(contexto,
		`UPDATE produtos SET estoque = estoque - $1 WHERE id = $2 AND estoque >= $1`,
		quantidade, produtoID,
	)
	err = traduzirErroEstoque(err)
	if err != nil {
		return 0, fmt.Errorf("erro ao atualizar estoque: %w", err)
	}

	return resultado.RowsAffected(), nil
}

func (repo *PedidoRepository) AtualizarEstoqueComChecagem(contexto context.Context, produtoID string, quantidade int) (int64, error) {
	return atualizarEstoqueComChecagem(contexto, repo.pool, produtoID, quantidade)
}

func (tx *pedidoRepositorioTx) ClienteExiste(contexto context.Context, clienteID string) (bool, error) {
	var existe bool
	err := tx.db.QueryRow(contexto,
		`
		SELECT EXISTS(
		SELECT 1 
		FROM clientes 
		WHERE id = $1
		)`,
		clienteID).Scan(&existe)

	if err != nil {
		return false, fmt.Errorf("erro ao verificar cliente: %w", err)
	}

	return existe, nil
}

func (tx *pedidoRepositorioTx) BuscarProduto(contexto context.Context, produtoID string) (model.Produto, error) {
	var produto model.Produto
	err := tx.db.QueryRow(contexto,
		`
		SELECT id, nome, preco, estoque 
		FROM produtos WHERE id = $1`,
		produtoID,
	).Scan(&produto.ID, &produto.Nome, &produto.Preco, &produto.Estoque)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Produto{}, model.ErrProdutoNaoEncontrado
	}

	if err != nil {
		return model.Produto{}, fmt.Errorf("erro ao buscar produto: %w", err)
	}

	return produto, nil
}

func (tx *pedidoRepositorioTx) InserirPedido(contexto context.Context, clienteID string) (model.Pedido, error) {
	var pedido model.Pedido
	err := tx.db.QueryRow(contexto,
		`
		INSERT INTO pedidos (cliente_id) 
		VALUES ($1)
		RETURNING id, cliente_id, status, created_at
		`,
		clienteID).Scan(&pedido.ID, &pedido.ClienteID, &pedido.Status, &pedido.CreatedAt)

	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao criar pedido: %w", err)
	}

	return pedido, nil
}

func (tx *pedidoRepositorioTx) InserirItem(contexto context.Context, pedidoID, produtoID string, precoNaCompra float64, quantidade int) (model.ItemPedido, error) {
	var item model.ItemPedido
	err := tx.db.QueryRow(contexto,
		`
		INSERT INTO itens_pedido (pedido_id, produto_id, preco_na_compra, quantidade)
		VALUES ($1, $2, $3, $4)
		RETURNING id, pedido_id, produto_id, preco_na_compra, quantidade`,
		pedidoID, produtoID, precoNaCompra, quantidade,
	).Scan(&item.ID, &item.PedidoID, &item.ProdutoID, &item.PrecoNaCompra, &item.Quantidade)

	if err != nil {
		return model.ItemPedido{}, fmt.Errorf("erro ao criar item: %w", err)
	}

	return item, nil
}

func (tx *pedidoRepositorioTx) AtualizarEstoqueComChecagem(contexto context.Context, produtoID string, quantidade int) (int64, error) {
	return atualizarEstoqueComChecagem(contexto, tx.db, produtoID, quantidade)
}

func (repo *PedidoRepository) ExecutarEmTransacao(contexto context.Context, fn func(tx PedidoRepositorioTransacao) error) error {
	transacao, err := repo.pool.Begin(contexto)
	if err != nil {
		return fmt.Errorf("erro ao abrir transacao: %w", err)
	}
	defer transacao.Rollback(contexto)

	tx := &pedidoRepositorioTx{db: transacao}

	err = fn(tx)
	if err != nil {
		return err
	}

	err = transacao.Commit(contexto)
	if err != nil {
		return fmt.Errorf("erro ao confirmar transacao: %w", err)
	}

	return nil
}
