
-- CLIENTES
CREATE TABLE IF NOT EXISTS clientes (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- PRODUTOS
CREATE TABLE IF NOT EXISTS produtos (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    nome VARCHAR(255) NOT NULL,
    preco NUMERIC(10, 2) NOT NULL CHECK (preco >= 0),
    estoque INTEGER NOT NULL CHECK (estoque >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);


-- PEDIDOS
-- pedidos.cliente_id aponta para clientes.id
CREATE TABLE IF NOT EXISTS pedidos (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    cliente_id UUID NOT NULL REFERENCES clientes(id),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);


-- ITENS DO PEDIDO
-- itens_pedido.pedido_id  aponta para pedidos.id
-- itens_pedido.produto_id aponta para produtos.id
CREATE TABLE IF NOT EXISTS itens_pedido (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    pedido_id UUID NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    produto_id UUID NOT NULL REFERENCES produtos(id),
    preco_na_compra NUMERIC(10, 2) NOT NULL,
    quantidade INTEGER NOT NULL CHECK (quantidade > 0)
);