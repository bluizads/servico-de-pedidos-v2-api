package model

import "time"

type StatusPedido string

const (
	StatusPendente  StatusPedido = "PENDING"
	StatusPago      StatusPedido = "PAID"
	StatusCancelado StatusPedido = "CANCELED"
)

type ItemPedido struct {
	ID            string  `json:"id"`
	PedidoID      string  `json:"pedidoId"`
	ProdutoID     string  `json:"produtoId"`
	PrecoNaCompra float64 `json:"precoNaCompra"`
	Quantidade    int     `json:"quantidade"`
}

type Pedido struct {
	ID        string       `json:"id"`
	ClienteID string       `json:"clienteId"`
	Status    StatusPedido `json:"status"`
	Itens     []ItemPedido `json:"itens"`
	CreatedAt time.Time    `json:"createdAt"`
}

type CriarPedidoRequest struct {
	ClienteID string              `json:"clienteId"`
	Itens     []ItemPedidoRequest `json:"itens"`
}

type ItemPedidoRequest struct {
	ProdutoID  string `json:"produtoId"`
	Quantidade int    `json:"quantidade"`
}

func (compra Pedido) CalcularTotal() float64 {
	total := 0.0
	for _, item := range compra.Itens {
		total += item.PrecoNaCompra * float64(item.Quantidade)
	}
	return total
}

func (compra Pedido) PodeSerPago() bool {
	return compra.Status == StatusPendente
}

func (compra Pedido) PodeSerCancelado() bool {
	return compra.Status == StatusPendente
}
