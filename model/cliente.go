package model

import "time"

type Cliente struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" faz este campo NUNCA sair no JSON
	CreatedAt    time.Time `json:"createdAt"`
}

type CriarClienteRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
