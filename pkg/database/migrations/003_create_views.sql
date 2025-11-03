-- Migration: 003_create_views.sql
-- Descrição: Cria views para facilitar consultas

-- View: Vitrine Completa do Trader
CREATE OR REPLACE VIEW vw_vitrine_trader AS
SELECT 
    pa.id as produto_id,
    pa.trader_id,
    t.nome as trader_nome,
    t.empresa as trader_empresa,
    
    -- Dados customizados
    pa.nome_customizado,
    pa.preco_venda,
    pa.descricao as descricao_customizada,
    pa.visivel,
    pa.destaque,
    
    -- Dados originais do cavalete
    c.id as cavalete_id,
    c.nome_material,
    c.nome_espessura,
    c.nome_acabamento,
    c.nome_classificacao,
    c.comprimento,
    c.largura,
    c.altura,
    c.peso,
    c.metragem,
    c.imagem_principal,
    c.produto_cliente,
    
    -- Dados da oferta
    o.id as oferta_id,
    o.uuid_link as oferta_uuid,
    o.nome_empresa as fornecedor,
    
    -- Metadados
    pa.created_at as aprovado_em,
    pa.updated_at
FROM produtos_aprovados pa
JOIN traders t ON pa.trader_id = t.id
JOIN cavaletes c ON pa.cavalete_id = c.id
JOIN ofertas o ON c.oferta_id = o.id
WHERE pa.visivel = true AND t.ativo = true;

-- View: Estatísticas do Trader
CREATE OR REPLACE VIEW vw_estatisticas_trader AS
SELECT 
    t.id as trader_id,
    t.nome,
    t.email,
    t.empresa,
    
    -- Ofertas importadas
    COUNT(DISTINCT o.id) as total_ofertas_importadas,
    COUNT(DISTINCT c.id) as total_cavaletes_importados,
    COALESCE(SUM(c.metragem), 0) as metragem_total_importada,
    
    -- Produtos aprovados
    COUNT(DISTINCT pa.id) as total_produtos_aprovados,
    COUNT(DISTINCT CASE WHEN pa.visivel THEN pa.id END) as produtos_visiveis,
    COUNT(DISTINCT CASE WHEN pa.destaque THEN pa.id END) as produtos_destaque,
    
    -- Valores
    COALESCE(SUM(pa.preco_venda), 0) as valor_total_vitrine,
    CASE 
        WHEN COUNT(pa.id) > 0 THEN AVG(pa.preco_venda)
        ELSE 0 
    END as preco_medio_vitrine
    
FROM traders t
LEFT JOIN ofertas o ON o.trader_id = t.id
LEFT JOIN cavaletes c ON c.oferta_id = o.id
LEFT JOIN produtos_aprovados pa ON pa.trader_id = t.id
WHERE t.ativo = true
GROUP BY t.id, t.nome, t.email, t.empresa;

-- View: Resumo de Ofertas
CREATE OR REPLACE VIEW vw_ofertas_resumo AS
SELECT 
    o.id,
    o.uuid_link,
    o.trader_id,
    t.nome as trader_nome,
    o.nome_empresa,
    COUNT(DISTINCT c.id) as total_cavaletes,
    COALESCE(SUM(c.metragem), 0) as metragem_total,
    o.created_at,
    o.updated_at
FROM ofertas o
JOIN traders t ON o.trader_id = t.id
LEFT JOIN cavaletes c ON c.oferta_id = o.id
GROUP BY o.id, o.uuid_link, o.trader_id, t.nome, o.nome_empresa, o.created_at, o.updated_at;

-- Comentários
COMMENT ON VIEW vw_vitrine_trader IS 'View completa da vitrine de cada trader com joins';
COMMENT ON VIEW vw_estatisticas_trader IS 'Estatísticas agregadas por trader';
COMMENT ON VIEW vw_ofertas_resumo IS 'Resumo de ofertas com totalizadores';