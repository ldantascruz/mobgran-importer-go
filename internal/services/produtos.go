package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"mobgran-importer-go/internal/models"
)

// ProdutosService gerencia operações relacionadas a produtos
type ProdutosService struct {
	db *sql.DB
}

// NewProdutosService cria uma nova instância do ProdutosService
func NewProdutosService(db *sql.DB) *ProdutosService {
	return &ProdutosService{db: db}
}

// ListarCavaletesDisponiveis lista cavaletes disponíveis para aprovação
func (s *ProdutosService) ListarCavaletesDisponiveis(traderID uuid.UUID, limit, offset int) ([]models.CavaleteDisponivel, error) {
	query := `
		SELECT 
			c.id, c.oferta_id, c.codigo, c.bloco, c.nome_material, c.nome_espessura,
			c.nome_classificacao, c.nome_acabamento, c.comprimento, c.altura, c.largura,
			c.metragem, c.peso, c.tipo_metragem, c.imagem_principal, c.imagens_adicionais,
			c.created_at, c.updated_at,
			o.trader_id, o.nome_empresa,
			CASE WHEN pa.id IS NOT NULL THEN true ELSE false END as ja_aprovado
		FROM cavaletes c
		JOIN ofertas o ON c.oferta_id = o.id
		LEFT JOIN produtos_aprovados pa ON pa.cavalete_id = c.id AND pa.trader_id = $1
		WHERE o.situacao = 'ativa' AND o.trader_id = $1
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, traderID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Erro ao buscar cavaletes disponíveis")
		return nil, fmt.Errorf("erro ao buscar cavaletes disponíveis")
	}
	defer rows.Close()

	var cavaletes []models.CavaleteDisponivel
	for rows.Next() {
		var c models.CavaleteDisponivel
		err := rows.Scan(
			&c.ID, &c.OfertaID, &c.Codigo, &c.Bloco, &c.NomeMaterial, &c.NomeEspessura,
			&c.NomeClassificacao, &c.NomeAcabamento, &c.Comprimento, &c.Altura, &c.Largura,
			&c.Metragem, &c.Peso, &c.TipoMetragem, &c.ImagemPrincipal, &c.ImagensAdicionais,
			&c.CreatedAt, &c.UpdatedAt,
			&c.TraderID, &c.NomeEmpresa, &c.JaAprovado,
		)
		if err != nil {
			logrus.WithError(err).Error("Erro ao escanear cavalete disponível")
			continue
		}
		cavaletes = append(cavaletes, c)
	}

	return cavaletes, nil
}

// AprovarProduto aprova um cavalete como produto do trader
func (s *ProdutosService) AprovarProduto(traderID uuid.UUID, request *models.ProdutoAprovarRequest) (*models.ProdutoAprovado, error) {
	// Verifica se o cavalete existe e está disponível
	var cavaleteExists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM cavaletes_disponiveis cd
			JOIN ofertas o ON cd.oferta_id = o.uuid_link
			WHERE cd.id = $1 AND o.situacao = 'ativa'
		)
	`, request.CavaleteID).Scan(&cavaleteExists)

	if err != nil {
		logrus.WithError(err).Error("Erro ao verificar cavalete")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	if !cavaleteExists {
		return nil, fmt.Errorf("cavalete não encontrado ou não disponível")
	}

	// Verifica se já foi aprovado pelo trader
	var jaAprovado bool
	err = s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM produtos_aprovados
			WHERE trader_id = $1 AND cavalete_id = $2
		)
	`, traderID, request.CavaleteID).Scan(&jaAprovado)

	if err != nil {
		logrus.WithError(err).Error("Erro ao verificar produto já aprovado")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	if jaAprovado {
		return nil, fmt.Errorf("produto já foi aprovado por este trader")
	}

	// Busca a próxima ordem de exibição
	var proximaOrdem int
	err = s.db.QueryRow(`
		SELECT COALESCE(MAX(ordem_exibicao), 0) + 1
		FROM produtos_aprovados
		WHERE trader_id = $1
	`, traderID).Scan(&proximaOrdem)

	if err != nil {
		logrus.WithError(err).Error("Erro ao buscar próxima ordem")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	// Cria o produto aprovado
	produto := &models.ProdutoAprovado{
		ID:              uuid.New(),
		TraderID:        traderID,
		CavaleteID:      request.CavaleteID,
		NomeCustomizado: request.NomeCustomizado,
		PrecoVenda:      request.PrecoVenda,
		Descricao:       request.Descricao,
		Visivel:         true, // Padrão visível
		Destaque:        false, // Padrão sem destaque
		OrdemExibicao:   proximaOrdem,
	}

	// Aplica configurações opcionais
	if request.Visivel != nil {
		produto.Visivel = *request.Visivel
	}
	if request.Destaque != nil {
		produto.Destaque = *request.Destaque
	}

	query := `
		INSERT INTO produtos_aprovados (
			id, trader_id, cavalete_id, nome_customizado, preco_venda, descricao,
			visivel, destaque, ordem_exibicao, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`

	_, err = s.db.Exec(query,
		produto.ID, produto.TraderID, produto.CavaleteID, produto.NomeCustomizado,
		produto.PrecoVenda, produto.Descricao, produto.Visivel, produto.Destaque,
		produto.OrdemExibicao,
	)

	if err != nil {
		logrus.WithError(err).Error("Erro ao inserir produto aprovado")
		return nil, fmt.Errorf("erro ao aprovar produto")
	}

	logrus.WithFields(logrus.Fields{
		"trader_id":   traderID,
		"produto_id":  produto.ID,
		"cavalete_id": request.CavaleteID,
	}).Info("Produto aprovado com sucesso")

	return produto, nil
}

// ListarProdutosAprovados lista produtos aprovados do trader
func (s *ProdutosService) ListarProdutosAprovados(traderID uuid.UUID, limit, offset int) ([]models.ProdutoAprovado, error) {
	query := `
		SELECT id, trader_id, cavalete_id, nome_customizado, preco_venda, descricao,
			   visivel, destaque, ordem_exibicao, created_at, updated_at
		FROM produtos_aprovados
		WHERE trader_id = $1
		ORDER BY ordem_exibicao ASC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, traderID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Erro ao buscar produtos aprovados")
		return nil, fmt.Errorf("erro ao buscar produtos aprovados")
	}
	defer rows.Close()

	var produtos []models.ProdutoAprovado
	for rows.Next() {
		var p models.ProdutoAprovado
		err := rows.Scan(
			&p.ID, &p.TraderID, &p.CavaleteID, &p.NomeCustomizado, &p.PrecoVenda,
			&p.Descricao, &p.Visivel, &p.Destaque, &p.OrdemExibicao,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Erro ao escanear produto aprovado")
			continue
		}
		produtos = append(produtos, p)
	}

	return produtos, nil
}

// AtualizarProduto atualiza um produto aprovado
func (s *ProdutosService) AtualizarProduto(traderID, produtoID uuid.UUID, request *models.ProdutoAtualizarRequest) (*models.ProdutoAprovado, error) {
	// Verifica se o produto existe e pertence ao trader
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM produtos_aprovados
			WHERE id = $1 AND trader_id = $2
		)
	`, produtoID, traderID).Scan(&exists)

	if err != nil {
		logrus.WithError(err).Error("Erro ao verificar produto")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	if !exists {
		return nil, fmt.Errorf("produto não encontrado")
	}

	// Constrói a query de atualização dinamicamente
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if request.NomeCustomizado != nil && *request.NomeCustomizado != "" {
		setParts = append(setParts, fmt.Sprintf("nome_customizado = $%d", argIndex))
		args = append(args, *request.NomeCustomizado)
		argIndex++
	}

	if request.PrecoVenda != nil && *request.PrecoVenda > 0 {
		setParts = append(setParts, fmt.Sprintf("preco_venda = $%d", argIndex))
		args = append(args, *request.PrecoVenda)
		argIndex++
	}

	if request.Descricao != nil {
		setParts = append(setParts, fmt.Sprintf("descricao = $%d", argIndex))
		args = append(args, request.Descricao)
		argIndex++
	}

	if request.Visivel != nil {
		setParts = append(setParts, fmt.Sprintf("visivel = $%d", argIndex))
		args = append(args, *request.Visivel)
		argIndex++
	}

	if request.Destaque != nil {
		setParts = append(setParts, fmt.Sprintf("destaque = $%d", argIndex))
		args = append(args, *request.Destaque)
		argIndex++
	}

	if request.OrdemExibicao != nil {
		setParts = append(setParts, fmt.Sprintf("ordem_exibicao = $%d", argIndex))
		args = append(args, *request.OrdemExibicao)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("nenhum campo para atualizar")
	}

	// Adiciona updated_at e IDs
	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, produtoID, traderID)

	query := fmt.Sprintf(`
		UPDATE produtos_aprovados
		SET %s
		WHERE id = $%d AND trader_id = $%d
	`, strings.Join(setParts, ", "), argIndex, argIndex+1)

	_, err = s.db.Exec(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Erro ao atualizar produto")
		return nil, fmt.Errorf("erro ao atualizar produto")
	}

	// Retorna o produto atualizado
	return s.BuscarProduto(traderID, produtoID)
}

// BuscarProduto busca um produto específico do trader
func (s *ProdutosService) BuscarProduto(traderID, produtoID uuid.UUID) (*models.ProdutoAprovado, error) {
	var produto models.ProdutoAprovado

	query := `
		SELECT id, trader_id, cavalete_id, nome_customizado, preco_venda, descricao,
			   visivel, destaque, ordem_exibicao, created_at, updated_at
		FROM produtos_aprovados
		WHERE id = $1 AND trader_id = $2
	`

	err := s.db.QueryRow(query, produtoID, traderID).Scan(
		&produto.ID, &produto.TraderID, &produto.CavaleteID, &produto.NomeCustomizado,
		&produto.PrecoVenda, &produto.Descricao, &produto.Visivel, &produto.Destaque,
		&produto.OrdemExibicao, &produto.CreatedAt, &produto.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("produto não encontrado")
	} else if err != nil {
		logrus.WithError(err).Error("Erro ao buscar produto")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	return &produto, nil
}

// RemoverProduto remove um produto aprovado
func (s *ProdutosService) RemoverProduto(traderID, produtoID uuid.UUID) error {
	result, err := s.db.Exec(`
		DELETE FROM produtos_aprovados
		WHERE id = $1 AND trader_id = $2
	`, produtoID, traderID)

	if err != nil {
		logrus.WithError(err).Error("Erro ao remover produto")
		return fmt.Errorf("erro ao remover produto")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("Erro ao verificar linhas afetadas")
		return fmt.Errorf("erro interno do servidor")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("produto não encontrado")
	}

	logrus.WithFields(logrus.Fields{
		"trader_id":  traderID,
		"produto_id": produtoID,
	}).Info("Produto removido com sucesso")

	return nil
}

// ListarVitrinePublica lista produtos da vitrine pública
func (s *ProdutosService) ListarVitrinePublica(limit, offset int, destaque bool) ([]models.VitrinePublica, error) {
	query := `
		SELECT * FROM vitrine_publica
		WHERE ($3 = false OR destaque = true)
		ORDER BY 
			CASE WHEN destaque THEN ordem_exibicao ELSE 999999 END ASC,
			ordem_exibicao ASC,
			created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset, destaque)
	if err != nil {
		logrus.WithError(err).Error("Erro ao buscar vitrine pública")
		return nil, fmt.Errorf("erro ao buscar vitrine pública")
	}
	defer rows.Close()

	var produtos []models.VitrinePublica
	for rows.Next() {
		var p models.VitrinePublica
		err := rows.Scan(
			&p.ID, &p.TraderID, &p.NomeCustomizado, &p.PrecoVenda, &p.Descricao,
			&p.Destaque, &p.OrdemExibicao, &p.Codigo, &p.Bloco, &p.NomeMaterial,
			&p.NomeEspessura, &p.NomeClassificacao, &p.NomeAcabamento,
			&p.Comprimento, &p.Altura, &p.Largura, &p.Metragem, &p.Peso,
			&p.TipoMetragem, &p.ImagemPrincipal, &p.ImagensAdicionais,
			&p.TraderNome, &p.TraderEmpresa, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Erro ao escanear produto da vitrine")
			continue
		}
		produtos = append(produtos, p)
	}

	return produtos, nil
}

// ObterEstatisticas retorna estatísticas dos produtos do trader
func (s *ProdutosService) ObterEstatisticas(traderID uuid.UUID) (*models.EstatisticasProdutos, error) {
	var stats models.EstatisticasProdutos

	// Log para debug
	logrus.WithField("trader_id", traderID).Info("Buscando estatísticas para trader")

	// Query para contar produtos aprovados do trader
	queryProdutos := `SELECT COUNT(*) FROM produtos_aprovados WHERE trader_id = $1`
	
	err := s.db.QueryRow(queryProdutos, traderID).Scan(&stats.TotalProdutos)
	if err != nil {
		logrus.WithError(err).Error("Erro ao contar produtos aprovados")
		return nil, fmt.Errorf("erro ao buscar estatísticas")
	}

	// Query para contar produtos visíveis (assumindo que todos os aprovados são visíveis)
	stats.ProdutosVisiveis = stats.TotalProdutos

	// Query para contar produtos em destaque (assumindo campo destaque ou similar)
	queryDestaque := `SELECT COUNT(*) FROM produtos_aprovados WHERE trader_id = $1 AND destaque = true`
	
	err = s.db.QueryRow(queryDestaque, traderID).Scan(&stats.ProdutosDestaque)
	if err != nil {
		// Se não existe campo destaque, definir como 0
		logrus.WithError(err).Warn("Campo destaque não encontrado, definindo como 0")
		stats.ProdutosDestaque = 0
	}

	// Query para contar cavaletes disponíveis
	queryCavaletes := `
		SELECT COUNT(*) 
		FROM cavaletes c 
		WHERE c.id NOT IN (
			SELECT DISTINCT cavalete_id 
			FROM produtos_aprovados 
			WHERE cavalete_id IS NOT NULL
		)`
	
	err = s.db.QueryRow(queryCavaletes).Scan(&stats.CavaletesDisponiveis)
	if err != nil {
		logrus.WithError(err).Error("Erro ao contar cavaletes disponíveis")
		return nil, fmt.Errorf("erro ao buscar estatísticas")
	}

	logrus.WithFields(logrus.Fields{
		"total_produtos":         stats.TotalProdutos,
		"produtos_visiveis":      stats.ProdutosVisiveis,
		"produtos_destaque":      stats.ProdutosDestaque,
		"cavaletes_disponiveis":  stats.CavaletesDisponiveis,
	}).Info("Estatísticas calculadas com sucesso")

	return &stats, nil
}

// LimparTodosRegistros remove todos os registros do banco de dados
func (s *ProdutosService) LimparTodosRegistros() error {
	logrus.Info("Iniciando limpeza completa do banco de dados")

	// Lista de tabelas para limpar na ordem correta (respeitando foreign keys)
	tabelas := []string{
		"produtos_aprovados",
		"cavaletes", 
		"ofertas",
		"traders",
	}

	// Iniciar transação
	tx, err := s.db.Begin()
	if err != nil {
		logrus.WithError(err).Error("Erro ao iniciar transação para limpeza")
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Limpar cada tabela
	for _, tabela := range tabelas {
		query := fmt.Sprintf("DELETE FROM %s", tabela)
		
		result, err := tx.Exec(query)
		if err != nil {
			logrus.WithError(err).WithField("tabela", tabela).Error("Erro ao limpar tabela")
			return fmt.Errorf("erro ao limpar tabela %s: %w", tabela, err)
		}

		rowsAffected, _ := result.RowsAffected()
		logrus.WithFields(logrus.Fields{
			"tabela": tabela,
			"registros_removidos": rowsAffected,
		}).Info("Tabela limpa com sucesso")
	}

	// Resetar sequências (auto increment)
	sequencias := []string{
		"ALTER SEQUENCE traders_id_seq RESTART WITH 1",
		"ALTER SEQUENCE ofertas_id_seq RESTART WITH 1", 
		"ALTER SEQUENCE cavaletes_id_seq RESTART WITH 1",
		"ALTER SEQUENCE produtos_aprovados_id_seq RESTART WITH 1",
	}

	for _, seq := range sequencias {
		_, err := tx.Exec(seq)
		if err != nil {
			// Log do erro mas não falha a operação, pois as sequências podem não existir
			logrus.WithError(err).WithField("sequencia", seq).Warn("Erro ao resetar sequência")
		}
	}

	// Commit da transação
	if err := tx.Commit(); err != nil {
		logrus.WithError(err).Error("Erro ao fazer commit da limpeza")
		return fmt.Errorf("erro ao fazer commit: %w", err)
	}

	logrus.Info("Limpeza completa do banco de dados concluída com sucesso")
	return nil
}