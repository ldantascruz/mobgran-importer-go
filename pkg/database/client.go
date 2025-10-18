package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/sirupsen/logrus"
	"mobgran-importer-go/internal/models"
)

// Client representa o cliente PostgreSQL
type Client struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewClient cria uma nova instância do cliente PostgreSQL
func NewClient(host, port, dbname, user, password, sslmode string, logger *logrus.Logger) (*Client, error) {
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, dbname, user, password, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com PostgreSQL: %w", err)
	}

	// Configurar pool de conexões
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Testar conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao testar conexão PostgreSQL: %w", err)
	}

	logger.Info("Conectado ao PostgreSQL com sucesso")

	return &Client{
		db:     db,
		logger: logger,
	}, nil
}

// Close fecha a conexão com o banco
func (c *Client) Close() error {
	return c.db.Close()
}

// GetDB retorna a instância do banco de dados
func (c *Client) GetDB() *sql.DB {
	return c.db
}

// VerificarOfertaExistente verifica se uma oferta já existe pelo UUID
func (c *Client) VerificarOfertaExistente(ofertaUUID string) (*string, error) {
	var id string
	query := "SELECT id FROM ofertas WHERE uuid_link = $1"
	
	err := c.db.QueryRow(query, ofertaUUID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Oferta não existe
		}
		c.logger.WithError(err).Error("Erro ao verificar oferta existente")
		return nil, err
	}

	return &id, nil
}

// SalvarOferta salva uma nova oferta no banco
func (c *Client) SalvarOferta(ofertaUUID string, dados *models.MobgranResponse) (*string, error) {
	// Serializar dados completos para JSON
	dadosJSON, err := json.Marshal(dados)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar dados: %w", err)
	}

	id := uuid.New().String()
	query := `
		INSERT INTO ofertas (id, uuid_link, situacao, nome_empresa, url_logo, dados_completos)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err = c.db.QueryRow(query, id, ofertaUUID, dados.Situacao, dados.NomeEmpresa, dados.URLLogo, dadosJSON).Scan(&id)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao salvar oferta")
		return nil, err
	}

	c.logger.WithField("oferta_id", id).Info("Oferta salva com sucesso")
	return &id, nil
}

// SalvarCavalete salva um cavalete no banco
func (c *Client) SalvarCavalete(ofertaID string, cavalete *models.Cavalete) (*string, error) {
	// Log detalhado do cavalete recebido
	c.logger.WithFields(logrus.Fields{
		"cavalete_codigo": cavalete.Codigo,
		"imagem_principal_ptr": fmt.Sprintf("%p", cavalete.ImagemPrincipal),
		"imagem_principal_nil": cavalete.ImagemPrincipal == nil,
	}).Debug("Cavalete recebido para salvamento")

	// Se ImagemPrincipal não é nil, vamos ver seus valores
	if cavalete.ImagemPrincipal != nil {
		c.logger.WithFields(logrus.Fields{
			"nome": cavalete.ImagemPrincipal.Nome,
			"url": cavalete.ImagemPrincipal.URL,
			"url_min": cavalete.ImagemPrincipal.URLMin,
		}).Debug("Valores de ImagemPrincipal")
	}

	// Serializar imagem principal para JSON ou usar NULL
	var imagemPrincipalJSON sql.NullString
	if cavalete.ImagemPrincipal != nil && 
		(cavalete.ImagemPrincipal.Nome != "" || cavalete.ImagemPrincipal.URL != "" || cavalete.ImagemPrincipal.URLMin != "") {
		// Para JSONB, usar sql.NullString para garantir que NULL seja passado corretamente
		jsonBytes, err := json.Marshal(cavalete.ImagemPrincipal)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar imagem principal: %w", err)
		}
		imagemPrincipalJSON = sql.NullString{String: string(jsonBytes), Valid: true}
		c.logger.WithField("imagem_principal", string(jsonBytes)).Debug("Imagem principal definida com dados válidos")
	} else {
		imagemPrincipalJSON = sql.NullString{Valid: false} // NULL no PostgreSQL
		c.logger.Debug("Imagem principal é nil ou vazia, usando NULL")
	}

	id := uuid.New().String()
	query := `
		INSERT INTO cavaletes (
			id, oferta_id, codigo, bloco, nome_material, nome_espessura,
			comprimento, altura, metragem, imagem_principal, quantidade_itens
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	c.logger.WithFields(logrus.Fields{
		"cavalete_codigo": cavalete.Codigo,
		"imagem_principal_type": fmt.Sprintf("%T", imagemPrincipalJSON),
		"imagem_principal_value": imagemPrincipalJSON,
		"imagem_principal_is_valid": imagemPrincipalJSON.Valid,
	}).Debug("Executando query de inserção")

	err := c.db.QueryRow(query,
		id, ofertaID, cavalete.Codigo, cavalete.Bloco, cavalete.NomeMaterial,
		cavalete.NomeEspessura, cavalete.Comprimento, cavalete.Altura,
		cavalete.Metragem, imagemPrincipalJSON, len(cavalete.Itens),
	).Scan(&id)

	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"cavalete_codigo": cavalete.Codigo,
			"query_params": fmt.Sprintf("id=%s, ofertaID=%s, codigo=%s, bloco=%s, nomeMaterial=%s, nomeEspessura=%s, comprimento=%f, altura=%f, metragem=%f, imagemPrincipal=%v, quantidadeItens=%d",
				id, ofertaID, cavalete.Codigo, cavalete.Bloco, cavalete.NomeMaterial,
				cavalete.NomeEspessura, cavalete.Comprimento, cavalete.Altura,
				cavalete.Metragem, imagemPrincipalJSON, len(cavalete.Itens)),
		}).Error("Erro ao salvar cavalete")
		return nil, err
	}

	c.logger.WithField("cavalete_id", id).Info("Cavalete salvo com sucesso")
	return &id, nil
}

// SalvarItem salva um item no banco
func (c *Client) SalvarItem(cavaleteID string, item *models.Item) error {
	id := uuid.New().String()
	query := `
		INSERT INTO itens (
			id, cavalete_id, codigo, bloco, nome_espessura, nome_classificacao,
			comprimento, altura, metragem
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := c.db.Exec(query,
		id, cavaleteID, item.Codigo, item.Bloco, item.NomeEspessura,
		item.NomeClassificacao, item.Comprimento, item.Altura, item.Metragem,
	)

	if err != nil {
		c.logger.WithError(err).Error("Erro ao salvar item")
		return err
	}

	c.logger.WithField("item_id", id).Info("Item salvo com sucesso")
	return nil
}

// AtualizarOferta atualiza uma oferta existente
func (c *Client) AtualizarOferta(ofertaID string, dados *models.MobgranResponse) error {
	// Serializar dados completos para JSON
	dadosJSON, err := json.Marshal(dados)
	if err != nil {
		return fmt.Errorf("erro ao serializar dados: %w", err)
	}

	query := `
		UPDATE ofertas 
		SET situacao = $2, nome_empresa = $3, url_logo = $4, dados_completos = $5, updated_at = NOW()
		WHERE id = $1`

	result, err := c.db.Exec(query, ofertaID, dados.Situacao, dados.NomeEmpresa, dados.URLLogo, dadosJSON)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao atualizar oferta")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("nenhuma oferta encontrada com ID: %s", ofertaID)
	}

	c.logger.WithField("oferta_id", ofertaID).Info("Oferta atualizada com sucesso")
	return nil
}

// RemoverCavaletesEItens remove todos os cavaletes e itens de uma oferta
func (c *Client) RemoverCavaletesEItens(ofertaID string) error {
	// Iniciar transação
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Remover itens (CASCADE vai cuidar disso, mas vamos ser explícitos)
	_, err = tx.Exec("DELETE FROM itens WHERE cavalete_id IN (SELECT id FROM cavaletes WHERE oferta_id = $1)", ofertaID)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao remover itens")
		return err
	}

	// Remover cavaletes
	_, err = tx.Exec("DELETE FROM cavaletes WHERE oferta_id = $1", ofertaID)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao remover cavaletes")
		return err
	}

	// Commit da transação
	if err = tx.Commit(); err != nil {
		c.logger.WithError(err).Error("Erro ao fazer commit da transação")
		return err
	}

	c.logger.WithField("oferta_id", ofertaID).Info("Cavaletes e itens removidos com sucesso")
	return nil
}