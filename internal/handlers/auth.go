package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"mobgran-importer-go/internal/middleware"
	"mobgran-importer-go/internal/models"
	"mobgran-importer-go/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// handleError processa erros de forma padronizada
func (h *AuthHandler) handleError(c *gin.Context, err error) {
	if apiErr, ok := err.(*models.APIError); ok {
		logrus.WithFields(logrus.Fields{
			"type":    apiErr.Type,
			"message": apiErr.Message,
			"details": apiErr.Details,
		}).Error("API Error")
		
		c.JSON(apiErr.StatusCode, models.ErrorResponse{Error: *apiErr})
		return
	}

	// Erro não tipado - trata como erro interno
	logrus.WithError(err).Error("Erro interno não tipado")
	internalErr := models.NewInternalError("Erro interno do servidor")
	c.JSON(internalErr.StatusCode, models.ErrorResponse{Error: *internalErr})
}

// @Summary Registrar novo trader
// @Description Registra um novo trader no sistema
// @Tags auth
// @Accept json
// @Produce json
// @Param trader body models.TraderRegistro true "Dados do trader"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/registrar [post]
func (h *AuthHandler) Registrar(c *gin.Context) {
	var req models.TraderRegistro
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Erro ao fazer bind do JSON")
		validationErr := models.NewValidationError("Dados inválidos", err.Error())
		c.JSON(validationErr.StatusCode, models.ErrorResponse{Error: *validationErr})
		return
	}

	authResponse, err := h.authService.RegistrarTrader(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, authResponse)
}

// @Summary Login do trader
// @Description Autentica um trader no sistema
// @Tags auth
// @Accept json
// @Produce json
// @Param login body models.TraderLogin true "Credenciais do trader"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.TraderLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Erro ao fazer bind do JSON")
		validationErr := models.NewValidationError("Dados inválidos", err.Error())
		c.JSON(validationErr.StatusCode, models.ErrorResponse{Error: *validationErr})
		return
	}

	authResponse, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// @Summary Refresh token
// @Description Renova o token de acesso usando o refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body map[string]string true "Refresh token"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Erro ao fazer bind do JSON")
		validationErr := models.NewValidationError("Dados inválidos", err.Error())
		c.JSON(validationErr.StatusCode, models.ErrorResponse{Error: *validationErr})
		return
	}

	refreshToken, exists := req["refresh_token"]
	if !exists || refreshToken == "" {
		validationErr := models.NewValidationError("Refresh token é obrigatório", "")
		c.JSON(validationErr.StatusCode, models.ErrorResponse{Error: *validationErr})
		return
	}

	// Por enquanto, apenas valida o formato
	// Em uma implementação real, você validaria e renovaria o token JWT
	authErr := models.NewAuthenticationError("Funcionalidade não implementada")
	c.JSON(authErr.StatusCode, models.ErrorResponse{Error: *authErr})
}

// @Summary Logout
// @Description Faz logout do trader (revoga refresh token)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param refresh body map[string]string true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		authErr := models.NewAuthenticationError("Trader não encontrado no contexto")
		c.JSON(authErr.StatusCode, models.ErrorResponse{Error: *authErr})
		return
	}

	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		validationErr := models.NewValidationError("Refresh token é obrigatório", err.Error())
		c.JSON(validationErr.StatusCode, models.ErrorResponse{Error: *validationErr})
		return
	}

	err = h.authService.Logout(c.Request.Context(), traderID.String())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Logout realizado com sucesso"})
}

// @Summary Perfil do trader
// @Description Obtém o perfil do trader autenticado
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.TraderResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/perfil [get]
func (h *AuthHandler) Perfil(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		authErr := models.NewAuthenticationError("Trader não encontrado no contexto")
		c.JSON(authErr.StatusCode, models.ErrorResponse{Error: *authErr})
		return
	}

	traderResponse, err := h.authService.BuscarTrader(c.Request.Context(), traderID.String())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, traderResponse)
}

// @Summary Atualizar perfil
// @Description Atualiza o perfil do trader autenticado
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param trader body models.TraderAtualizar true "Dados para atualização"
// @Success 200 {object} models.TraderResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/perfil [put]
func (h *AuthHandler) AtualizarPerfil(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		authErr := models.NewAuthenticationError("Trader não encontrado no contexto")
		c.JSON(authErr.StatusCode, models.ErrorResponse{Error: *authErr})
		return
	}

	var dados models.TraderAtualizar
	if err := c.ShouldBindJSON(&dados); err != nil {
		validationErr := models.NewValidationError("Dados inválidos", err.Error())
		c.JSON(validationErr.StatusCode, models.ErrorResponse{Error: *validationErr})
		return
	}

	traderResponse, err := h.authService.AtualizarTrader(c.Request.Context(), traderID.String(), &dados)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, traderResponse)
}