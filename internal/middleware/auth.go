package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"mobgran-importer-go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// SupabaseAuthMiddleware verifica se o token JWT do Supabase é válido
func SupabaseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.APIError{
					Type:    "authentication_error",
					Message: "Token de autorização não fornecido",
				},
			})
			c.Abort()
			return
		}

		// Verifica se o header tem o formato "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.APIError{
					Type:    "authentication_error",
					Message: "Formato de token inválido. Use 'Bearer <token>'",
				},
			})
			c.Abort()
			return
		}

		// Valida o token JWT do Supabase
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verifica se o método de assinatura é HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
			}

			// Retorna a chave secreta do JWT do Supabase
			jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
			if jwtSecret == "" {
				return nil, fmt.Errorf("SUPABASE_JWT_SECRET não configurado")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			logrus.WithError(err).Error("Token JWT inválido")
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.APIError{
					Type:    "authentication_error",
					Message: "Token inválido ou expirado",
				},
			})
			c.Abort()
			return
		}

		// Extrai as claims do token
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Adiciona as informações do usuário ao contexto
			c.Set("user_id", claims["sub"])
			c.Set("user_email", claims["email"])
			c.Set("user_role", claims["role"])
		}

		c.Next()
	}
}

// GetSupabaseUserFromContext obtém dados do usuário do contexto Supabase
func GetSupabaseUserFromContext(c *gin.Context) (string, string, string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", "", "", fmt.Errorf("user_id não encontrado no contexto")
	}

	userEmail, exists := c.Get("user_email")
	if !exists {
		return "", "", "", fmt.Errorf("user_email não encontrado no contexto")
	}

	userRole, exists := c.Get("user_role")
	if !exists {
		return "", "", "", fmt.Errorf("user_role não encontrado no contexto")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", "", "", fmt.Errorf("user_id não é uma string válida")
	}

	userEmailStr, ok := userEmail.(string)
	if !ok {
		return "", "", "", fmt.Errorf("user_email não é uma string válida")
	}

	userRoleStr, ok := userRole.(string)
	if !ok {
		return "", "", "", fmt.Errorf("user_role não é uma string válida")
	}

	return userIDStr, userEmailStr, userRoleStr, nil
}

// CORSMiddleware adiciona headers CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Lista de origens permitidas
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001", 
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"http://127.0.0.1:8080",
		}

		// Verifica se a origem está na lista permitida
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adiciona headers de segurança
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Previne ataques XSS
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Content Security Policy básico
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		// Força HTTPS em produção
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		c.Next()
	}
}

// LoggerMiddleware adiciona logging customizado
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RecoveryMiddleware adiciona recuperação de panic customizada
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logrus.WithField("panic", recovered).Error("Panic recuperado")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.APIError{
				Type:    "internal_error",
				Message: "Erro interno do servidor",
			},
		})
	})
}