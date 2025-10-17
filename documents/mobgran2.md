#docker-compose.yml
version: '3.8'

services:
  mobgran-importer:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: mobgran-importer
    restart: unless-stopped
    ports:
      - "${PORT:-8080}:8080"
    environment:
      - SUPABASE_URL=${SUPABASE_URL}
      - SUPABASE_KEY=${SUPABASE_KEY}
      - PORT=${PORT:-8080}
      - LOG_LEVEL=${LOG_LEVEL:-info}
    env_file:
      - .env
    volumes:
      - ./logs:/app/logs  # Para persistir logs (opcional)
    networks:
      - mobgran-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

networks:
  mobgran-network:
    driver: bridge

volumes:
  logs:
    driver: local




#Dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

# Instalar dependências do sistema
RUN apk add --no-cache git

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mobgran-importer .

# Final stage
FROM alpine:latest

# Instalar certificados SSL
RUN apk --no-cache add ca-certificates

# Criar usuário não-root
RUN addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser

# Definir diretório de trabalho
WORKDIR /app

# Copiar binário do build stage
COPY --from=builder /app/mobgran-importer .

# Copiar arquivo de exemplo de configuração (opcional)
COPY --from=builder /app/.env.example .

# Dar permissões ao usuário
RUN chown -R appuser:appuser /app

# Mudar para usuário não-root
USER appuser

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Comando para executar a aplicação
CMD ["./mobgran-importer"]


#.env.example
# Configuração do Supabase
SUPABASE_URL=https://seu-projeto.supabase.co
SUPABASE_KEY=sua-chave-api-aqui

# Configuração do Servidor (opcional)
PORT=8080

# Configuração de Log (opcional)
LOG_LEVEL=info

# Configurações de Rate Limiting (opcional)
MAX_REQUESTS_PER_MINUTE=60
MAX_BATCH_SIZE=50


#go.mod
module mobgran-importer

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/bytedance/sonic v1.10.2 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.16.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.6.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)



#utils/supabase.go
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SupabaseClient é um cliente para interagir com a API do Supabase
type SupabaseClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewSupabaseClient cria um novo cliente Supabase
func NewSupabaseClient(baseURL, apiKey string) *SupabaseClient {
	return &SupabaseClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest faz uma requisição HTTP genérica ao Supabase
func (s *SupabaseClient) makeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s", s.BaseURL, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	// Headers padrão do Supabase
	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	// Verificar status da resposta
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get busca registros de uma tabela
func (s *SupabaseClient) Get(table, query string) ([]byte, error) {
	endpoint := table
	if query != "" {
		endpoint = fmt.Sprintf("%s?%s", table, query)
	}
	return s.makeRequest("GET", endpoint, nil)
}

// GetWithQuery busca registros com query string customizada
func (s *SupabaseClient) GetWithQuery(table, queryString string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s?%s", table, queryString)
	return s.makeRequest("GET", endpoint, nil)
}

// Insert insere um novo registro em uma tabela
func (s *SupabaseClient) Insert(table string, data interface{}) error {
	_, err := s.makeRequest("POST", table, data)
	return err
}

// Update atualiza um registro existente
func (s *SupabaseClient) Update(table, uuid string, data interface{}) error {
	endpoint := fmt.Sprintf("%s?uuid=eq.%s", table, uuid)
	_, err := s.makeRequest("PATCH", endpoint, data)
	return err
}

// Delete remove registros de uma tabela
func (s *SupabaseClient) Delete(table, query string) error {
	endpoint := fmt.Sprintf("%s?%s", table, query)
	_, err := s.makeRequest("DELETE", endpoint, nil)
	return err
}

// Upsert insere ou atualiza um registro
func (s *SupabaseClient) Upsert(table string, data interface{}) error {
	url := fmt.Sprintf("%s/rest/v1/%s", s.BaseURL, table)

	jsonBody, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("erro ao serializar dados: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %v", err)
	}

	// Headers para upsert
	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RPC chama uma função RPC no Supabase
func (s *SupabaseClient) RPC(functionName string, params interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/rpc/%s", s.BaseURL, functionName)

	jsonBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar parâmetros: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}


#models/models.go
package models

import (
	"time"
)

// ImportRequest representa uma requisição de importação
type ImportRequest struct {
	URL                string `json:"url" binding:"required"`
	AtualizarExistente *bool  `json:"atualizar_existente,omitempty"`
}

// BatchImportRequest representa uma requisição de importação em lote
type BatchImportRequest struct {
	URLs               []string `json:"urls" binding:"required,min=1,max=50"`
	AtualizarExistente *bool    `json:"atualizar_existente,omitempty"`
}

// ImportResult representa o resultado de uma importação
type ImportResult struct {
	UUID               string        `json:"uuid"`
	URL                string        `json:"url"`
	Sucesso            bool          `json:"sucesso"`
	Mensagem           string        `json:"mensagem"`
	Erro               string        `json:"erro,omitempty"`
	TotalCavaletes     int           `json:"total_cavaletes"`
	TotalItens         int           `json:"total_itens"`
	MetragemTotal      float64       `json:"metragem_total"`
	TempoProcessamento time.Duration `json:"tempo_processamento"`
	Timestamp          time.Time     `json:"timestamp"`
}

// DadosMobgran representa os dados retornados pela API do Mobgran
type DadosMobgran struct {
	Empresa          string              `json:"empresa"`
	Vendedor         string              `json:"vendedor"`
	Cliente          string              `json:"cliente"`
	Observacoes      string              `json:"observacoes"`
	DataConferencia  string              `json:"dataConferencia"`
	Cavaletes        []CavaleteMobgran   `json:"cavaletes"`
	Configuracoes    map[string]interface{} `json:"configuracoes,omitempty"`
}

// CavaleteMobgran representa um cavalete nos dados do Mobgran
type CavaleteMobgran struct {
	CodigoNumerico   string         `json:"codigoNumerico"`
	NumeroBloco      string         `json:"numeroBloco"`
	NomeMaterial     string         `json:"nomeMaterial"`
	NomeEspessura    string         `json:"nomeEspessura"`
	TipoProduto      string         `json:"tipoProduto"`
	MedidaA          float64        `json:"medidaA"`
	MedidaB          float64        `json:"medidaB"`
	MedidaC          float64        `json:"medidaC"`
	Volume           float64        `json:"volume"`
	Peso             float64        `json:"peso"`
	Metragem         float64        `json:"metragem"`
	QuantidadeChapas int            `json:"quantidadeChapas"`
	Imagem           string         `json:"imagem"`
	Classificacao    string         `json:"classificacao"`
	Observacoes      string         `json:"observacoes"`
	Itens            []ItemMobgran  `json:"itens"`
}

// ItemMobgran representa um item (chapa) do cavalete
type ItemMobgran struct {
	Codigo     string  `json:"codigo"`
	MedidaA    float64 `json:"medidaA"`
	MedidaB    float64 `json:"medidaB"`
	Espessura  float64 `json:"espessura"`
	Metragem   float64 `json:"metragem"`
	Observacao string  `json:"observacao"`
}

// Oferta representa uma oferta no banco de dados
type Oferta struct {
	ID             int64         `json:"id,omitempty" db:"id"`
	UUID           string        `json:"uuid" db:"uuid"`
	NomeEmpresa    string        `json:"nome_empresa" db:"nome_empresa"`
	NomeVendedor   string        `json:"nome_vendedor" db:"nome_vendedor"`
	DadosCompletos *DadosMobgran `json:"dados_completos" db:"dados_completos"`
	DataImportacao time.Time     `json:"data_importacao" db:"data_importacao"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
}

// Cavalete representa um cavalete no banco de dados
type Cavalete struct {
	ID               int64            `json:"id,omitempty" db:"id"`
	OfertaUUID       string           `json:"oferta_uuid" db:"oferta_uuid"`
	CodigoNumerico   string           `json:"codigo_numerico" db:"codigo_numerico"`
	NumeroBloco      string           `json:"numero_bloco" db:"numero_bloco"`
	NomeMaterial     string           `json:"nome_material" db:"nome_material"`
	NomeEspessura    string           `json:"nome_espessura" db:"nome_espessura"`
	TipoProduto      string           `json:"tipo_produto" db:"tipo_produto"`
	MedidaA          float64          `json:"medida_a" db:"medida_a"`
	MedidaB          float64          `json:"medida_b" db:"medida_b"`
	MedidaC          float64          `json:"medida_c" db:"medida_c"`
	Volume           float64          `json:"volume" db:"volume"`
	Peso             float64          `json:"peso" db:"peso"`
	Metragem         float64          `json:"metragem" db:"metragem"`
	QuantidadeChapas int              `json:"quantidade_chapas" db:"quantidade_chapas"`
	Imagem           string           `json:"imagem" db:"imagem"`
	DadosCompletos   *CavaleteMobgran `json:"dados_completos" db:"dados_completos"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
}

// Item representa um item (chapa) no banco de dados
type Item struct {
	ID             int64        `json:"id,omitempty" db:"id"`
	OfertaUUID     string       `json:"oferta_uuid" db:"oferta_uuid"`
	CavaleteID     string       `json:"cavalete_id" db:"cavalete_id"`
	Codigo         string       `json:"codigo" db:"codigo"`
	MedidaA        float64      `json:"medida_a" db:"medida_a"`
	MedidaB        float64      `json:"medida_b" db:"medida_b"`
	Espessura      float64      `json:"espessura" db:"espessura"`
	DadosCompletos *ItemMobgran `json:"dados_completos" db:"dados_completos"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
}

// Estatisticas representa estatísticas de importação
type Estatisticas struct {
	TotalCavaletes int            `json:"total_cavaletes"`
	TotalItens     int            `json:"total_itens"`
	MetragemTotal  float64        `json:"metragem_total"`
	PorMaterial    map[string]int `json:"por_material"`
}

// EstatisticasGlobais representa estatísticas do sistema
type EstatisticasGlobais struct {
	TotalOfertas      int            `json:"total_ofertas"`
	TotalCavaletes    int            `json:"total_cavaletes"`
	MetragemTotal     float64        `json:"metragem_total"`
	MaterialMaisComum map[string]int `json:"material_mais_comum"`
	MaterialTop       string         `json:"material_top"`
}


#handlers/mobgran_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mobgran-importer/models"
	"mobgran-importer/services"
)

// MobgranHandler gerencia as requisições HTTP relacionadas ao Mobgran
type MobgranHandler struct {
	importer *services.MobgranImporter
}

// NewMobgranHandler cria um novo handler
func NewMobgranHandler(importer *services.MobgranImporter) *MobgranHandler {
	return &MobgranHandler{
		importer: importer,
	}
}

// ImportarURL processa uma requisição para importar uma URL
func (h *MobgranHandler) ImportarURL(c *gin.Context) {
	var req models.ImportRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "Requisição inválida",
			"detalhes": err.Error(),
		})
		return
	}

	// Validar URL
	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "URL é obrigatória",
		})
		return
	}

	// Valor padrão para atualizar
	if req.AtualizarExistente == nil {
		defaultValue := true
		req.AtualizarExistente = &defaultValue
	}

	// Executar importação
	result, err := h.importer.ImportarURL(req.URL, *req.AtualizarExistente)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"erro": "Erro ao importar dados",
			"detalhes": err.Error(),
			"resultado": result,
		})
		return
	}

	// Resposta de sucesso
	statusCode := http.StatusOK
	if result.Sucesso && result.TotalCavaletes > 0 {
		statusCode = http.StatusCreated
	}

	c.JSON(statusCode, gin.H{
		"sucesso": result.Sucesso,
		"mensagem": result.Mensagem,
		"dados": gin.H{
			"uuid": result.UUID,
			"url": result.URL,
			"cavaletes": result.TotalCavaletes,
			"itens": result.TotalItens,
			"metragem_total": result.MetragemTotal,
			"tempo_processamento": result.TempoProcessamento.Seconds(),
			"timestamp": result.Timestamp,
		},
	})
}

// ImportarBatch processa múltiplas URLs
func (h *MobgranHandler) ImportarBatch(c *gin.Context) {
	var req models.BatchImportRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "Requisição inválida",
			"detalhes": err.Error(),
		})
		return
	}

	// Validar URLs
	if len(req.URLs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "Pelo menos uma URL deve ser fornecida",
		})
		return
	}

	if len(req.URLs) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "Máximo de 50 URLs por vez",
		})
		return
	}

	// Valor padrão para atualizar
	atualizarExistente := true
	if req.AtualizarExistente != nil {
		atualizarExistente = *req.AtualizarExistente
	}

	// Executar importações
	results, err := h.importer.ImportarBatch(req.URLs, atualizarExistente)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"erro": "Erro ao processar lote",
			"detalhes": err.Error(),
		})
		return
	}

	// Calcular estatísticas do lote
	var sucessos, falhas int
	var metragTotal float64
	var cavaletesTotal int
	
	for _, r := range results {
		if r.Sucesso {
			sucessos++
			metragTotal += r.MetragemTotal
			cavaletesTotal += r.TotalCavaletes
		} else {
			falhas++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"resumo": gin.H{
			"total_processados": len(results),
			"sucessos": sucessos,
			"falhas": falhas,
			"metragem_total": metragTotal,
			"cavaletes_total": cavaletesTotal,
		},
		"detalhes": results,
	})
}

// VerificarOferta verifica o status de uma oferta específica
func (h *MobgranHandler) VerificarOferta(c *gin.Context) {
	uuid := c.Param("uuid")
	
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "UUID é obrigatório",
		})
		return
	}

	oferta, err := h.importer.ObterOferta(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"erro": "Oferta não encontrada",
			"uuid": uuid,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sucesso": true,
		"oferta": oferta,
	})
}

// ListarOfertas lista todas as ofertas com paginação
func (h *MobgranHandler) ListarOfertas(c *gin.Context) {
	// Parâmetros de paginação
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Buscar ofertas
	ofertas, err := h.importer.ListarOfertas(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"erro": "Erro ao buscar ofertas",
			"detalhes": err.Error(),
		})
		return
	}

	// Preparar resposta simplificada
	var ofertasSimplificadas []gin.H
	for _, oferta := range ofertas {
		ofertasSimplificadas = append(ofertasSimplificadas, gin.H{
			"uuid": oferta.UUID,
			"empresa": oferta.NomeEmpresa,
			"vendedor": oferta.NomeVendedor,
			"data_importacao": oferta.DataImportacao,
			"total_cavaletes": len(oferta.DadosCompletos.Cavaletes),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"sucesso": true,
		"dados": gin.H{
			"ofertas": ofertasSimplificadas,
			"paginacao": gin.H{
				"limit": limit,
				"offset": offset,
				"total": len(ofertas),
				"proxima_pagina": offset + limit,
			},
		},
	})
}

// ObterEstatisticas retorna estatísticas globais do sistema
func (h *MobgranHandler) ObterEstatisticas(c *gin.Context) {
	stats, err := h.importer.ObterEstatisticasGlobais()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"erro": "Erro ao obter estatísticas",
			"detalhes": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sucesso": true,
		"estatisticas": gin.H{
			"total_ofertas": stats.TotalOfertas,
			"total_cavaletes": stats.TotalCavaletes,
			"metragem_total": stats.MetragemTotal,
			"material_mais_comum": stats.MaterialTop,
			"distribuicao_materiais": stats.MaterialMaisComum,
		},
	})
}


#services/mobgran_importer.go
package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"mobgran-importer/models"
	"mobgran-importer/utils"
)

// MobgranImporter é o serviço principal para importar dados do Mobgran
type MobgranImporter struct {
	supabaseURL    string
	supabaseKey    string
	httpClient     *http.Client
	supabaseClient *utils.SupabaseClient
}

// NewMobgranImporter cria uma nova instância do importador
func NewMobgranImporter(supabaseURL, supabaseKey string) *MobgranImporter {
	return &MobgranImporter{
		supabaseURL: supabaseURL,
		supabaseKey: supabaseKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		supabaseClient: utils.NewSupabaseClient(supabaseURL, supabaseKey),
	}
}

// ImportarURL importa dados de uma URL do Mobgran
func (m *MobgranImporter) ImportarURL(url string, atualizarExistente bool) (*models.ImportResult, error) {
	startTime := time.Now()
	result := &models.ImportResult{
		URL:        url,
		Timestamp:  time.Now(),
		Sucesso:    false,
	}

	log.Printf("🔄 Iniciando importação da URL: %s", url)

	// Extrair UUID da URL
	uuid, err := m.extrairUUID(url)
	if err != nil {
		result.Erro = fmt.Sprintf("Erro ao extrair UUID: %v", err)
		log.Printf("❌ %s", result.Erro)
		return result, err
	}
	result.UUID = uuid

	log.Printf("📋 UUID extraído: %s", uuid)

	// Verificar se já existe
	existe, err := m.verificarOfertaExiste(uuid)
	if err != nil {
		result.Erro = fmt.Sprintf("Erro ao verificar oferta: %v", err)
		log.Printf("❌ %s", result.Erro)
		return result, err
	}

	if existe && !atualizarExistente {
		result.Mensagem = fmt.Sprintf("Oferta %s já existe no banco de dados", uuid)
		result.Sucesso = true
		log.Printf("⚠️  %s", result.Mensagem)
		return result, nil
	}

	// Buscar dados da API do Mobgran
	log.Printf("🌐 Buscando dados da API do Mobgran...")
	dadosMobgran, err := m.buscarDadosMobgran(uuid)
	if err != nil {
		result.Erro = fmt.Sprintf("Erro ao buscar dados: %v", err)
		log.Printf("❌ %s", result.Erro)
		return result, err
	}

	// Processar e salvar dados
	log.Printf("💾 Processando e salvando dados no Supabase...")
	if err := m.salvarDados(uuid, dadosMobgran, existe); err != nil {
		result.Erro = fmt.Sprintf("Erro ao salvar dados: %v", err)
		log.Printf("❌ %s", result.Erro)
		return result, err
	}

	// Calcular estatísticas
	stats := m.calcularEstatisticas(dadosMobgran)
	
	result.Sucesso = true
	result.TempoProcessamento = time.Since(startTime)
	result.TotalCavaletes = stats.TotalCavaletes
	result.TotalItens = stats.TotalItens
	result.MetragemTotal = stats.MetragemTotal
	
	if existe {
		result.Mensagem = fmt.Sprintf("✅ Oferta %s atualizada com sucesso", uuid)
	} else {
		result.Mensagem = fmt.Sprintf("✅ Oferta %s importada com sucesso", uuid)
	}
	
	log.Printf("%s | Cavaletes: %d | Itens: %d | Metragem: %.2f m² | Tempo: %.2fs",
		result.Mensagem,
		result.TotalCavaletes,
		result.TotalItens,
		result.MetragemTotal,
		result.TempoProcessamento.Seconds())

	return result, nil
}

// extrairUUID extrai o UUID de uma URL do Mobgran
func (m *MobgranImporter) extrairUUID(url string) (string, error) {
	// Padrões para extrair UUID
	patterns := []string{
		`uuid-([a-zA-Z0-9]+)`,
		`o=([a-zA-Z0-9]+)/`,
		`o=([a-zA-Z0-9]+)$`,
		`uuid/([a-zA-Z0-9]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			uuid := matches[1]
			// Remover "uuid-" se presente
			uuid = strings.TrimPrefix(uuid, "uuid-")
			return uuid, nil
		}
	}

	return "", fmt.Errorf("UUID não encontrado na URL: %s", url)
}

// verificarOfertaExiste verifica se uma oferta já existe no banco
func (m *MobgranImporter) verificarOfertaExiste(uuid string) (bool, error) {
	query := fmt.Sprintf(`{"uuid": "eq.%s"}`, uuid)
	
	resp, err := m.supabaseClient.Get("ofertas", query)
	if err != nil {
		return false, err
	}

	var ofertas []map[string]interface{}
	if err := json.Unmarshal(resp, &ofertas); err != nil {
		return false, err
	}

	return len(ofertas) > 0, nil
}

// buscarDadosMobgran busca os dados na API do Mobgran
func (m *MobgranImporter) buscarDadosMobgran(uuid string) (*models.DadosMobgran, error) {
	apiURL := fmt.Sprintf("https://mobgran.com/app/api/controller.php?o=obterCavaletesPorConferencia&uuid=%s", uuid)
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Headers necessários
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.mobgran.com/")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dados models.DadosMobgran
	if err := json.Unmarshal(body, &dados); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %v", err)
	}

	// Validar dados
	if dados.Empresa == "" {
		return nil, fmt.Errorf("dados inválidos: empresa não encontrada")
	}

	return &dados, nil
}

// salvarDados salva ou atualiza os dados no Supabase
func (m *MobgranImporter) salvarDados(uuid string, dados *models.DadosMobgran, atualizar bool) error {
	// Preparar dados da oferta
	oferta := models.Oferta{
		UUID:           uuid,
		NomeEmpresa:    dados.Empresa,
		NomeVendedor:   dados.Vendedor,
		DadosCompletos: dados,
		DataImportacao: time.Now(),
	}

	// Salvar ou atualizar oferta
	var err error
	if atualizar {
		err = m.supabaseClient.Update("ofertas", uuid, oferta)
	} else {
		err = m.supabaseClient.Insert("ofertas", oferta)
	}
	
	if err != nil {
		return fmt.Errorf("erro ao salvar oferta: %v", err)
	}

	// Deletar cavaletes e itens existentes se for atualização
	if atualizar {
		if err := m.deletarDadosExistentes(uuid); err != nil {
			log.Printf("⚠️  Aviso: erro ao deletar dados existentes: %v", err)
		}
	}

	// Salvar cavaletes
	for _, cavaleteDados := range dados.Cavaletes {
		cavalete := models.Cavalete{
			OfertaUUID:      uuid,
			CodigoNumerico:  cavaleteDados.CodigoNumerico,
			NumeroBloco:     cavaleteDados.NumeroBloco,
			NomeMaterial:    cavaleteDados.NomeMaterial,
			NomeEspessura:   cavaleteDados.NomeEspessura,
			TipoProduto:     cavaleteDados.TipoProduto,
			MedidaA:         cavaleteDados.MedidaA,
			MedidaB:         cavaleteDados.MedidaB,
			MedidaC:         cavaleteDados.MedidaC,
			Volume:          cavaleteDados.Volume,
			Peso:            cavaleteDados.Peso,
			Metragem:        cavaleteDados.Metragem,
			QuantidadeChapas: cavaleteDados.QuantidadeChapas,
			Imagem:          cavaleteDados.Imagem,
			DadosCompletos:  cavaleteDados,
		}

		if err := m.supabaseClient.Insert("cavaletes", cavalete); err != nil {
			log.Printf("⚠️  Erro ao salvar cavalete %s: %v", cavalete.CodigoNumerico, err)
			continue
		}

		// Salvar itens do cavalete
		for _, itemDados := range cavaleteDados.Itens {
			item := models.Item{
				OfertaUUID:     uuid,
				CavaleteID:     cavalete.CodigoNumerico,
				Codigo:         itemDados.Codigo,
				MedidaA:        itemDados.MedidaA,
				MedidaB:        itemDados.MedidaB,
				Espessura:      itemDados.Espessura,
				DadosCompletos: itemDados,
			}

			if err := m.supabaseClient.Insert("itens", item); err != nil {
				log.Printf("⚠️  Erro ao salvar item %s: %v", item.Codigo, err)
			}
		}
	}

	return nil
}

// deletarDadosExistentes remove cavaletes e itens de uma oferta
func (m *MobgranImporter) deletarDadosExistentes(uuid string) error {
	// Deletar itens
	if err := m.supabaseClient.Delete("itens", fmt.Sprintf(`{"oferta_uuid": "eq.%s"}`, uuid)); err != nil {
		return err
	}

	// Deletar cavaletes
	if err := m.supabaseClient.Delete("cavaletes", fmt.Sprintf(`{"oferta_uuid": "eq.%s"}`, uuid)); err != nil {
		return err
	}

	return nil
}

// calcularEstatisticas calcula estatísticas dos dados importados
func (m *MobgranImporter) calcularEstatisticas(dados *models.DadosMobgran) *models.Estatisticas {
	stats := &models.Estatisticas{
		TotalCavaletes: len(dados.Cavaletes),
	}

	for _, cavalete := range dados.Cavaletes {
		stats.TotalItens += len(cavalete.Itens)
		stats.MetragemTotal += cavalete.Metragem
		
		// Contar por material
		if stats.PorMaterial == nil {
			stats.PorMaterial = make(map[string]int)
		}
		stats.PorMaterial[cavalete.NomeMaterial]++
	}

	return stats
}

// ImportarBatch importa múltiplas URLs em lote
func (m *MobgranImporter) ImportarBatch(urls []string, atualizarExistente bool) ([]*models.ImportResult, error) {
	var results []*models.ImportResult

	for _, url := range urls {
		result, err := m.ImportarURL(url, atualizarExistente)
		if err != nil {
			log.Printf("❌ Erro ao importar %s: %v", url, err)
		}
		results = append(results, result)
		
		// Pequena pausa entre requisições para não sobrecarregar
		time.Sleep(500 * time.Millisecond)
	}

	return results, nil
}

// ObterOferta busca uma oferta específica
func (m *MobgranImporter) ObterOferta(uuid string) (*models.Oferta, error) {
	query := fmt.Sprintf(`{"uuid": "eq.%s"}`, uuid)
	
	resp, err := m.supabaseClient.Get("ofertas", query)
	if err != nil {
		return nil, err
	}

	var ofertas []models.Oferta
	if err := json.Unmarshal(resp, &ofertas); err != nil {
		return nil, err
	}

	if len(ofertas) == 0 {
		return nil, fmt.Errorf("oferta não encontrada")
	}

	return &ofertas[0], nil
}

// ListarOfertas lista todas as ofertas com paginação
func (m *MobgranImporter) ListarOfertas(limit, offset int) ([]models.Oferta, error) {
	query := fmt.Sprintf("limit=%d&offset=%d&order=data_importacao.desc", limit, offset)
	
	resp, err := m.supabaseClient.GetWithQuery("ofertas", query)
	if err != nil {
		return nil, err
	}

	var ofertas []models.Oferta
	if err := json.Unmarshal(resp, &ofertas); err != nil {
		return nil, err
	}

	return ofertas, nil
}

// ObterEstatisticasGlobais obtém estatísticas gerais do sistema
func (m *MobgranImporter) ObterEstatisticasGlobais() (*models.EstatisticasGlobais, error) {
	stats := &models.EstatisticasGlobais{}

	// Buscar total de ofertas
	ofertas, err := m.ListarOfertas(1000, 0)
	if err != nil {
		return nil, err
	}
	stats.TotalOfertas = len(ofertas)

	// Buscar totais dos cavaletes
	query := "select=metragem,nome_material"
	resp, err := m.supabaseClient.GetWithQuery("cavaletes", query)
	if err != nil {
		return nil, err
	}

	var cavaletes []struct {
		Metragem     float64 `json:"metragem"`
		NomeMaterial string  `json:"nome_material"`
	}
	
	if err := json.Unmarshal(resp, &cavaletes); err != nil {
		return nil, err
	}

	stats.TotalCavaletes = len(cavaletes)
	stats.MaterialMaisComum = make(map[string]int)

	for _, c := range cavaletes {
		stats.MetragemTotal += c.Metragem
		stats.MaterialMaisComum[c.NomeMaterial]++
	}

	// Encontrar material mais comum
	maxCount := 0
	for material, count := range stats.MaterialMaisComum {
		if count > maxCount {
			maxCount = count
			stats.MaterialTop = material
		}
	}

	return stats, nil
}


#main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"mobgran-importer/handlers"
	"mobgran-importer/services"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Printf("Aviso: arquivo .env não encontrado: %v", err)
	}

	// Validar variáveis de ambiente obrigatórias
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	
	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("SUPABASE_URL e SUPABASE_KEY devem estar configurados")
	}

	// Inicializar o importador
	importer := services.NewMobgranImporter(supabaseURL, supabaseKey)
	handler := handlers.NewMobgranHandler(importer)

	// Configurar o roteador Gin
	router := gin.Default()
	
	// Middleware de CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Rotas da API
	api := router.Group("/api/v1")
	{
		// Rota de health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"message": "Mobgran Importer API está funcionando",
			})
		})

		// Rotas do Mobgran
		mobgran := api.Group("/mobgran")
		{
			// Importar uma URL
			mobgran.POST("/import", handler.ImportarURL)
			
			// Importar múltiplas URLs
			mobgran.POST("/import/batch", handler.ImportarBatch)
			
			// Verificar status de uma oferta
			mobgran.GET("/oferta/:uuid", handler.VerificarOferta)
			
			// Listar ofertas
			mobgran.GET("/ofertas", handler.ListarOfertas)
			
			// Estatísticas
			mobgran.GET("/stats", handler.ObterEstatisticas)
		}
	}

	// Página inicial com documentação
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"nome": "Mobgran Data Importer API",
			"versao": "1.0.0",
			"endpoints": gin.H{
				"POST /api/v1/mobgran/import": "Importar uma URL do Mobgran",
				"POST /api/v1/mobgran/import/batch": "Importar múltiplas URLs",
				"GET /api/v1/mobgran/oferta/:uuid": "Verificar status de uma oferta",
				"GET /api/v1/mobgran/ofertas": "Listar todas as ofertas",
				"GET /api/v1/mobgran/stats": "Obter estatísticas",
				"GET /api/v1/health": "Verificar status da API",
			},
			"documentacao": "https://github.com/seu-usuario/mobgran-importer-go",
		})
	})

	// Configurar porta
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Servidor iniciado na porta %s", port)
	log.Printf("📚 Documentação disponível em http://localhost:%s", port)
	
	// Iniciar servidor
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}



