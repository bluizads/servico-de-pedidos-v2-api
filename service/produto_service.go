package service

import (
	"context"
	"servico-de-pedidos-v2-api/model"
)

type ProdutoRepositorio interface {
	Criar(ctx context.Context, produto model.Produto) (model.Produto, error)
	Listar(ctx context.Context) ([]model.Produto, error)
	BuscarPorID(ctx context.Context, id string) (model.Produto, error)
}

type ProdutoService struct {
	repo ProdutoRepositorio
}

func NovoProdutoService(repo ProdutoRepositorio) *ProdutoService {
	return &ProdutoService{repo: repo}
}

func (s *ProdutoService) Criar(ctx context.Context, produto model.Produto) (model.Produto, error) {
	return s.repo.Criar(ctx, produto)
}

func (s *ProdutoService) Listar(ctx context.Context) ([]model.Produto, error) {
	return s.repo.Listar(ctx)
}

func (s *ProdutoService) BuscarPorID(ctx context.Context, id string) (model.Produto, error) {
	return s.repo.BuscarPorID(ctx, id)
}
