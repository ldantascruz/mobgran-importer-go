package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"mobgran-importer-go/internal/models"
	"mobgran-importer-go/internal/services"
)

// ImporterHandler representa o handler para operações de importação
type ImporterHandler struct {
	importerService *services.MobgranImporter
	logger          *logrus.Logger
}

// NewImporterHandler cria uma nova instância do handler
func NewImporterHandler(importerService *services.MobgranImporter, logger *logrus.Logger) *ImporterHandler {
	return &ImporterHandler{
		importerService: importerService,
		logger:          logger,
	}
}

// ImportarOferta importa uma oferta do Mobgran
// @Summary Importa uma oferta do Mobgran
// @Description Importa dados de uma oferta do Mobgran para o Supabase
// @Tags importacao
// @Accept json
// @Produce json
// @Param request body models.ImportRequest true "Dados da importação"
// @Success 200 {object} models.ImportResponse
// @Failure 400 {object} models.ImportResponse
// @Failure 500 {object} models.ImportResponse
// @Router /api/importar [post]
func (h *ImporterHandler) ImportarOferta(c *gin.Context) {
	var request models.ImportRequest

	// Validar JSON de entrada
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("Erro ao validar JSON de entrada")
		c.JSON(http.StatusBadRequest, models.ImportResponse{
			Sucesso:  false,
			Mensagem: fmt.Sprintf("Dados inválidos: %v", err),
		})
		return
	}

	// Log da requisição
	h.logger.WithFields(logrus.Fields{
		"url":                 request.URL,
		"atualizar_existente": request.AtualizarExistente,
		"client_ip":           c.ClientIP(),
	}).Info("Recebida requisição de importação")

	// Validar URL
	if err := h.importerService.ValidarURL(request.URL); err != nil {
		h.logger.WithError(err).Error("URL inválida")
		c.JSON(http.StatusBadRequest, models.ImportResponse{
			Sucesso:  false,
			Mensagem: fmt.Sprintf("URL inválida: %v", err),
		})
		return
	}

	// Executar importação
	sucesso, mensagem, uuid, err := h.importerService.Importar(
		request.URL,
		request.AtualizarExistente,
	)

	// Preparar resposta
	response := models.ImportResponse{
		Sucesso:  sucesso,
		Mensagem: mensagem,
	}

	if uuid != nil {
		response.UUIDLink = *uuid
	}

	// Determinar status HTTP
	statusCode := http.StatusOK
	if !sucesso {
		if err != nil {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusBadRequest
		}
	}

	// Log do resultado
	h.logger.WithFields(logrus.Fields{
		"sucesso":     sucesso,
		"uuid":        response.UUIDLink,
		"status_code": statusCode,
	}).Info("Importação processada")

	c.JSON(statusCode, response)
}

// HealthCheck verifica a saúde da aplicação
// @Summary Health check
// @Description Verifica se a aplicação está funcionando
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *ImporterHandler) HealthCheck(c *gin.Context) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "mobgran-importer-go",
		"version":   "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}

// ValidarURL valida uma URL do Mobgran
// @Summary Valida URL do Mobgran
// @Description Valida se uma URL é um link válido do Mobgran
// @Tags validacao
// @Accept json
// @Produce json
// @Param request body map[string]string true "URL para validar"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/validar-url [post]
func (h *ImporterHandler) ValidarURL(c *gin.Context) {
	var request map[string]string

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"valida":   false,
			"mensagem": fmt.Sprintf("Dados inválidos: %v", err),
		})
		return
	}

	url, exists := request["url"]
	if !exists {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"valida":   false,
			"mensagem": "Campo 'url' é obrigatório",
		})
		return
	}

	// Validar URL
	err := h.importerService.ValidarURL(url)
	if err != nil {
		c.JSON(http.StatusOK, map[string]interface{}{
			"valida":   false,
			"mensagem": err.Error(),
		})
		return
	}

	// Extrair UUID para mostrar na resposta
	uuid, err := h.importerService.ExtrairUUIDLink(url)
	response := map[string]interface{}{
		"valida":   true,
		"mensagem": "URL válida",
	}

	if err == nil && uuid != nil {
		response["uuid"] = *uuid
	}

	c.JSON(http.StatusOK, response)
}

// ExtrairUUID extrai o UUID de uma URL do Mobgran
// @Summary Extrai UUID da URL
// @Description Extrai o UUID de uma URL do Mobgran
// @Tags utilidades
// @Accept json
// @Produce json
// @Param request body map[string]string true "URL para extrair UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/extrair-uuid [post]
func (h *ImporterHandler) ExtrairUUID(c *gin.Context) {
	var request map[string]string

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"sucesso":  false,
			"mensagem": fmt.Sprintf("Dados inválidos: %v", err),
		})
		return
	}

	url, exists := request["url"]
	if !exists {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"sucesso":  false,
			"mensagem": "Campo 'url' é obrigatório",
		})
		return
	}

	// Extrair UUID
	uuid, err := h.importerService.ExtrairUUIDLink(url)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"sucesso":  false,
			"mensagem": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"sucesso": true,
		"uuid":    *uuid,
	})
}