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