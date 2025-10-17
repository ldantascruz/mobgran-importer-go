package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"mobgran-importer-go/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Client representa o cliente Supabase
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewClient cria uma nova instância do cliente Supabase
func NewClient(url, key string, logger *logrus.Logger) (*Client, error) {
	// Configuração mais robusta do transporte HTTP com fallback de DNS
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		// Configurações de DNS mais robustas com fallback para IPs diretos
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Se for pflcrfnkfzzfamchqcav.supabase.co, usar IPs diretos como fallback
			if strings.Contains(addr, "pflcrfnkfzzfamchqcav.supabase.co") {
				// Tentar primeiro IP
				conn, err := (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext(ctx, network, strings.Replace(addr, "pflcrfnkfzzfamchqcav.supabase.co", "104.18.38.10", 1))

				if err != nil {
					// Fallback para segundo IP
					conn, err = (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext(ctx, network, strings.Replace(addr, "pflcrfnkfzzfamchqcav.supabase.co", "172.64.149.246", 1))
				}

				return conn, err
			}

			// Para outros hosts, usar dialer padrão
			return (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext(ctx, network, addr)
		},
		// Configurações de TLS
		TLSHandshakeTimeout: 10 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	return &Client{
		baseURL:    url + "/rest/v1",
		apiKey:     key,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// makeRequest faz uma requisição HTTP para o Supabase
func (c *Client) makeRequest(method, endpoint string, body interface{}, result interface{}) error {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("erro ao serializar body: %w", err)
		}
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao fazer requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// Ler o corpo da resposta para obter detalhes do erro
		bodyBytes, _ := json.Marshal(body)
		c.logger.WithFields(logrus.Fields{
			"status_code":  resp.StatusCode,
			"endpoint":     endpoint,
			"method":       method,
			"request_body": string(bodyBytes),
		}).Error("Erro HTTP detalhado")

		return fmt.Errorf("erro HTTP %d", resp.StatusCode)
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		// Log da resposta para debug
		respBody := make([]byte, 0)
		if resp.Body != nil {
			respBody, _ = json.Marshal(resp.Body)
		}
		c.logger.WithFields(logrus.Fields{
			"status_code":   resp.StatusCode,
			"response_body": string(respBody),
		}).Debug("Resposta recebida")

		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("erro ao decodificar resposta: %w", err)
		}
	}

	return nil
}

// VerificarOfertaExistente verifica se uma oferta já existe pelo UUID
func (c *Client) VerificarOfertaExistente(ofertaUUID string) (*string, error) {
	c.logger.WithField("uuid", ofertaUUID).Info("Verificando se oferta já existe")

	var ofertas []models.Oferta
	endpoint := fmt.Sprintf("/ofertas?uuid_link=eq.%s&select=id", ofertaUUID)

	err := c.makeRequest("GET", endpoint, nil, &ofertas)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao verificar oferta existente")
		return nil, fmt.Errorf("erro ao verificar oferta existente: %w", err)
	}

	if len(ofertas) > 0 {
		c.logger.WithField("oferta_id", ofertas[0].ID).Info("Oferta já existe")
		return &ofertas[0].ID, nil
	}

	c.logger.Info("Oferta não existe")
	return nil, nil
}

// SalvarOferta salva uma nova oferta no banco de dados
func (c *Client) SalvarOferta(ofertaUUID string, dados *models.MobgranResponse) (*string, error) {
	c.logger.WithField("uuid", ofertaUUID).Info("Salvando nova oferta")

	// Converter dados originais para JSON
	dadosOriginaisJSON, err := json.Marshal(dados)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar dados originais: %w", err)
	}

	var dadosOriginaisMap map[string]interface{}
	if err = json.Unmarshal(dadosOriginaisJSON, &dadosOriginaisMap); err != nil {
		return nil, fmt.Errorf("erro ao converter dados originais: %w", err)
	}

	oferta := models.Oferta{
		ID:             uuid.New().String(),
		UUIDLink:       ofertaUUID,
		Situacao:       dados.Situacao,
		NomeEmpresa:    dados.NomeEmpresa,
		URLLogo:        dados.URLLogo,
		DadosCompletos: dadosOriginaisMap,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	var resultado []models.Oferta
	err = c.makeRequest("POST", "/ofertas", oferta, &resultado)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao salvar oferta")
		return nil, fmt.Errorf("erro ao salvar oferta: %w", err)
	}

	c.logger.WithField("oferta_id", oferta.ID).Info("Oferta salva com sucesso")
	return &oferta.ID, nil
}

// SalvarCavalete salva um cavalete no banco de dados
func (c *Client) SalvarCavalete(ofertaID string, cavalete *models.Cavalete) (*string, error) {
	c.logger.WithFields(logrus.Fields{
		"oferta_id":      ofertaID,
		"nome_material":  cavalete.NomeMaterial,
		"nome_espessura": cavalete.NomeEspessura,
	}).Info("Salvando cavalete")

	// Obter classificação do primeiro item (todos os itens de um cavalete têm a mesma classificação)
	nomeClassificacao := ""
	if len(cavalete.Itens) > 0 {
		nomeClassificacao = cavalete.Itens[0].NomeClassificacao
	}

	cavaleteDB := models.CavaleteDB{
		ID:                    uuid.New().String(),
		OfertaID:              ofertaID,
		Codigo:                cavalete.Codigo,
		Bloco:                 cavalete.Bloco,
		NomeMaterial:          cavalete.NomeMaterial,
		NomeEspessura:         cavalete.NomeEspessura,
		NomeClassificacao:     nomeClassificacao,
		NomeAcabamento:        nil, // Pode ser nil
		Comprimento:           &cavalete.Comprimento,
		Altura:                &cavalete.Altura,
		Largura:               nil, // Não disponível no Mobgran
		Metragem:              &cavalete.Metragem,
		Peso:                  nil, // Não disponível no Mobgran
		TipoMetragem:          nil, // Não disponível no Mobgran
		Aprovado:              false,
		Importado:             true,
		DescricaoChapas:       nil, // Pode ser nil
		QuantidadeItens:       func() *int { count := len(cavalete.Itens); return &count }(),
		Valor:                 nil, // Não disponível no Mobgran
		Observacao:            nil, // Pode ser nil
		ObservacaoConferencia: nil, // Pode ser nil
		ProdutoCliente:        nil, // Pode ser nil
		EspessuraCliente:      nil, // Pode ser nil
		ImagemPrincipal:       make(map[string]interface{}),
		ImagensAdicionais:     make(map[string]interface{}),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Adicionar informações de imagem se disponíveis
	if cavalete.ImagemPrincipal != nil {
		cavaleteDB.ImagemPrincipal = map[string]interface{}{
			"nome":   cavalete.ImagemPrincipal.Nome,
			"url":    cavalete.ImagemPrincipal.URL,
			"urlMin": cavalete.ImagemPrincipal.URLMin,
		}
	}

	// Para POST no Supabase, não esperamos uma resposta com dados, apenas status 201
	err := c.makeRequest("POST", "/cavaletes", cavaleteDB, nil)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao salvar cavalete")
		return nil, fmt.Errorf("erro ao salvar cavalete: %w", err)
	}

	c.logger.WithField("cavalete_id", cavaleteDB.ID).Info("Cavalete salvo com sucesso")
	return &cavaleteDB.ID, nil
}

// SalvarItem salva um item no banco de dados
func (c *Client) SalvarItem(cavaleteID string, item *models.Item) error {
	c.logger.WithFields(logrus.Fields{
		"cavalete_id":        cavaleteID,
		"nome_espessura":     item.NomeEspessura,
		"nome_classificacao": item.NomeClassificacao,
	}).Info("Salvando item")

	itemDB := models.ItemDB{
		ID:                    uuid.New().String(),
		CavaleteID:            cavaleteID,
		Codigo:                item.Codigo,
		Bloco:                 item.Bloco,
		NomeEspessura:         item.NomeEspessura,
		NomeClassificacao:     item.NomeClassificacao,
		NomeAcabamento:        nil, // Pode ser nil
		Comprimento:           &item.Comprimento,
		Altura:                &item.Altura,
		Largura:               nil, // Não disponível no Mobgran
		Metragem:              &item.Metragem,
		Peso:                  nil, // Não disponível no Mobgran
		TipoMetragem:          nil, // Não disponível no Mobgran
		Aprovado:              false,
		Importado:             true,
		Valor:                 nil, // Não disponível no Mobgran
		Observacao:            nil, // Pode ser nil
		ObservacaoConferencia: nil, // Pode ser nil
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Para POST no Supabase, não esperamos uma resposta com dados, apenas status 201
	err := c.makeRequest("POST", "/itens", itemDB, nil)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao salvar item")
		return fmt.Errorf("erro ao salvar item: %w", err)
	}

	c.logger.WithField("item_id", itemDB.ID).Info("Item salvo com sucesso")
	return nil
}

// AtualizarOferta atualiza uma oferta existente
func (c *Client) AtualizarOferta(ofertaID string, dados *models.MobgranResponse) error {
	c.logger.WithField("oferta_id", ofertaID).Info("Atualizando oferta existente")

	// Converter dados originais para JSON
	dadosOriginaisJSON, err := json.Marshal(dados)
	if err != nil {
		return fmt.Errorf("erro ao serializar dados originais: %w", err)
	}

	var dadosOriginaisMap map[string]interface{}
	if err = json.Unmarshal(dadosOriginaisJSON, &dadosOriginaisMap); err != nil {
		return fmt.Errorf("erro ao converter dados originais: %w", err)
	}

	updates := map[string]interface{}{
		"situacao":        dados.Situacao,
		"nome_empresa":    dados.NomeEmpresa,
		"url_logo":        dados.URLLogo,
		"dados_completos": dadosOriginaisMap,
		"updated_at":      time.Now(),
	}

	endpoint := fmt.Sprintf("/ofertas?id=eq.%s", ofertaID)
	var resultado []models.Oferta
	err = c.makeRequest("PATCH", endpoint, updates, &resultado)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao atualizar oferta")
		return fmt.Errorf("erro ao atualizar oferta: %w", err)
	}

	c.logger.Info("Oferta atualizada com sucesso")
	return nil
}

// RemoverCavaletesEItens remove todos os cavaletes e itens de uma oferta
func (c *Client) RemoverCavaletesEItens(ofertaID string) error {
	c.logger.WithField("oferta_id", ofertaID).Info("Removendo cavaletes e itens da oferta")

	// Primeiro, buscar todos os cavaletes desta oferta
	var cavaletes []models.CavaleteDB
	endpoint := fmt.Sprintf("/cavaletes?oferta_id=eq.%s&select=id", ofertaID)
	err := c.makeRequest("GET", endpoint, nil, &cavaletes)
	if err != nil {
		return fmt.Errorf("erro ao buscar cavaletes: %w", err)
	}

	// Remover todos os itens de cada cavalete
	for _, cavalete := range cavaletes {
		endpoint = fmt.Sprintf("/itens?cavalete_id=eq.%s", cavalete.ID)
		err = c.makeRequest("DELETE", endpoint, nil, nil)
		if err != nil {
			c.logger.WithError(err).Error("Erro ao remover itens")
			return fmt.Errorf("erro ao remover itens: %w", err)
		}
	}

	// Depois, remover todos os cavaletes desta oferta
	endpoint = fmt.Sprintf("/cavaletes?oferta_id=eq.%s", ofertaID)
	err = c.makeRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		c.logger.WithError(err).Error("Erro ao remover cavaletes")
		return fmt.Errorf("erro ao remover cavaletes: %w", err)
	}

	c.logger.Info("Cavaletes e itens removidos com sucesso")
	return nil
}
