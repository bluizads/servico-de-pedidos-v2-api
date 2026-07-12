package model

import "errors"

var (
	ErrProdutoNaoEncontrado   = errors.New("produto nao encontrado")
	ErrPedidoNaoEncontrado    = errors.New("pedido nao encontrado")
	ErrQuantidadeInvalida     = errors.New("quantidade invalida")
	ErrEstoqueInsuficiente    = errors.New("estoque insuficiente")
	ErrClienteInvalido        = errors.New("cliente inválido")
	ErrPedidoVazio            = errors.New("pedido vazio")
	ErrMudancasStatusInvalida = errors.New("mudanca de status invalida")
	ErrClienteNaoEncontrado   = errors.New("cliente nao encontrado")
	ErrEmailJaCadastrado      = errors.New("email já cadastrado")
)
