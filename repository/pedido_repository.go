package repository

import (
	"context"
	"errors"
	"fmt"
	"servico-de-pedidos-v2-api/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PedidoRepository struct {
	pool *pgxpool.Pool
}

func NovoPedidoRepository(pool *pgxpool.Pool) *PedidoRepository {
	return &PedidoRepository{pool: pool}
}

func (repo *PedidoRepository) Criar(contexto context.Context, req model.CriarPedidoRequest) (model.Pedido, error) {
	// abre a transacao
	transacao, err := repo.pool.Begin(contexto)
	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao abrir transacao: %w", err)
	}
	defer transacao.Rollback(contexto)

	// cliente existe?
	var clienteExiste bool

	err = transacao.QueryRow(contexto,
		`
		SELECT EXISTS(
		SELECT 1 
		FROM clientes 
		WHERE id = $1
		)`,
		req.ClienteID).Scan(&clienteExiste)

	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao verificar cliente: %w", err)
	}

	if !clienteExiste {
		return model.Pedido{}, model.ErrClienteNaoEncontrado
	}

	// criando o pedido
	var pedido model.Pedido

	err = transacao.QueryRow(contexto,
		`
		INSERT INTO pedidos (cliente_id) 
		VALUES ($1)
		RETURNING id, cliente_id, status, created_at
		`,
		req.ClienteID).Scan(&pedido.ID, &pedido.ClienteID, &pedido.Status, &pedido.CreatedAt)

	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao criar pedido: %w", err)
	}

	// cada item: busca o produto, valida estoque, insere item, reduz estoque
	for _, itemReq := range req.Itens {
		if itemReq.Quantidade <= 0 {
			return model.Pedido{}, model.ErrQuantidadeInvalida
		}

		// busca o produto
		var produto model.Produto
		err = transacao.QueryRow(contexto,
			`
			SELECT id, nome, preco, estoque 
			FROM produtos WHERE id = $1`,
			itemReq.ProdutoID,
		).Scan(&produto.ID, &produto.Nome, &produto.Preco, &produto.Estoque)

		if errors.Is(err, pgx.ErrNoRows) {
			return model.Pedido{}, model.ErrProdutoNaoEncontrado
		}

		if err != nil {
			return model.Pedido{}, fmt.Errorf("erro ao buscar produto: %w", err)
		}

		// valida estoque
		if !produto.TemEstoqueSuficiente(itemReq.Quantidade) {
			return model.Pedido{}, model.ErrEstoqueInsuficiente
		}

		// insere o item
		var item model.ItemPedido
		err = transacao.QueryRow(contexto,
			`
			INSERT INTO itens_pedido (pedido_id, produto_id, preco_na_compra, quantidade)
			VALUES ($1, $2, $3, $4)
			RETURNING id, pedido_id, produto_id, preco_na_compra, quantidade`,
			pedido.ID, produto.ID, produto.Preco, itemReq.Quantidade,
		).Scan(&item.ID, &item.PedidoID, &item.ProdutoID, &item.PrecoNaCompra, &item.Quantidade)

		if err != nil {
			return model.Pedido{}, fmt.Errorf("erro ao criar item: %w", err)
		}

		// reduz o estoque
		_, err = transacao.Exec(contexto,
			`
			UPDATE produtos
			SET estoque = estoque - $1
			WHERE id = $2`,
			itemReq.Quantidade, produto.ID,
		)

		if err != nil {
			return model.Pedido{}, fmt.Errorf("erro ao reduzir estoque: %w", err)
		}

		pedido.Itens = append(pedido.Itens, item)
	}

	// deu tudo certo
	err = transacao.Commit(contexto)

	if err != nil {
		return model.Pedido{}, fmt.Errorf("erro ao confirmar transacao: %w", err)
	}

	return pedido, nil

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
