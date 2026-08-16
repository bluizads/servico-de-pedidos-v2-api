package service

import (
	"context"
	"errors"
	"servico-de-pedidos-v2-api/model"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type clienteRepoFake struct {
	nomeRecebido        string
	emailRecebido       string
	hashRecebido        string
	clienteParaDevolver model.Cliente
	erroParaDevolver    error
}

func (f *clienteRepoFake) Criar(ctx context.Context, nome, email, passwordHash string) (model.Cliente, error) {
	f.nomeRecebido = nome
	f.emailRecebido = email
	f.hashRecebido = passwordHash
	return f.clienteParaDevolver, f.erroParaDevolver
}

func (f *clienteRepoFake) Listar(ctx context.Context) ([]model.Cliente, error) { return nil, nil }
func (f *clienteRepoFake) BuscarPorID(ctx context.Context, id string) (model.Cliente, error) {
	return model.Cliente{}, nil
}

type hashDe string

func (h hashDe) Match(valor any) bool {
	texto, ok := valor.(string)
	if !ok {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(texto), []byte(h)) == nil
}

func TestClienteServiceCriar_SenhaVaiHasheada(t *testing.T) {
	repo := &clienteRepoFake{clienteParaDevolver: model.Cliente{ID: "1"}}
	s := NovoClienteService(repo)

	_, err := s.Criar(context.Background(), model.CriarClienteRequest{
		Name: "bruna", Email: "bruna@email.com", Password: "segredo123",
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !hashDe("segredo123").Match(repo.hashRecebido) {
		t.Error("hashRecebido nao confere com a senha original")
	}
}

func TestClienteServiceCriar_NomeEEmail(t *testing.T) {
	repo := &clienteRepoFake{clienteParaDevolver: model.Cliente{ID: "1"}}
	s := NovoClienteService(repo)

	_, err := s.Criar(context.Background(), model.CriarClienteRequest{
		Name: "bruna", Email: "bruna@email.com", Password: "segredo123",
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repo.nomeRecebido != "bruna" || repo.emailRecebido != "bruna@email.com" {
		t.Error("nome e/ou email nao confere com os originais")
	}
}

func TestClienteServiceCriar_ErroDoRepoPropaga(t *testing.T) {
	erroDoBanco := errors.New("falha no banco")
	repo := &clienteRepoFake{erroParaDevolver: erroDoBanco}
	s := NovoClienteService(repo)

	_, err := s.Criar(context.Background(), model.CriarClienteRequest{
		Name: "bruna", Email: "bruna@email.com", Password: "segredo123",
	})

	if !errors.Is(err, erroDoBanco) {
		t.Errorf("err = %v, wanted: %v", err, erroDoBanco)
	}
}
