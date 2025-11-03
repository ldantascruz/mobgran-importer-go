package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"mobgran-importer-go/internal/models"
	"mobgran-importer-go/internal/services"
)

type SupabaseAuthHandler struct {
	supabaseAuthService *services.SupabaseAuthService
	logger              *logrus.Logger
}

func NewSupabaseAuthHandler(supabaseAuthService *services.SupabaseAuthService, logger *logrus.Logger) *SupabaseAuthHandler {
	return &SupabaseAuthHandler{
		supabaseAuthService: supabaseAuthService,
		logger:              logger,
	}
}

// handleError processa erros de forma padronizada
func (h *SupabaseAuthHandler) handleError(c *gin.Context, err error) {
	if apiErr, ok := err.(*models.APIError); ok {
		h.logger.WithFields(logrus.Fields{
			"type":    apiErr.Type,
			"message": apiErr.Message,
			"details": apiErr.Details,
		}).Error("API Error")
		
		c.JSON(apiErr.StatusCode, models.ErrorResponse{Error: *apiErr})
		return
	}

	// Erro não tipado - trata como erro interno
	h.logger.WithError(err).Error("Erro interno não tipado")
	internalErr := models.NewInternalError("Erro interno do servidor")
	c.JSON(internalErr.StatusCode, models.ErrorResponse{Error: *internalErr})
}

// Estruturas para requests
type SupabaseRegistroRequest struct {
	Email    string                 `json:"email" binding:"required,email"`
	Password string                 `json:"password" binding:"required,min=6"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type SupabaseLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary Criar usuário admin no Supabase
// @Description Cria um novo usuário admin pré-confirmado usando Supabase Auth
// @Tags supabase-auth
// @Accept json
// @Produce json
// @Param user body SupabaseRegistroRequest true "Dados do usuário admin"
// @Success 201 {object} models.SupabaseAuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /supabase/auth/admin/create [post]
func (h *SupabaseAuthHandler) CriarUsuarioAdmin(c *gin.Context) {
	var req SupabaseRegistroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, models.NewValidationError("Dados inválidos", err.Error()))
		return
	}

	// Validações básicas
	if req.Email == "" || req.Password == "" {
		h.handleError(c, models.NewValidationError("Email e senha são obrigatórios", ""))
		return
	}

	// Criar usuário admin
	resp, err := h.supabaseAuthService.CriarUsuarioAdmin(req.Email, req.Password, req.Data)
	if err != nil {
		h.logger.WithError(err).Error("Erro ao criar usuário admin")
		h.handleError(c, models.NewInternalError("Erro ao criar usuário admin"))
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *SupabaseAuthHandler) Registrar(c *gin.Context) {
	var req SupabaseRegistroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Erro ao fazer bind do JSON")
		apiErr := models.NewValidationError("Dados inválidos", err.Error())
		h.handleError(c, apiErr)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"email": req.Email,
	}).Info("Tentativa de registro no Supabase")

	resp, err := h.supabaseAuthService.RegistrarUsuario(req.Email, req.Password, req.Data)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": resp.User.ID,
		"email":   resp.User.Email,
	}).Info("Usuário registrado com sucesso no Supabase")

	c.JSON(http.StatusCreated, resp)
}

// @Summary Login no Supabase
// @Description Autentica um usuário usando Supabase Auth
// @Tags supabase-auth
// @Accept json
// @Produce json
// @Param credentials body SupabaseLoginRequest true "Credenciais do usuário"
// @Success 200 {object} models.SupabaseAuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /supabase/auth/login [post]
func (h *SupabaseAuthHandler) Login(c *gin.Context) {
	var req SupabaseLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Erro ao fazer bind do JSON")
		apiErr := models.NewValidationError("Dados inválidos", err.Error())
		h.handleError(c, apiErr)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"email": req.Email,
	}).Info("Tentativa de login no Supabase")

	resp, err := h.supabaseAuthService.FazerLogin(req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": resp.User.ID,
		"email":   resp.User.Email,
	}).Info("Login realizado com sucesso no Supabase")

	c.JSON(http.StatusOK, resp)
}

// @Summary Obter usuário atual
// @Description Obtém informações do usuário autenticado no Supabase
// @Tags supabase-auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SupabaseUser
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /supabase/auth/user [get]
func (h *SupabaseAuthHandler) ObterUsuario(c *gin.Context) {
	// Obter token do header Authorization
	token := c.GetHeader("Authorization")
	if token == "" {
		apiErr := models.NewAuthenticationError("Token de acesso não fornecido")
		h.handleError(c, apiErr)
		return
	}

	// Remover "Bearer " do token se presente
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	h.logger.Info("Obtendo usuário do Supabase")

	user, err := h.supabaseAuthService.ObterUsuario(token)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Usuário obtido com sucesso do Supabase")

	c.JSON(http.StatusOK, user)
}

// @Summary Renovar token
// @Description Renova o token de acesso usando o refresh token
// @Tags supabase-auth
// @Accept json
// @Produce json
// @Param refresh body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.SupabaseSession
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /supabase/auth/refresh [post]
func (h *SupabaseAuthHandler) RenovarToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Erro ao fazer bind do JSON")
		apiErr := models.NewValidationError("Dados inválidos", err.Error())
		h.handleError(c, apiErr)
		return
	}

	h.logger.Info("Renovando token no Supabase")

	session, err := h.supabaseAuthService.RenovarToken(req.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("Token renovado com sucesso no Supabase")

	c.JSON(http.StatusOK, session)
}

// @Summary Logout
// @Description Faz logout do usuário no Supabase
// @Tags supabase-auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /supabase/auth/logout [post]
func (h *SupabaseAuthHandler) Logout(c *gin.Context) {
	// Obter token do header Authorization
	token := c.GetHeader("Authorization")
	if token == "" {
		apiErr := models.NewAuthenticationError("Token de acesso não fornecido")
		h.handleError(c, apiErr)
		return
	}

	// Remover "Bearer " do token se presente
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	h.logger.Info("Fazendo logout no Supabase")

	err := h.supabaseAuthService.FazerLogout(token)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logger.Info("Logout realizado com sucesso no Supabase")

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout realizado com sucesso",
	})
}