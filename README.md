# Serviço de Pedidos — API REST (v2)

API REST em Go com PostgreSQL, migrations e transações.
Evolução da [versão 1](https://github.com/bluizads/Servico-de-Pedidos.git) (terminal + repositórios em memória).

- **Go** — linguagem
- **PostgreSQL 18** — banco de dados (com `uuidv7()` nativo)
- **pgx / pgxpool** — driver e pool de conexões
- **chi** — roteador HTTP
- **bcrypt** — hash de senha
- **godotenv** — variáveis de ambiente

---


## Como rodar

### 1. Pré-requisitos

- Go instalado
- PostgreSQL 18 instalado e rodando na porta 5432

### 2. Criar o banco

No pgAdmin (ou psql), execute:

```sql
CREATE DATABASE pedidos;
```

### 3. Rodar as migrations

Conecte-se ao banco `pedidos` e execute o conteúdo de:

```
migrations/000001_create_tables.up.sql
```

Isso cria as 4 tabelas: `clientes`, `produtos`, `pedidos` e `itens_pedido`,
com as chaves estrangeiras entre elas.

Para desfazer, execute `migrations/000001_create_tables.down.sql`.

### 4. Configurar o `.env`

Copie o `.env.example` para `.env` e preencha com a sua senha do Postgres:

```
DATABASE_URL=postgres://postgres:SUA_SENHA@localhost:5432/pedidos?sslmode=disable
PORT=8080
```

### 5. Rodar a aplicação

```bash
go run .
```

O servidor sobe em `http://localhost:8080`.

---


## Testando

### Opção rápida: script automatizado

Com a API rodando (`go run .`), abra outro terminal e rode:

```powershell
.\testar-api.ps1
```

O script executa 15 cenários em sequência (fluxos bem-sucedidos + fluxos de erro),
capturando os IDs automaticamente. Ele demonstra:

- criação de cliente (sem expor o hash da senha), produtos e pedido;
- congelamento do preço no momento da compra;
- redução e devolução de estoque;
- **prova da transação**: após uma falha por estoque insuficiente, o estoque
  permanece intacto (o rollback desfez o pedido);
- todos os status de erro: 400, 404 e 409.

Se o PowerShell bloquear a execução, rode antes:

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
```

### Opção manual: comando a comando (PowerShell)

#### Fluxo bem-sucedido

**1. Criar cliente**
```powershell
Invoke-RestMethod -Uri http://localhost:8080/clientes -Method Post -ContentType "application/json" -Body '{"name":"Ana","email":"ana@email.com","password":"senha123"}'
```

**2. Criar produtos**
```powershell
Invoke-RestMethod -Uri http://localhost:8080/produtos -Method Post -ContentType "application/json" -Body '{"nome":"Notebook","preco":3500,"estoque":5}'

Invoke-RestMethod -Uri http://localhost:8080/produtos -Method Post -ContentType "application/json" -Body '{"nome":"Mouse","preco":80,"estoque":10}'
```

**3. Criar pedido** (substitua os IDs pelos retornados acima)
```powershell
Invoke-RestMethod -Uri http://localhost:8080/pedidos -Method Post -ContentType "application/json" -Body '{"clienteId":"ID_DO_CLIENTE","itens":[{"produtoId":"ID_DO_NOTEBOOK","quantidade":1},{"produtoId":"ID_DO_MOUSE","quantidade":2}]}'
```

**4. Conferir que o estoque diminuiu** (Notebook 5→4, Mouse 10→8)
```powershell
Invoke-RestMethod -Uri http://localhost:8080/produtos
```

**5. Pagar o pedido** (status vira PAID)
```powershell
Invoke-RestMethod -Uri http://localhost:8080/pedidos/ID_DO_PEDIDO/pagar -Method Post
```

#### Fluxos de erro

**Cancelar pedido já pago → 409**
```powershell
Invoke-RestMethod -Uri http://localhost:8080/pedidos/ID_DO_PEDIDO/cancelar -Method Post
```

**Email duplicado → 409**
```powershell
Invoke-RestMethod -Uri http://localhost:8080/clientes -Method Post -ContentType "application/json" -Body '{"name":"Outra","email":"ana@email.com","password":"123"}'
```

**Estoque insuficiente → 409** (e o estoque permanece intacto, provando o rollback)
```powershell
Invoke-RestMethod -Uri http://localhost:8080/pedidos -Method Post -ContentType "application/json" -Body '{"clienteId":"ID_DO_CLIENTE","itens":[{"produtoId":"ID_DO_NOTEBOOK","quantidade":999}]}'
```

**Pedido inexistente → 404**
```powershell
Invoke-RestMethod -Uri http://localhost:8080/pedidos/00000000-0000-0000-0000-000000000000
```

### Paginação
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/pedidos?limit=5&offset=0"
```

---

## Arquitetura

O projeto é organizado em camadas, com as dependências apontando sempre numa direção:

```
main → routes → controllers → repository → banco
                                  ↓
                                model
```

|     Camada     |        Responsabilidade                                 |
|----------------|---------------------------------------------------------|
|    `model/`    | entidades, DTOs e erros de domínio                      |
|  `repository/` | acesso ao banco (SQL)                                   |
| `controllers/` | HTTP e JSON: decodifica requisição, traduz erro → status|
|   `routes/`    | mapeia URLs para os controllers                         |
|    `config/`   | lê variáveis de ambiente                                |
|  `database/`   | abre o pool de conexões                                 |
| `migrations/`  | scripts SQL de criação das tabelas                      |

As regras de negócio ficam no domínio/repository — **nunca** no controller.

### Estrutura das pastas:
servico-de-pedidos-v2-api/
    .env.example
    .gitignore
    go.mod
    main.go
    migrations/
    config/
    database/
    model/
    repository/
    controllers/
    routes/

---

## Modelagem do banco

```
clientes                    pedidos                   itens_pedido
┌──────────────┐           ┌──────────────┐          ┌──────────────────┐
│ id (PK)      │◄──────────│ cliente_id   │◄─────────│ pedido_id        │
│ name         │           │ id (PK)      │          │ id (PK)          │
│ email UNIQUE │           │ status       │          │ produto_id       │──┐
│ password_hash│           │ created_at   │          │ preco_na_compra  │  │
│ created_at   │           └──────────────┘          │ quantidade       │  │
└──────────────┘                                     └──────────────────┘  │
                            produtos                                       │
                           ┌──────────────┐                                │
                           │ id (PK)      │◄───────────────────────────────┘
                           │ nome         │
                           │ preco        │
                           │ estoque      │
                           │ created_at   │
                           └──────────────┘
```

---

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| POST | `/clientes` | cadastra cliente (gera hash da senha) |
| GET | `/clientes` | lista clientes |
| GET | `/clientes/{id}` | busca cliente por id |
| POST | `/produtos` | cadastra produto |
| GET | `/produtos` | lista produtos |
| GET | `/produtos/{id}` | busca produto por id |
| POST | `/pedidos` | cria pedido (dentro de transação) |
| GET | `/pedidos?limit=10&offset=0` | lista pedidos com paginação |
| GET | `/pedidos/{id}` | busca pedido por id |
| POST | `/pedidos/{id}/pagar` | paga o pedido |
| POST | `/pedidos/{id}/cancelar` | cancela o pedido e devolve o estoque |
| GET | `/health` | verifica se a API está no ar |

---

## Regras de negócio

- cliente do pedido é obrigatório e precisa existir;
- pedido precisa ter pelo menos um item;
- quantidade deve ser maior que zero;
- produto precisa existir e ter estoque suficiente;
- ao criar pedido, o estoque diminui;
- o preço usado no pedido é o preço do produto no momento da criação (congelado);
- pedido nasce como `PENDING`;
- pedido pago vira `PAID`;
- pedido cancelado vira `CANCELED` e devolve o estoque;
- pedido pago ou cancelado não pode mudar de status novamente;
- senha nunca é salva em texto puro — apenas o hash (bcrypt);
- o hash nunca é retornado no JSON.

---

## Transação

A criação do pedido acontece dentro de uma transação, pois altera três tabelas:
insere o pedido, insere os itens e reduz o estoque dos produtos.

Se qualquer etapa falhar (produto inexistente, estoque insuficiente), o
`defer transacao.Rollback(...)` desfaz **tudo** — o banco volta ao estado anterior,
sem pedidos órfãos nem estoque descontado indevidamente.

---

## Status HTTP

| Situação | Status |
|---|---|
| criação bem-sucedida | 201 |
| leitura bem-sucedida | 200 |
| dados inválidos | 400 |
| cliente/produto/pedido não encontrado | 404 |
| estoque insuficiente | 409 |
| email já cadastrado | 409 |
| mudança de status inválida | 409 |
| erro inesperado | 500 |

---

