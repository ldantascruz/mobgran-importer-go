package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"mobgran-importer-go/internal/middleware"
	"mobgran-importer-go/internal/models"
	"mobgran-importer-go/internal/services"
)

type ProdutosHandler struct {
	produtosService *services.ProdutosService
}

func NewProdutosHandler(produtosService *services.ProdutosService) *ProdutosHandler {
	return &ProdutosHandler{
		produtosService: produtosService,
	}
}

// @Summary Listar cavaletes disponíveis
// @Description Lista cavaletes disponíveis para aprovação pelo trader
// @Tags produtos
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limite de resultados" default(20)
// @Param offset query int false "Offset para paginação" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /produtos/cavaletes [get]
func (h *ProdutosHandler) ListarCavaletesDisponiveis(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	// Parâmetros de paginação
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	cavaletes, err := h.produtosService.ListarCavaletesDisponiveis(traderID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Erro ao listar cavaletes disponíveis")
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cavaletes": cavaletes,
		"total":     len(cavaletes),
		"limit":     limit,
		"offset":    offset,
	})
}

// @Summary Aprovar produto
// @Description Aprova um cavalete como produto para venda
// @Tags produtos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param produto body models.ProdutoAprovarRequest true "Dados do produto"
// @Success 201 {object} models.ProdutoAprovado
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /produtos/aprovar [post]
func (h *ProdutosHandler) AprovarProduto(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	var req models.ProdutoAprovarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Erro ao fazer bind do JSON")
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos", "detalhes": err.Error()})
		return
	}

	produto, err := h.produtosService.AprovarProduto(traderID, &req)
	if err != nil {
		logrus.WithError(err).Error("Erro ao aprovar produto")
		if err.Error() == "produto já foi aprovado" {
			c.JSON(http.StatusConflict, gin.H{"erro": "Produto já foi aprovado"})
			return
		}
		if err.Error() == "cavalete não encontrado" {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "Cavalete não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusCreated, produto)
}

// @Summary Listar produtos aprovados
// @Description Lista produtos aprovados pelo trader
// @Tags produtos
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limite de resultados" default(20)
// @Param offset query int false "Offset para paginação" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /produtos/aprovados [get]
func (h *ProdutosHandler) ListarProdutosAprovados(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	// Parâmetros de paginação
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	produtos, err := h.produtosService.ListarProdutosAprovados(traderID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Erro ao listar produtos aprovados")
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"produtos": produtos,
		"total":    len(produtos),
		"limit":    limit,
		"offset":   offset,
	})
}

// @Summary Atualizar produto
// @Description Atualiza um produto aprovado
// @Tags produtos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do produto"
// @Param produto body models.ProdutoAtualizarRequest true "Dados para atualização"
// @Success 200 {object} models.ProdutoAprovado
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /produtos/{id} [put]
func (h *ProdutosHandler) AtualizarProduto(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	produtoIDStr := c.Param("id")
	produtoID, err := uuid.Parse(produtoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "ID do produto inválido"})
		return
	}

	var req models.ProdutoAtualizarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Erro ao fazer bind do JSON")
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos", "detalhes": err.Error()})
		return
	}

	produto, err := h.produtosService.AtualizarProduto(traderID, produtoID, &req)
	if err != nil {
		logrus.WithError(err).Error("Erro ao atualizar produto")
		if err.Error() == "produto não encontrado" {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Produto não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, produto)
}

// @Summary Buscar produto
// @Description Busca um produto específico do trader
// @Tags produtos
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do produto"
// @Success 200 {object} models.ProdutoAprovado
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /produtos/{id} [get]
func (h *ProdutosHandler) BuscarProduto(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	produtoIDStr := c.Param("id")
	produtoID, err := uuid.Parse(produtoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "ID do produto inválido"})
		return
	}

	produto, err := h.produtosService.BuscarProduto(traderID, produtoID)
	if err != nil {
		logrus.WithError(err).Error("Erro ao buscar produto")
		if err.Error() == "produto não encontrado" {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Produto não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, produto)
}

// @Summary Remover produto
// @Description Remove um produto aprovado
// @Tags produtos
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do produto"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /produtos/{id} [delete]
func (h *ProdutosHandler) RemoverProduto(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	produtoIDStr := c.Param("id")
	produtoID, err := uuid.Parse(produtoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "ID do produto inválido"})
		return
	}

	err = h.produtosService.RemoverProduto(traderID, produtoID)
	if err != nil {
		logrus.WithError(err).Error("Erro ao remover produto")
		if err.Error() == "produto não encontrado" {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Produto não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Produto removido com sucesso"})
}

// @Summary Vitrine pública
// @Description Lista produtos na vitrine pública (não requer autenticação)
// @Tags produtos
// @Produce json
// @Param limit query int false "Limite de resultados" default(20)
// @Param offset query int false "Offset para paginação" default(0)
// @Param trader_id query string false "Filtrar por trader específico"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /vitrine/publica [get]
func (h *ProdutosHandler) ListarVitrinePublica(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	destaqueStr := c.DefaultQuery("destaque", "false")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	destaque, err := strconv.ParseBool(destaqueStr)
	if err != nil {
		destaque = false
	}

	produtos, err := h.produtosService.ListarVitrinePublica(limit, offset, destaque)
	if err != nil {
		logrus.WithError(err).Error("Erro ao listar vitrine pública")
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"produtos": produtos,
		"total":    len(produtos),
	})
}

// @Summary Estatísticas de produtos
// @Description Obtém estatísticas dos produtos do trader
// @Tags produtos
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.EstatisticasProdutos
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /produtos/estatisticas [get]
func (h *ProdutosHandler) ObterEstatisticas(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	estatisticas, err := h.produtosService.ObterEstatisticas(traderID)
	if err != nil {
		logrus.WithError(err).Error("Erro ao obter estatísticas")
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, estatisticas)
}

// @Summary Limpar todos os registros do banco de dados
// @Description Remove todos os registros de produtos, cavaletes, ofertas e dados relacionados do banco de dados
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/limpar-dados [delete]
func (h *ProdutosHandler) LimparTodosRegistros(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	logrus.WithField("trader_id", traderID).Info("Iniciando limpeza de todos os registros")

	err = h.produtosService.LimparTodosRegistros()
	if err != nil {
		logrus.WithError(err).Error("Erro ao limpar todos os registros")
		c.JSON(http.StatusInternalServerError, gin.H{
			"erro": "Erro interno do servidor ao limpar registros",
		})
		return
	}

	logrus.WithField("trader_id", traderID).Info("Limpeza de todos os registros concluída com sucesso")

	c.JSON(http.StatusOK, gin.H{
		"sucesso":  true,
		"mensagem": "Todos os registros foram removidos com sucesso",
	})
}