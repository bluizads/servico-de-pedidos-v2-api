package model

import "errors"

var (
	ErrProdutoNaoEncontrado   = errors.New("Produto nao encontrado")
	ErrPedidoNaoEncontrado    = errors.New("pedido nao encontrado")
	ErrQuantidadeInvalida     = errors.New("Quantidade invalida")
	ErrEstoqueInsuficiente    = errors.New("estoque insuficiente")
	ErrClienteInvalido        = errors.New("cliente inválido")
	ErrPedidoVazio            = errors.New("Pedido vazio")
	ErrMudancasStatusInvalida = errors.New("mudanca de status invalida")
	ErrClienteNaoEncontrado   = errors.New("cliente nao encontrado")
	ErrEmailJaCadastrado      = errors.New("email já cadastrado")
)
