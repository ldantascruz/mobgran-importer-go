package handlers

import (
	"github.com/gin-gonic/gin"
	"mobgran-importer-go/internal/auth"
	"mobgran-importer-go/internal/models"
	"net/http"
)

// GetUserFromContext extrai o contexto do usuário da requisição
func GetUserFromContext(c *gin.Context) (*auth.UserContext, error) {
	userCtx, err := auth.GetUserFromContext(c.Request.Context())
	if err != nil {
		return nil, err
	}
	return userCtx, nil
}

// RequireUser middleware que garante que o usuário está autenticado e disponível no contexto
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx, err := GetUserFromContext(c)
		if err != nil || userCtx == nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.APIError{
					Type:    "authentication_error",
					Message: "Usuário não autenticado",
				},
			})
			c.Abort()
			return
		}

		// Adicionar informações do usuário ao contexto do Gin para compatibilidade
		c.Set("user_context", userCtx)
		c.Set("user_id", userCtx.UserID)
		c.Set("user_email", userCtx.Email)
		c.Set("user_role", userCtx.Role)

		c.Next()
	}
}

// RequireRole middleware que verifica se o usuário tem uma role específica
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx, err := GetUserFromContext(c)
		if err != nil || userCtx == nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.APIError{
					Type:    "authentication_error",
					Message: "Usuário não autenticado",
				},
			})
			c.Abort()
			return
		}

		if userCtx.Role != requiredRole {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: models.APIError{
					Type:    "authorization_error",
					Message: "Acesso negado: permissões insuficientes",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}