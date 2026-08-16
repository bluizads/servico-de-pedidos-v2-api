package service

import (
	"context"
	"fmt"
	"servico-de-pedidos-v2-api/model"

	"golang.org/x/crypto/bcrypt"
)

type ClienteRepositorio interface {
	Criar(contexto context.Context, nome, email, passwordHash string) (model.Cliente, error)
	Listar(contexto context.Context) ([]model.Cliente, error)
	BuscarPorID(contexto context.Context, id string) (model.Cliente, error)
}

type ClienteService struct {
	repo ClienteRepositorio
}

func NovoClienteService(repo ClienteRepositorio) *ClienteService {
	return &ClienteService{repo: repo}
}

func (s *ClienteService) Criar(ctx context.Context, req model.CriarClienteRequest) (model.Cliente, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.Cliente{}, fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}

	return s.repo.Criar(ctx, req.Name, req.Email, string(hash))
}

func (s *ClienteService) Listar(ctx context.Context) ([]model.Cliente, error) {
	return s.repo.Listar(ctx)
}

func (s *ClienteService) BuscarPorID(ctx context.Context, id string) (model.Cliente, error) {
	return s.repo.BuscarPorID(ctx, id)
}
