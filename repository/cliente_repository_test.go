package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"

	"servico-de-pedidos-v2-api/model"
)

// colunasCliente sao as colunas que o repositorio espera nos SELECTs/RETURNING
var colunasCliente = []string{"id", "name", "email", "password_hash", "created_at"}

// hashDe casa com qualquer string que seja o hash bcrypt da senha esperada.
// Precisa ser um matcher (e nao um valor no WithArgs) porque o bcrypt sorteia
// um salt novo a cada chamada: o hash exato e imprevisivel, so da pra conferir
// por comparacao.
//type hashDe string

//func (h hashDe) Match(valor any) bool {
//	texto, ok := valor.(string)
//	if !ok {
//		return false
//	}
//	return bcrypt.CompareHashAndPassword([]byte(texto), []byte(h)) == nil
//}

func TestClienteCriar_Sucesso(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	criadoEm := time.Now()
	mock.ExpectQuery("INSERT INTO clientes").
		WithArgs("bruna", "bruna@email.com", pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows(colunasCliente).
			AddRow("1", "bruna", "bruna@email.com", "$2a$10$hashvindodobanco", criadoEm))

	cliente, err := repo.Criar(context.Background(), "bruna", "bruna@email.com", "segredo123")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if cliente.ID != "1" {
		t.Errorf("ID = %q, wanted: %q", cliente.ID, "1")
	}

	if cliente.Name != "bruna" {
		t.Errorf("Name = %q, wanted: %q", cliente.Name, "bruna")
	}

	if cliente.Email != "bruna@email.com" {
		t.Errorf("Email = %q, wanted: %q", cliente.Email, "bruna@email.com")
	}
}

// a senha NUNCA pode chegar crua no banco: o 3o argumento tem que ser
// um hash bcrypt que confere com a senha enviada
//func TestClienteCriar_SenhaVaiHasheada(t *testing.T) {
//	mock := novoMock(t)
//	repo := NovoClienteRepository(mock)

//	mock.ExpectQuery("INSERT INTO clientes").
//		WithArgs("bruna", "bruna@email.com", hashDe("segredo123")).
//		WillReturnRows(pgxmock.NewRows(colunasCliente).
//			AddRow("1", "bruna", "bruna@email.com", "$2a$10$hashvindodobanco", time.Now()))

//	_, err := repo.Criar(context.Background(), model.CriarClienteRequest{
//		Name: "bruna", Email: "bruna@email.com", Password: "segredo123",
//	})

//	if err != nil {
//		t.Fatalf("erro inesperado: %v", err)
//	}
//}

// 23505 = unique_violation do Postgres. O repositorio tem que traduzir
// esse codigo pro erro de dominio, senao a camada de cima nao sabe
// que o problema foi email repetido
func TestClienteCriar_EmailDuplicado(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("INSERT INTO clientes").
		WithArgs("bruna", "bruna@email.com", pgxmock.AnyArg()).
		WillReturnError(&pgconn.PgError{
			Code:           "23505",
			Message:        "duplicate key value violates unique constraint",
			ConstraintName: "clientes_email_key",
		})

	_, err := repo.Criar(context.Background(), "bruna", "bruna@email.com", "segredo123")

	if !errors.Is(err, model.ErrEmailJaCadastrado) {
		t.Errorf("err = %v, wanted: %v", err, model.ErrEmailJaCadastrado)
	}
}

// outro erro do Postgres (23503 = foreign_key_violation) nao pode
// ser confundido com email duplicado
func TestClienteCriar_OutroErroDoPostgres(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("INSERT INTO clientes").
		WithArgs("bruna", "bruna@email.com", pgxmock.AnyArg()).
		WillReturnError(&pgconn.PgError{Code: "23503", Message: "foreign key violation"})

	_, err := repo.Criar(context.Background(), "bruna", "bruna@email.com", "segredo123")

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}

	if errors.Is(err, model.ErrEmailJaCadastrado) {
		t.Error("erro 23503 NAO deveria virar ErrEmailJaCadastrado")
	}
}

func TestClienteCriar_ErroNoBanco(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("INSERT INTO clientes").
		WithArgs("bruna", "bruna@email.com", pgxmock.AnyArg()).
		WillReturnError(errors.New("conexao caiu"))

	_, err := repo.Criar(context.Background(), "bruna", "bruna@email.com", "segredo123")

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}

	if errors.Is(err, model.ErrEmailJaCadastrado) {
		t.Error("erro generico NAO deveria virar ErrEmailJaCadastrado")
	}
}

func TestClienteBuscarPorID_Sucesso(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at").
		WithArgs("1").
		WillReturnRows(pgxmock.NewRows(colunasCliente).
			AddRow("1", "bruna", "bruna@email.com", "$2a$10$hashvindodobanco", time.Now()))

	cliente, err := repo.BuscarPorID(context.Background(), "1")

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if cliente.Email != "bruna@email.com" {
		t.Errorf("Email = %q, wanted: %q", cliente.Email, "bruna@email.com")
	}
}

// pgx.ErrNoRows tem que virar o erro de dominio, nao um erro generico
func TestClienteBuscarPorID_NaoEncontrado(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at").
		WithArgs("xyz").
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.BuscarPorID(context.Background(), "xyz")

	if !errors.Is(err, model.ErrClienteNaoEncontrado) {
		t.Errorf("err = %v, wanted: %v", err, model.ErrClienteNaoEncontrado)
	}
}

func TestClienteBuscarPorID_ErroGenerico(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at").
		WithArgs("1").
		WillReturnError(errors.New("timeout"))

	_, err := repo.BuscarPorID(context.Background(), "1")

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}

	if errors.Is(err, model.ErrClienteNaoEncontrado) {
		t.Error("erro generico NAO deveria virar ErrClienteNaoEncontrado")
	}
}

func TestClienteListar_Sucesso(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at").
		WillReturnRows(pgxmock.NewRows(colunasCliente).
			AddRow("1", "bruna", "bruna@email.com", "$2a$10$hash1", time.Now()).
			AddRow("2", "ana", "ana@email.com", "$2a$10$hash2", time.Now()))

	lista, err := repo.Listar(context.Background())

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(lista) != 2 {
		t.Fatalf("len(lista) = %d, wanted: 2", len(lista))
	}

	if lista[1].Name != "ana" {
		t.Errorf("lista[1].Name = %q, wanted: %q", lista[1].Name, "ana")
	}
}

// lista vazia tem que ser slice vazio (nao nil), pra virar [] no JSON
func TestClienteListar_Vazio(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at").
		WillReturnRows(pgxmock.NewRows(colunasCliente))

	lista, err := repo.Listar(context.Background())

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if lista == nil {
		t.Fatal("lista = nil, wanted: slice vazio")
	}

	if len(lista) != 0 {
		t.Errorf("len(lista) = %d, wanted: 0", len(lista))
	}
}

func TestClienteListar_ErroNaQuery(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at").
		WillReturnError(errors.New("conexao caiu"))

	_, err := repo.Listar(context.Background())

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
}

// linha com menos colunas do que o Scan espera: falha ao ler a linha
func TestClienteListar_ErroNoScan(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at").
		WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).
			AddRow("1", "bruna"))

	_, err := repo.Listar(context.Background())

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
}

// erro que so aparece depois de percorrer as linhas: linhas.Err()
func TestClienteListar_ErroAoPercorrer(t *testing.T) {
	mock := novoMock(t)
	repo := NovoClienteRepository(mock)

	mock.ExpectQuery("SELECT id, name, email, password_hash, created_at").
		WillReturnRows(pgxmock.NewRows(colunasCliente).
			AddRow("1", "bruna", "bruna@email.com", "$2a$10$hash1", time.Now()).
			RowError(0, errors.New("conexao caiu no meio")))

	_, err := repo.Listar(context.Background())

	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
}
