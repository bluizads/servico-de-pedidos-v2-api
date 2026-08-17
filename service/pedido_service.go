package service

import (
	"context"
	"servico-de-pedidos-v2-api/model"
	"servico-de-pedidos-v2-api/repository"
)

type PedidoRepositorio interface {
	BuscarPorID(ctx context.Context, id string) (model.Pedido, error)
	Listar(ctx context.Context, limit, offset int) ([]model.Pedido, error)
	ExecutarEmTransacao(ctx context.Context, fn func(tx repository.PedidoRepositorioTransacao) error) error
	Pagar(ctx context.Context, id string) (model.Pedido, error)
	Cancelar(ctx context.Context, id string) (model.Pedido, error)
}

type PedidoService struct {
	repo PedidoRepositorio
}

func NovoPedidoService(repo PedidoRepositorio) *PedidoService {
	return &PedidoService{repo: repo}
}

func (s *PedidoService) Criar(ctx context.Context, req model.CriarPedidoRequest) (model.Pedido, error) {
	var pedido model.Pedido

	err := s.repo.ExecutarEmTransacao(ctx, func(tx repository.PedidoRepositorioTransacao) error {
		clienteExiste, err := tx.ClienteExiste(ctx, req.ClienteID)
		if err != nil {
			return err
		}
		if !clienteExiste {
			return model.ErrClienteNaoEncontrado
		}

		pedido, err = tx.InserirPedido(ctx, req.ClienteID)
		if err != nil {
			return err
		}

		for _, itemReq := range req.Itens {
			if itemReq.Quantidade <= 0 {
				return model.ErrQuantidadeInvalida
			}

			produto, err := tx.BuscarProduto(ctx, itemReq.ProdutoID)
			if err != nil {
				return err
			}

			item, err := tx.InserirItem(ctx, pedido.ID, produto.ID, produto.Preco, itemReq.Quantidade)
			if err != nil {
				return err
			}

			linhasAfetadas, err := tx.AtualizarEstoqueComChecagem(ctx, produto.ID, itemReq.Quantidade)
			if err != nil {
				return err
			}
			if linhasAfetadas == 0 {
				return model.ErrEstoqueInsuficiente
			}

			pedido.Itens = append(pedido.Itens, item)
		}

		return nil
	})

	if err != nil {
		return model.Pedido{}, err
	}

	return pedido, nil
}

func (s *PedidoService) Listar(ctx context.Context, limit, offset int) ([]model.Pedido, error) {
	return s.repo.Listar(ctx, limit, offset)
}

func (s *PedidoService) BuscarPorID(ctx context.Context, id string) (model.Pedido, error) {
	return s.repo.BuscarPorID(ctx, id)
}

func (s *PedidoService) Pagar(ctx context.Context, id string) (model.Pedido, error) {
	return s.repo.Pagar(ctx, id)
}

func (s *PedidoService) Cancelar(ctx context.Context, id string) (model.Pedido, error) {
	return s.repo.Cancelar(ctx, id)
}
