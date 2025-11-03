-- Migration: 002_create_ofertas.sql
-- Descrição: Cria estrutura completa de ofertas, cavaletes e produtos

-- Tabela de Ofertas (importadas do Mobgran)
CREATE TABLE IF NOT EXISTS ofertas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uuid_link VARCHAR(255) UNIQUE NOT NULL,
    trader_id UUID NOT NULL REFERENCES traders(id) ON DELETE CASCADE,
    
    -- Dados básicos
    nome_empresa VARCHAR(500),
    nome_vendedor VARCHAR(255),
    
    -- Dados completos em JSON
    dados_completos JSONB,
    
    -- Metadados
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Cavaletes (blocos de pedra)
CREATE TABLE IF NOT EXISTS cavaletes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    oferta_id UUID NOT NULL REFERENCES ofertas(id) ON DELETE CASCADE,
    oferta_uuid_link VARCHAR(255) NOT NULL,
    
    -- Informações do material
    nome_material VARCHAR(255) NOT NULL,
    nome_espessura VARCHAR(100),
    nome_acabamento VARCHAR(100),
    
    -- Dimensões
    largura DECIMAL(10,2),
    altura DECIMAL(10,2),
    peso DECIMAL(10,2),
    metragem DECIMAL(10,2),
    
    -- Imagem
    imagem_url TEXT,
    
    -- Classificação
    classificacao VARCHAR(100),
    
    -- Dados completos
    dados_completos JSONB,
    
    -- Metadados
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Itens (chapas individuais)
CREATE TABLE IF NOT EXISTS itens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cavalete_id UUID NOT NULL REFERENCES cavaletes(id) ON DELETE CASCADE,
    oferta_id UUID NOT NULL REFERENCES ofertas(id) ON DELETE CASCADE,
    
    -- Dimensões
    largura DECIMAL(10,2),
    altura DECIMAL(10,2),
    peso DECIMAL(10,2),
    metragem DECIMAL(10,2),
    
    -- Dados completos
    dados_completos JSONB,
    
    -- Metadados
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Produtos Aprovados (Vitrine do Trader)
CREATE TABLE IF NOT EXISTS produtos_aprovados (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trader_id UUID NOT NULL REFERENCES traders(id) ON DELETE CASCADE,
    cavalete_id UUID NOT NULL REFERENCES cavaletes(id) ON DELETE CASCADE,
    oferta_id UUID NOT NULL REFERENCES ofertas(id) ON DELETE CASCADE,
    
    -- Dados customizados pelo trader
    nome_customizado VARCHAR(500) NOT NULL,
    preco_venda DECIMAL(12,2) NOT NULL CHECK (preco_venda > 0),
    descricao TEXT,
    
    -- Status
    visivel BOOLEAN DEFAULT true,
    destaque BOOLEAN DEFAULT false,
    
    -- Metadados
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint: um cavalete só pode ser aprovado uma vez por trader
    CONSTRAINT unique_trader_cavalete UNIQUE(trader_id, cavalete_id)
);

-- Índices (usando IF NOT EXISTS)
CREATE INDEX IF NOT EXISTS idx_ofertas_trader_id ON ofertas(trader_id);
CREATE INDEX IF NOT EXISTS idx_ofertas_uuid_link ON ofertas(uuid_link);
CREATE INDEX IF NOT EXISTS idx_cavaletes_oferta_id ON cavaletes(oferta_id);
CREATE INDEX IF NOT EXISTS idx_cavaletes_material ON cavaletes(nome_material);
CREATE INDEX IF NOT EXISTS idx_itens_cavalete_id ON itens(cavalete_id);
CREATE INDEX IF NOT EXISTS idx_produtos_trader_id ON produtos_aprovados(trader_id);
CREATE INDEX IF NOT EXISTS idx_produtos_visivel ON produtos_aprovados(visivel);
CREATE INDEX IF NOT EXISTS idx_produtos_destaque ON produtos_aprovados(destaque);

-- Triggers para updated_at (com verificação se já existem)
DROP TRIGGER IF EXISTS update_ofertas_updated_at ON ofertas;
CREATE TRIGGER update_ofertas_updated_at 
    BEFORE UPDATE ON ofertas
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_cavaletes_updated_at ON cavaletes;
CREATE TRIGGER update_cavaletes_updated_at 
    BEFORE UPDATE ON cavaletes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_produtos_updated_at ON produtos_aprovados;
CREATE TRIGGER update_produtos_updated_at 
    BEFORE UPDATE ON produtos_aprovados
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comentários
COMMENT ON TABLE ofertas IS 'Ofertas importadas do Mobgran';
COMMENT ON TABLE cavaletes IS 'Blocos/cavaletes de pedra';
COMMENT ON TABLE itens IS 'Chapas individuais de cada cavalete';
COMMENT ON TABLE produtos_aprovados IS 'Produtos aprovados para vitrine do trader';