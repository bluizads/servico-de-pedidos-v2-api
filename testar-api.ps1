# ====================================================
# Script de teste da API de Pedidos
#
# Como usar:
#   1. Deixe a API rodando:     go run .                em um terminal
#   2. Rode:                    .\testar-api.ps1        em outro terminal
#
# Se o PowerShell bloquear a execucao, rode antes:
#   Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
# =====================================================

$base = "http://localhost:8080"

# Gera um email unico a cada execucao, para nao colidir com testes anteriores
$emailUnico = "ana$(Get-Random)@email.com"

function Titulo($texto) {
    Write-Host ""
    Write-Host "=== $texto ===" -ForegroundColor Cyan
}

function Esperado($texto) {
    Write-Host "  [esperado] $texto" -ForegroundColor Yellow
}

# Chama a API e mostra o erro
function Chamar($metodo, $rota, $corpo) {
    try {
        if ($corpo) {
            $resultado = Invoke-RestMethod -Uri "$base$rota" -Method $metodo -ContentType "application/json" -Body $corpo
        }
        else {
            $resultado = Invoke-RestMethod -Uri "$base$rota" -Method $metodo
        }
        return $resultado
    }
    catch {
        $status = $_.Exception.Response.StatusCode.value__
        $mensagem = $_.ErrorDetails.Message
        Write-Host "  -> HTTP $status  $mensagem" -ForegroundColor Red
        return $null
    }
}

# ===================== CASOS DE SUCESSO ===========================

Titulo "1. Criar cliente"
$cliente = Chamar Post "/clientes" "{""name"":""Ana"",""email"":""$emailUnico"",""password"":""senha123""}"
$cliente | Format-List
Write-Host "  (repare: o password_hash NAO aparece na resposta)" -ForegroundColor DarkGray

Titulo "2. Criar produtos"
$notebook = Chamar Post "/produtos" '{"nome":"Notebook","preco":3500,"estoque":5}'
$mouse    = Chamar Post "/produtos" '{"nome":"Mouse","preco":80,"estoque":10}'
Write-Host "  Notebook: $($notebook.id)  estoque=$($notebook.estoque)"
Write-Host "  Mouse:    $($mouse.id)  estoque=$($mouse.estoque)"

Titulo "3. Criar pedido (1 Notebook + 2 Mouses)"
$corpoPedido = @"
{
  "clienteId": "$($cliente.id)",
  "itens": [
    { "produtoId": "$($notebook.id)", "quantidade": 1 },
    { "produtoId": "$($mouse.id)", "quantidade": 2 }
  ]
}
"@
$pedido = Chamar Post "/pedidos" $corpoPedido
Write-Host "  Pedido:  $($pedido.id)"
Write-Host "  Status:  $($pedido.status)"
Write-Host "  Itens:"
foreach ($item in $pedido.itens) {
    Write-Host "    - produto=$($item.produtoId)  qtd=$($item.quantidade)  precoNaCompra=$($item.precoNaCompra)"
}
Write-Host "  (o preco foi CONGELADO no momento da compra)" -ForegroundColor DarkGray

Titulo "4. Estoque apos o pedido"
Esperado "Notebook 5->4, Mouse 10->8"
$nb = Chamar Get "/produtos/$($notebook.id)"
$ms = Chamar Get "/produtos/$($mouse.id)"
Write-Host "  Notebook: estoque = $($nb.estoque)"
Write-Host "  Mouse:    estoque = $($ms.estoque)"

Titulo "5. Pagar o pedido"
Esperado "status vira PAID"
$pago = Chamar Post "/pedidos/$($pedido.id)/pagar"
Write-Host "  Status: $($pago.status)"

Titulo "6. Buscar o pedido pago"
$encontrado = Chamar Get "/pedidos/$($pedido.id)"
Write-Host "  Pedido $($encontrado.id) | status=$($encontrado.status) | $($encontrado.itens.Count) itens"

# ===================== CASOS DE ERRO ===========================

Titulo "7. ERRO: cancelar pedido ja pago"
Esperado "409 - mudanca de status invalida"
Chamar Post "/pedidos/$($pedido.id)/cancelar" | Out-Null

Titulo "8. ERRO: email ja cadastrado"
Esperado "409 - email ja cadastrado"
Chamar Post "/clientes" "{""name"":""Outra"",""email"":""$emailUnico"",""password"":""123""}" | Out-Null

Titulo "9. ERRO: estoque insuficiente"
Esperado "409 - estoque insuficiente"
$corpoEstoque = @"
{
  "clienteId": "$($cliente.id)",
  "itens": [ { "produtoId": "$($notebook.id)", "quantidade": 999 } ]
}
"@
Chamar Post "/pedidos" $corpoEstoque | Out-Null

Titulo "10. PROVA DA TRANSACAO: o estoque continua intacto?"
Esperado "Notebook ainda em 4 (o rollback desfez o pedido que falhou)"
$nb = Chamar Get "/produtos/$($notebook.id)"
Write-Host "  Notebook: estoque = $($nb.estoque)"

Titulo "11. ERRO: pedido sem cliente"
Esperado "400 - clienteId e obrigatorio"
Chamar Post "/pedidos" '{"clienteId":"","itens":[]}' | Out-Null

Titulo "12. ERRO: pedido inexistente"
Esperado "404 - pedido nao encontrado"
Chamar Get "/pedidos/00000000-0000-0000-0000-000000000000" | Out-Null

# ================= CANCELAMENTO (estoque volta) ==========================

Titulo "13. Criar novo pedido para cancelar"
$corpoCancelar = @"
{
  "clienteId": "$($cliente.id)",
  "itens": [ { "produtoId": "$($mouse.id)", "quantidade": 3 } ]
}
"@
$pedido2 = Chamar Post "/pedidos" $corpoCancelar
$ms = Chamar Get "/produtos/$($mouse.id)"
Write-Host "  Pedido criado: $($pedido2.id) | status=$($pedido2.status)"
Write-Host "  Mouse em estoque: $($ms.estoque)  (era 8, comprou 3)"

Titulo "14. Cancelar o pedido (estoque deve VOLTAR)"
Esperado "status CANCELED e Mouse volta para 8"
$cancelado = Chamar Post "/pedidos/$($pedido2.id)/cancelar"
Write-Host "  Status: $($cancelado.status)"
$ms = Chamar Get "/produtos/$($mouse.id)"
Write-Host "  Mouse em estoque: $($ms.estoque)"

# ===================== PAGINACAO ==================================

Titulo "15. Paginacao (limit=1&offset=0)"
$pagina = Chamar Get "/pedidos?limit=1&offset=0"
Write-Host "  Retornou $(@($pagina).Count) pedido(s)"

Write-Host ""
Write-Host "=== FIM DOS TESTES ===" -ForegroundColor Green
Write-Host ""