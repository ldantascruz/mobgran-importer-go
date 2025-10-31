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

// @Summary Registrar novo trader
// @Description Registra um novo trader no sistema
// @Tags auth
// @Accept json
// @Produce json
// @Param trader body models.TraderRegistro true "Dados do trader"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/registrar [post]
func (h *AuthHandler) Registrar(c *gin.Context) {
	var req models.TraderRegistro
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Erro ao fazer bind do JSON")
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos", "detalhes": err.Error()})
		return
	}

	authResponse, err := h.authService.RegistrarTrader(c.Request.Context(), &req)
	if err != nil {
		logrus.WithError(err).Error("Erro ao registrar trader")
		if err.Error() == "email já está em uso" {
			c.JSON(http.StatusConflict, gin.H{"erro": "Email já está em uso"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
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
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.TraderLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Erro ao fazer bind do JSON")
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos", "detalhes": err.Error()})
		return
	}

	authResponse, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		logrus.WithError(err).Error("Erro ao fazer login")
		if err.Error() == "credenciais inválidas" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Email ou senha incorretos"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
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
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Erro ao fazer bind do JSON")
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	refreshToken, ok := req["refresh_token"]
	if !ok || refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Refresh token é obrigatório"})
		return
	}

	authResponse, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		logrus.WithError(err).Error("Erro ao renovar token")
		if err.Error() == "refresh token inválido ou expirado" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Refresh token inválido ou expirado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// @Summary Logout
// @Description Faz logout do trader (revoga refresh token)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param refresh body map[string]string true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Refresh token é obrigatório"})
		return
	}

	err = h.authService.Logout(c.Request.Context(), traderID.String())
	if err != nil {
		logrus.WithError(err).Error("Erro ao fazer logout")
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
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
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/perfil [get]
func (h *AuthHandler) Perfil(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	traderResponse, err := h.authService.BuscarTrader(c.Request.Context(), traderID.String())
	if err != nil {
		logrus.WithError(err).Error("Erro ao buscar perfil do trader")
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
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
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/perfil [put]
func (h *AuthHandler) AtualizarPerfil(c *gin.Context) {
	traderID, _, _, err := middleware.GetTraderFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Trader não encontrado no contexto"})
		return
	}

	var dados models.TraderAtualizar
	if err := c.ShouldBindJSON(&dados); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos"})
		return
	}

	traderResponse, err := h.authService.AtualizarTrader(c.Request.Context(), traderID.String(), &dados)
	if err != nil {
		logrus.WithError(err).Error("Erro ao atualizar perfil do trader")
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, traderResponse)
}