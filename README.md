# Serviço de Pedidos versão2 API
API REST de pedidos em Go com PostgreSQL, migrations e transações. Evolui a v1 (terminal + repositórios em memória) para uma API HTTP com persistência real.


Estrutura das pastas:
servico-de-pedidos-v2-api/
├── .env
├── .gitignore
├── go.mod
├── main.go
├── migrations/
│   ├── 000001_create_tables.up.sql
│   └── 000001_create_tables.down.sql
├── config/
├── database/
├── model/
├── repository/
├── controllers/
└── routes/