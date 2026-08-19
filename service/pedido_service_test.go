package service

import (
	"context"
	"errors"
	"servico-de-pedidos-v2-api/model"
	"servico-de-pedidos-v2-api/repository"
	"testing"
)

type pedidoRepositorioTxFake struct {
	clienteExiste     bool
	erroClienteExiste error

	pedidoInserido    model.Pedido
	erroInserirPedido error

	produtosPorID     map[string]model.Produto
	erroBuscarProduto error

	itemInserido    model.ItemPedido
	erroInserirItem error

	linhasAfetadasEstoque int64
	erroAtualizarEstoque  error
}

func (f *pedidoRepositorioTxFake) ClienteExiste(ctx context.Context, clienteID string) (bool, error) {
	return f.clienteExiste, f.erroClienteExiste
}

func (f *pedidoRepositorioTxFake) InserirPedido(ctx context.Context, clienteID string) (model.Pedido, error) {
	return f.pedidoInserido, f.erroInserirPedido
}

func (f *pedidoRepositorioTxFake) BuscarProduto(ctx context.Context, produtoID string) (model.Produto, error) {
	if f.erroBuscarProduto != nil {
		return model.Produto{}, f.erroBuscarProduto
	}
	return f.produtosPorID[produtoID], nil
}

func (f *pedidoRepositorioTxFake) InserirItem(ctx context.Context, pedidoID, produtoID string, precoNaCompra float64, quantidade int) (model.ItemPedido, error) {
	return f.itemInserido, f.erroInserirItem
}

func (f *pedidoRepositorioTxFake) AtualizarEstoqueComChecagem(ctx context.Context, produtoID string, quantidade int) (int64, error) {
	return f.linhasAfetadasEstoque, f.erroAtualizarEstoque
}

type pedidoRepoFake struct {
	tx *pedidoRepositorioTxFake
}

func (f *pedidoRepoFake) ExecutarEmTransacao(ctx context.Context, fn func(tx repository.PedidoRepositorioTransacao) error) error {
	return fn(f.tx)
}

func (f *pedidoRepoFake) BuscarPorID(ctx context.Context, id string) (model.Pedido, error) {
	return model.Pedido{}, nil
}
func (f *pedidoRepoFake) Listar(ctx context.Context, limit, offset int) ([]model.Pedido, error) {
	return nil, nil
}
func (f *pedidoRepoFake) Pagar(ctx context.Context, id string) (model.Pedido, error) {
	return model.Pedido{}, nil
}
func (f *pedidoRepoFake) Cancelar(ctx context.Context, id string) (model.Pedido, error) {
	return model.Pedido{}, nil
}

func TestPedidoServiceCriar_Sucesso(t *testing.T) {
	tx := &pedidoRepositorioTxFake{
		clienteExiste:  true,
		pedidoInserido: model.Pedido{ID: "1", ClienteID: "c1"},
		produtosPorID: map[string]model.Produto{
			"prod1": {ID: "prod1", Nome: "mouse", Preco: 99.99},
		},
		itemInserido:          model.ItemPedido{ID: "item1", Quantidade: 2},
		linhasAfetadasEstoque: 1,
	}
	repo := &pedidoRepoFake{tx: tx}
	s := NovoPedidoService(repo)

	pedido, err := s.Criar(context.Background(), model.CriarPedidoRequest{
		ClienteID: "c1",
		Itens: []model.ItemPedidoRequest{
			{ProdutoID: "prod1", Quantidade: 2},
		},
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if pedido.ID != "1" {
		t.Errorf("pedido.ID = %q, wanted: %q", pedido.ID, "1")
	}
	if len(pedido.Itens) != 1 {
		t.Fatalf("len(pedido.Itens) = %d, wanted: 1", len(pedido.Itens))
	}
}

func TestPedidoServiceCriar_ClienteNaoEncontrado(t *testing.T) {
	tx := &pedidoRepositorioTxFake{
		clienteExiste: false,
	}
	repo := &pedidoRepoFake{tx: tx}
	s := NovoPedidoService(repo)

	_, err := s.Criar(context.Background(), model.CriarPedidoRequest{
		ClienteID: "c1",
		Itens: []model.ItemPedidoRequest{
			{ProdutoID: "prod1", Quantidade: 2},
		},
	})

	if !errors.Is(err, model.ErrClienteNaoEncontrado) {
		t.Errorf("err = %v, wanted: %v", err, model.ErrClienteNaoEncontrado)
	}
}

func TestPedidoServiceCriar_QuantidadeInvalida(t *testing.T) {
	tx := &pedidoRepositorioTxFake{
		clienteExiste: true,
	}
	repo := &pedidoRepoFake{tx: tx}
	s := NovoPedidoService(repo)

	_, err := s.Criar(context.Background(), model.CriarPedidoRequest{
		ClienteID: "c1",
		Itens: []model.ItemPedidoRequest{
			{ProdutoID: "prod1", Quantidade: 0},
		},
	})

	if !errors.Is(err, model.ErrQuantidadeInvalida) {
		t.Errorf("err = %v, wanted: %v", err, model.ErrQuantidadeInvalida)
	}
}

func TestPedidoServiceCriar_ProdutoNaoEncontrado(t *testing.T) {
	tx := &pedidoRepositorioTxFake{
		clienteExiste:     true,
		erroBuscarProduto: model.ErrProdutoNaoEncontrado,
	}
	repo := &pedidoRepoFake{tx: tx}
	s := NovoPedidoService(repo)

	_, err := s.Criar(context.Background(), model.CriarPedidoRequest{
		ClienteID: "c1",
		Itens: []model.ItemPedidoRequest{
			{ProdutoID: "prod1", Quantidade: 2},
		},
	})

	if !errors.Is(err, model.ErrProdutoNaoEncontrado) {
		t.Errorf("err = %v, wanted: %v", err, model.ErrProdutoNaoEncontrado)
	}
}

func TestPedidoServiceCriar_EstoqueInsuficiente(t *testing.T) {
	tx := &pedidoRepositorioTxFake{
		clienteExiste:  true,
		pedidoInserido: model.Pedido{ID: "1", ClienteID: "c1"},
		produtosPorID: map[string]model.Produto{
			"prod1": {ID: "prod1", Nome: "mouse", Preco: 99.99},
		},
		itemInserido:          model.ItemPedido{ID: "item1", Quantidade: 2},
		linhasAfetadasEstoque: 0,
	}
	repo := &pedidoRepoFake{tx: tx}
	s := NovoPedidoService(repo)

	_, err := s.Criar(context.Background(), model.CriarPedidoRequest{
		ClienteID: "c1",
		Itens: []model.ItemPedidoRequest{
			{ProdutoID: "prod1", Quantidade: 3},
		},
	})

	if !errors.Is(err, model.ErrEstoqueInsuficiente) {
		t.Errorf("err = %v, wanted: %v", err, model.ErrEstoqueInsuficiente)
	}
}
