package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"mobgran-importer-go/internal/models"
	"mobgran-importer-go/pkg/database"
)

// MobgranImporter representa o serviço de importação do Mobgran
type MobgranImporter struct {
	dbClient   *database.Client
	httpClient *http.Client
	logger     *logrus.Logger
	apiBaseURL string
}

// NewMobgranImporter cria uma nova instância do importador
func NewMobgranImporter(dbClient *database.Client, logger *logrus.Logger) *MobgranImporter {
	// Cliente HTTP simples e padrão
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	return &MobgranImporter{
		dbClient:   dbClient,
		httpClient: client,
		logger:     logger,
		apiBaseURL: "https://www.mobgran.com/app/api/link-produto",
	}
}

// ExtrairUUIDLink extrai o UUID do link mobgran
func (m *MobgranImporter) ExtrairUUIDLink(url string) (*string, error) {
	m.logger.WithField("url", url).Info("Extraindo UUID do link")

	// Padrão regex para extrair UUID do link mobgran
	// Exemplo: https://www.mobgran.com/app/conferencia/?p=link&o=cae15fe7-86a3-4a7b-9a4d-5ed91ae6d568/
	pattern := `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	re := regexp.MustCompile(pattern)

	match := re.FindString(url)
	if match == "" {
		m.logger.WithField("url", url).Error("UUID não encontrado no link")
		return nil, fmt.Errorf("UUID não encontrado no link: %s", url)
	}

	m.logger.WithField("uuid", match).Info("UUID extraído com sucesso")
	return &match, nil
}

// BuscarDadosAPI busca os dados da API do Mobgran
func (m *MobgranImporter) BuscarDadosAPI(uuid string) (*models.MobgranResponse, error) {
	m.logger.WithField("uuid", uuid).Info("Buscando dados da API Mobgran")

	url := fmt.Sprintf("%s/%s", m.apiBaseURL, uuid)
	m.logger.WithField("url_completa", url).Info("URL da API construída")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://www.mobgran.com/")
	req.Header.Set("Origin", "https://www.mobgran.com")

	m.logger.WithFields(logrus.Fields{
		"method":     req.Method,
		"url":        req.URL.String(),
		"headers":    req.Header,
	}).Info("Fazendo requisição HTTP")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.logger.WithError(err).Error("Erro ao fazer requisição para API")
		return nil, fmt.Errorf("erro ao fazer requisição para API: %w", err)
	}
	defer resp.Body.Close()

	m.logger.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
	}).Info("Resposta recebida da API")

	if resp.StatusCode != http.StatusOK {
		// Ler o corpo da resposta para debug
		body, _ := io.ReadAll(resp.Body)
		m.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"body":        string(body),
		}).Error("API retornou erro")
		return nil, fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	var dados models.MobgranResponse
	if err := json.NewDecoder(resp.Body).Decode(&dados); err != nil {
		m.logger.WithError(err).Error("Erro ao decodificar resposta da API")
		return nil, fmt.Errorf("erro ao decodificar resposta da API: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"situacao":      dados.Situacao,
		"nome_empresa":  dados.NomeEmpresa,
		"num_cavaletes": len(dados.Cavaletes),
	}).Info("Dados da API obtidos com sucesso")

	return &dados, nil
}

// Importar executa o processo completo de importação
func (m *MobgranImporter) Importar(url string, atualizarExistente bool) (bool, string, *string, error) {
	m.logger.WithField("url", url).Info("Iniciando importação")

	// Validar URL
	if err := m.ValidarURL(url); err != nil {
		return false, "URL inválida", nil, err
	}

	// Extrair UUID do link
	uuid, err := m.ExtrairUUIDLink(url)
	if err != nil {
		return false, "Erro ao extrair UUID do link", nil, err
	}

	// Verificar se a oferta já existe
	ofertaExistente, err := m.dbClient.VerificarOfertaExistente(*uuid)
	if err != nil {
		return false, "Erro ao verificar oferta existente", nil, err
	}

	// Buscar dados da API
	dados, err := m.BuscarDadosAPI(*uuid)
	if err != nil {
		return false, "Erro ao buscar dados da API", nil, err
	}

	var ofertaID string

	if ofertaExistente != nil {
		// Oferta já existe
		if !atualizarExistente {
			return false, "Oferta já existe e atualização não foi solicitada", ofertaExistente, nil
		}

		// Atualizar oferta existente
		if err := m.dbClient.AtualizarOferta(*ofertaExistente, dados); err != nil {
			return false, "Erro ao atualizar oferta", ofertaExistente, err
		}

		// Remover cavaletes e itens antigos
		if err := m.dbClient.RemoverCavaletesEItens(*ofertaExistente); err != nil {
			return false, "Erro ao remover cavaletes e itens antigos", ofertaExistente, err
		}

		ofertaID = *ofertaExistente
		m.logger.WithField("oferta_id", ofertaID).Info("Oferta atualizada com sucesso")
	} else {
		// Criar nova oferta
		novoOfertaID, err := m.dbClient.SalvarOferta(*uuid, dados)
		if err != nil {
			return false, "Erro ao salvar nova oferta", nil, err
		}
		ofertaID = *novoOfertaID
		m.logger.WithField("oferta_id", ofertaID).Info("Nova oferta criada com sucesso")
	}

	// Salvar cavaletes e itens
	if err := m.salvarCavaletesEItens(ofertaID, dados.Cavaletes); err != nil {
		return false, "Erro ao salvar cavaletes e itens", &ofertaID, err
	}

	return true, "Importação realizada com sucesso", &ofertaID, nil
}

// salvarCavaletesEItens salva os cavaletes e seus itens
func (m *MobgranImporter) salvarCavaletesEItens(ofertaID string, cavaletes []models.Cavalete) error {
	m.logger.WithField("oferta_id", ofertaID).WithField("total_cavaletes", len(cavaletes)).Info("Salvando cavaletes e itens")

	for i, cavalete := range cavaletes {
		m.logger.WithField("cavalete_index", i).WithField("codigo", cavalete.Codigo).Info("Processando cavalete")

		// Salvar cavalete
		cavaleteID, err := m.dbClient.SalvarCavalete(ofertaID, &cavalete)
		if err != nil {
			m.logger.WithError(err).WithField("cavalete_codigo", cavalete.Codigo).Error("Erro ao salvar cavalete")
			return fmt.Errorf("erro ao salvar cavalete %s: %w", cavalete.Codigo, err)
		}

		// Salvar itens do cavalete
		for j, item := range cavalete.Itens {
			m.logger.WithField("item_index", j).WithField("codigo", item.Codigo).Info("Processando item")

			if err := m.dbClient.SalvarItem(*cavaleteID, &item); err != nil {
				m.logger.WithError(err).WithField("item_codigo", item.Codigo).Error("Erro ao salvar item")
				return fmt.Errorf("erro ao salvar item %s do cavalete %s: %w", item.Codigo, cavalete.Codigo, err)
			}
		}

		m.logger.WithField("cavalete_id", *cavaleteID).WithField("total_itens", len(cavalete.Itens)).Info("Cavalete e itens salvos com sucesso")
	}

	return nil
}

// ValidarURL valida se a URL é um link válido do Mobgran
func (m *MobgranImporter) ValidarURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL não pode estar vazia")
	}

	if !strings.Contains(url, "mobgran.com") {
		return fmt.Errorf("URL deve ser do domínio mobgran.com")
	}

	// Tentar extrair UUID para validar formato
	_, err := m.ExtrairUUIDLink(url)
	if err != nil {
		return fmt.Errorf("URL não contém um UUID válido: %w", err)
	}

	return nil
}
