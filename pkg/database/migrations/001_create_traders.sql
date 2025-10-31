-- Migration: 001_create_traders.sql
-- Descrição: Cria tabela de traders (usuários do sistema)

-- Extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabela de Traders
CREATE TABLE IF NOT EXISTS traders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    senha_hash TEXT NOT NULL,
    telefone VARCHAR(20),
    empresa VARCHAR(255),
    
    -- Status
    ativo BOOLEAN DEFAULT true,
    email_verificado BOOLEAN DEFAULT false,
    
    -- Metadados
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT email_valido CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- Índices para performance (usando IF NOT EXISTS)
CREATE INDEX IF NOT EXISTS idx_traders_email ON traders(email);
CREATE INDEX IF NOT EXISTS idx_traders_ativo ON traders(ativo);
CREATE INDEX IF NOT EXISTS idx_traders_created_at ON traders(created_at);

-- Função para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger para atualizar updated_at (com verificação se já existe)
DROP TRIGGER IF EXISTS update_traders_updated_at ON traders;
CREATE TRIGGER update_traders_updated_at 
    BEFORE UPDATE ON traders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comentários nas colunas
COMMENT ON TABLE traders IS 'Usuários traders do sistema';
COMMENT ON COLUMN traders.id IS 'Identificador único do trader';
COMMENT ON COLUMN traders.email IS 'Email único para login';
COMMENT ON COLUMN traders.senha_hash IS 'Hash bcrypt da senha';
COMMENT ON COLUMN traders.ativo IS 'Indica se o trader está ativo no sistema';