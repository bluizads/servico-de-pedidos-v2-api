package model

import "testing"

// Como rodar:
// go test ./... -coverpkg=./... "-coverprofile=$HOME\cobertura.out"
//go tool cover "-func=$HOME\cobertura.out"

func TestCalcularTotal(t *testing.T) {
	casos := []struct {
		nome   string
		itens  []ItemPedido
		wanted float64
	}{
		{"- Sem itens", nil, 0},
		{"- Um item", []ItemPedido{{PrecoNaCompra: 10, Quantidade: 2}}, 20},
		{"- Vários itens", []ItemPedido{
			{PrecoNaCompra: 3500, Quantidade: 1},
			{PrecoNaCompra: 25, Quantidade: 2},
			{PrecoNaCompra: 1.33, Quantidade: 5},
		}, 3_556.65},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			p := Pedido{Itens: c.itens}
			if got := p.CalcularTotal(); got != c.wanted {
				t.Errorf("CalcularTotal() = %v, desejado %v", got, c.wanted)
			}
		})
	}
}

func TestPodeSerPagoECancelado(t *testing.T) {
	casos := []struct {
		nome   string
		status StatusPedido
		pode   bool
	}{
		{"pendente", StatusPendente, true},
		{"pago", StatusPago, false},
		{"cancelado", StatusCancelado, false},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			p := Pedido{Status: c.status}
			if p.PodeSerPago() != c.pode {
				t.Errorf("PodeSerPago() = %v, desejado %v", p.PodeSerPago(), c.pode)
			}
			if p.PodeSerCancelado() != c.pode {
				t.Errorf("PodeSerCancelado() = %v, desejado %v", p.PodeSerCancelado(), c.pode)
			}
		})

	}
}

func TestTemEstoqueSuficiente(t *testing.T) {
	casos := []struct {
		nome       string
		quantidade int
		tem        bool
	}{
		{"Estoque igual", 5, true},
		{"Estoque maior", 3, true},
		{"Estoque menor", 6, false},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			p := Produto{Estoque: 5}
			if p.TemEstoqueSuficiente(c.quantidade) != c.tem {
				t.Errorf("TemEstoqueSuficiente(%d) = %v, desejado %v", c.quantidade, p.TemEstoqueSuficiente(c.quantidade), c.tem)
			}
		})
	}
}
