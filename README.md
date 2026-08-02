# Serviço de Pedidos — API REST (v2)
 
API REST em Go com PostgreSQL, migrations, transações e testes automatizados.
Evolução da [versão 1](https://github.com/bluizads/Servico-de-Pedidos.git) (terminal + repositórios em memória).
 
**Stack:** Go · PostgreSQL 18 (com `uuidv7()` nativo) · pgx/pgxpool · chi · bcrypt · godotenv
**Testes:** cobertura ~50%, com `pgxmock` na camada de repositório (não exige banco rodando).
 
---
 
## Como rodar
 
**Pré-requisitos:** Go e PostgreSQL 18 (na porta 5432).
 
1. **Criar o banco** — no psql ou pgAdmin:
```sql
   CREATE DATABASE pedidos;
```
 
2. **Rodar as migrations** — conecte-se ao banco `pedidos` e execute
   `migrations/000001_create_tables.up.sql` (cria as tabelas `clientes`,
   `produtos`, `pedidos` e `itens_pedido`). Para desfazer, use o `.down.sql`.
3. **Configurar o `.env`** — copie `.env.example` para `.env` e preencha:
```
   DATABASE_URL=postgres://postgres:SUA_SENHA@localhost:5432/pedidos?sslmode=disable
   PORT=8080
```
 
4. **Subir a aplicação:**
```bash
   go run .
```
   O servidor sobe em `http://localhost:8080`.
 
---
 
## Testes e cobertura
 
Os testes rodam **sem precisar de banco** — a camada de repositório usa `pgxmock`
(um pool falso), e as demais camadas usam fakes que implementam as interfaces.
 
Rodar todos os testes:
```bash
go test ./...
```
 
Ver a cobertura por pacote:
```bash
go test ./... -cover
```
 
Gerar o relatório de cobertura do projeto inteiro e ver o total por função:
```bash
go test ./... -coverpkg=./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```
> No **PowerShell**, use aspas por causa do `=`:
> `go tool cover "-func=coverage.out"`
 
Ver o relatório visual no navegador (linhas verdes = cobertas, vermelhas = não):
```bash
go tool cover -html=coverage.out
```
 
**O que é testado:** regras de domínio (`model`), tradução de erros e respostas
HTTP (`controllers`), acesso ao banco e mapeamento de erros do Postgres
(`repository`) e a leitura de configuração (`config`).
 
---
 
## Testando a API manualmente
 
Com a API rodando, use o script automatizado (PowerShell) que executa 15 cenários
em sequência (fluxos de sucesso e de erro), capturando os IDs automaticamente:
 
```powershell
.\testar-api.ps1
```
> Se o PowerShell bloquear: `Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass`
 
O script demonstra: criação de cliente (sem expor o hash), congelamento do preço na
compra, redução/devolução de estoque, a **prova da transação** (após falha por
estoque, o rollback mantém o estoque intacto) e os status de erro 400, 404 e 409.
 
### Exemplos avulsos (PowerShell)
```powershell
# criar cliente
Invoke-RestMethod -Uri http://localhost:8080/clientes -Method Post -ContentType "application/json" -Body '{"name":"Ana","email":"ana@email.com","password":"senha123"}'
 
# criar produto
Invoke-RestMethod -Uri http://localhost:8080/produtos -Method Post -ContentType "application/json" -Body '{"nome":"Notebook","preco":3500,"estoque":5}'
 
# criar pedido (troque os IDs)
Invoke-RestMethod -Uri http://localhost:8080/pedidos -Method Post -ContentType "application/json" -Body '{"clienteId":"ID","itens":[{"produtoId":"ID","quantidade":1}]}'
 
# pagar / cancelar
Invoke-RestMethod -Uri http://localhost:8080/pedidos/ID/pagar -Method Post
 
# paginação
Invoke-RestMethod -Uri "http://localhost:8080/pedidos?limit=5&offset=0"
```
 
---
 
## Arquitetura
 
Organizado em camadas, com as dependências apontando sempre numa direção:
 
```
main → routes → controllers → repository → banco
                                  ↓
                                model
```
 
| Camada | Responsabilidade |
|---|---|
| `model/` | entidades, DTOs e erros de domínio |
| `repository/` | acesso ao banco (SQL); interface `DB` permite injetar mock nos testes |
| `controllers/` | HTTP e JSON: decodifica requisição, traduz erro → status |
| `routes/` | mapeia URLs para os controllers |
| `config/` | lê variáveis de ambiente (devolve erro; quem decide sair é a `main`) |
| `database/` | abre o pool de conexões |
| `migrations/` | scripts SQL de criação das tabelas |
 
As regras de negócio ficam no domínio/repository — **nunca** no controller.
 
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
                            produtos                                        │
                           ┌──────────────┐                                 │
                           │ id (PK)      │◄────────────────────────────────┘
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
| GET | `/clientes` · `/clientes/{id}` | lista / busca cliente |
| POST | `/produtos` | cadastra produto |
| GET | `/produtos` · `/produtos/{id}` | lista / busca produto |
| POST | `/pedidos` | cria pedido (dentro de transação) |
| GET | `/pedidos?limit=10&offset=0` · `/pedidos/{id}` | lista com paginação / busca |
| POST | `/pedidos/{id}/pagar` · `/pedidos/{id}/cancelar` | paga / cancela (devolve estoque) |
| GET | `/health` | verifica se a API está no ar |
 
---
 
## Regras de negócio
 
- cliente do pedido é obrigatório e precisa existir; pedido precisa de ao menos um item;
- quantidade > 0; produto precisa existir e ter estoque suficiente;
- ao criar pedido o estoque diminui, e o preço é **congelado** no momento da criação;
- pedido nasce `PENDING`; pode virar `PAID` ou `CANCELED` (cancelar devolve o estoque);
- pedido pago ou cancelado não muda de status novamente;
- a senha nunca é salva em texto puro (apenas o hash bcrypt) e o hash nunca sai no JSON.
## Transação
 
A criação do pedido acontece dentro de uma transação, pois altera três tabelas
(insere o pedido, insere os itens, reduz o estoque). Se qualquer etapa falhar, o
`defer transacao.Rollback(...)` desfaz **tudo** — sem pedidos órfãos nem estoque
descontado indevidamente.
 
## Status HTTP
 
| Situação | Status |
|---|---|
| criação / leitura bem-sucedida | 201 / 200 |
| dados inválidos | 400 |
| cliente/produto/pedido não encontrado | 404 |
| estoque insuficiente · email já cadastrado · status inválido | 409 |
| erro inesperado | 500 |