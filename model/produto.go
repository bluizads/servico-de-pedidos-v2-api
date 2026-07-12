package model

import "time"

type Produto struct {
	ID        string    `json:"id"`
	Nome      string    `json:"nome"`
	Preco     float64   `json:"preco"`
	Estoque   int       `json:"estoque"`
	CreatedAt time.Time `json:"createdAt"`
}

func (prod Produto) TemEstoqueSuficiente(quantidade int) bool {
	return prod.Estoque >= quantidade
}
