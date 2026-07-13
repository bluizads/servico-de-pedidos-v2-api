package controllers

import (
	"encoding/json"
	"net/http"
)

// escreve a resposta em JSON
func responderJSON(w http.ResponseWriter, status int, dados any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dados)
}

// escreve mensagem de erro em JSON
func responderErro(w http.ResponseWriter, status int, mensagem string) {
	responderJSON(w, status, map[string]string{"erro": mensagem})
}
