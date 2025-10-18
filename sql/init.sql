-- Criação do banco de dados e extensões
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabela de ofertas
CREATE TABLE IF NOT EXISTS ofertas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uuid_link VARCHAR(255) UNIQUE NOT NULL,
    situacao VARCHAR(100) NOT NULL,
    nome_empresa VARCHAR(255) NOT NULL,
    url_logo TEXT,
    dados_completos JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabela de cavaletes
CREATE TABLE IF NOT EXISTS cavaletes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    oferta_id UUID NOT NULL REFERENCES ofertas(id) ON DELETE CASCADE,
    codigo VARCHAR(255) NOT NULL,
    bloco VARCHAR(255) NOT NULL,
    nome_material VARCHAR(255) NOT NULL,
    nome_espessura VARCHAR(255) NOT NULL,
    nome_classificacao VARCHAR(255),
    nome_acabamento VARCHAR(255),
    comprimento DECIMAL(10,3),
    altura DECIMAL(10,3),
    largura DECIMAL(10,3),
    metragem DECIMAL(10,3),
    peso DECIMAL(10,3),
    tipo_metragem VARCHAR(50),
    aprovado BOOLEAN DEFAULT FALSE,
    importado BOOLEAN DEFAULT FALSE,
    descricao_chapas TEXT,
    quantidade_itens INTEGER,
    valor DECIMAL(12,2),
    observacao TEXT,
    observacao_conferencia TEXT,
    produto_cliente VARCHAR(255),
    espessura_cliente VARCHAR(255),
    imagem_principal JSONB,
    imagens_adicionais JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabela de itens
CREATE TABLE IF NOT EXISTS itens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cavalete_id UUID NOT NULL REFERENCES cavaletes(id) ON DELETE CASCADE,
    codigo VARCHAR(255) NOT NULL,
    bloco VARCHAR(255) NOT NULL,
    nome_espessura VARCHAR(255) NOT NULL,
    nome_classificacao VARCHAR(255) NOT NULL,
    nome_acabamento VARCHAR(255),
    comprimento DECIMAL(10,3),
    altura DECIMAL(10,3),
    largura DECIMAL(10,3),
    metragem DECIMAL(10,3),
    peso DECIMAL(10,3),
    tipo_metragem VARCHAR(50),
    aprovado BOOLEAN DEFAULT FALSE,
    importado BOOLEAN DEFAULT FALSE,
    valor DECIMAL(12,2),
    observacao TEXT,
    observacao_conferencia TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Índices para melhor performance
CREATE INDEX IF NOT EXISTS idx_ofertas_uuid_link ON ofertas(uuid_link);
CREATE INDEX IF NOT EXISTS idx_ofertas_created_at ON ofertas(created_at);
CREATE INDEX IF NOT EXISTS idx_cavaletes_oferta_id ON cavaletes(oferta_id);
CREATE INDEX IF NOT EXISTS idx_cavaletes_codigo ON cavaletes(codigo);
CREATE INDEX IF NOT EXISTS idx_itens_cavalete_id ON itens(cavalete_id);
CREATE INDEX IF NOT EXISTS idx_itens_codigo ON itens(codigo);

-- Função para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para atualizar updated_at
CREATE TRIGGER update_ofertas_updated_at BEFORE UPDATE ON ofertas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cavaletes_updated_at BEFORE UPDATE ON cavaletes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_itens_updated_at BEFORE UPDATE ON itens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- SISTEMA DE AUTENTICAÇÃO
-- ========================================

-- Tabela de traders (usuários do sistema)
CREATE TABLE IF NOT EXISTS traders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    senha_hash VARCHAR(255) NOT NULL,
    telefone VARCHAR(20),
    empresa VARCHAR(255),
    ativo BOOLEAN DEFAULT TRUE,
    email_verificado BOOLEAN DEFAULT FALSE,
    ultimo_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabela de produtos aprovados (vitrine do trader)
CREATE TABLE IF NOT EXISTS produtos_aprovados (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trader_id UUID NOT NULL REFERENCES traders(id) ON DELETE CASCADE,
    cavalete_id UUID NOT NULL REFERENCES cavaletes(id) ON DELETE CASCADE,
    nome_customizado VARCHAR(255) NOT NULL,
    preco_venda DECIMAL(12,2) NOT NULL,
    descricao TEXT,
    visivel BOOLEAN DEFAULT TRUE,
    destaque BOOLEAN DEFAULT FALSE,
    ordem_exibicao INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(trader_id, cavalete_id)
);

-- Tabela de refresh tokens para JWT
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trader_id UUID NOT NULL REFERENCES traders(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revogado BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Vincular ofertas aos traders
ALTER TABLE ofertas ADD COLUMN IF NOT EXISTS trader_id UUID REFERENCES traders(id) ON DELETE SET NULL;

-- Índices para autenticação
CREATE INDEX IF NOT EXISTS idx_traders_email ON traders(email);
CREATE INDEX IF NOT EXISTS idx_traders_ativo ON traders(ativo);
CREATE INDEX IF NOT EXISTS idx_produtos_aprovados_trader_id ON produtos_aprovados(trader_id);
CREATE INDEX IF NOT EXISTS idx_produtos_aprovados_visivel ON produtos_aprovados(visivel);
CREATE INDEX IF NOT EXISTS idx_produtos_aprovados_destaque ON produtos_aprovados(destaque);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_trader_id ON refresh_tokens(trader_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_ofertas_trader_id ON ofertas(trader_id);

-- Triggers para autenticação
CREATE TRIGGER update_traders_updated_at BEFORE UPDATE ON traders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_produtos_aprovados_updated_at BEFORE UPDATE ON produtos_aprovados
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- View para cavaletes disponíveis para aprovação (por trader)
CREATE OR REPLACE VIEW cavaletes_disponiveis AS
SELECT 
    c.*,
    o.trader_id,
    o.nome_empresa,
    CASE WHEN pa.id IS NOT NULL THEN TRUE ELSE FALSE END as ja_aprovado
FROM cavaletes c
INNER JOIN ofertas o ON c.oferta_id = o.id
LEFT JOIN produtos_aprovados pa ON c.id = pa.cavalete_id
WHERE o.trader_id IS NOT NULL;

-- View para vitrine pública do trader
CREATE OR REPLACE VIEW vitrine_publica AS
SELECT 
    pa.id,
    pa.trader_id,
    pa.nome_customizado,
    pa.preco_venda,
    pa.descricao,
    pa.destaque,
    pa.ordem_exibicao,
    c.codigo,
    c.bloco,
    c.nome_material,
    c.nome_espessura,
    c.nome_classificacao,
    c.nome_acabamento,
    c.comprimento,
    c.altura,
    c.largura,
    c.metragem,
    c.peso,
    c.tipo_metragem,
    c.imagem_principal,
    c.imagens_adicionais,
    t.nome as trader_nome,
    t.empresa as trader_empresa,
    pa.created_at,
    pa.updated_at
FROM produtos_aprovados pa
INNER JOIN cavaletes c ON pa.cavalete_id = c.id
INNER JOIN traders t ON pa.trader_id = t.id
WHERE pa.visivel = TRUE AND t.ativo = TRUE
ORDER BY pa.destaque DESC, pa.ordem_exibicao ASC, pa.created_at DESC;