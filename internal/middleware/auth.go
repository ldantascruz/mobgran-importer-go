package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"mobgran-importer-go/internal/auth"
	"mobgran-importer-go/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims representa as claims customizadas do JWT (mantido para compatibilidade)
type JWTClaims struct {
	TraderID uuid.UUID `json:"trader_id"`
	Email    string    `json:"email"`
	Nome     string    `json:"nome"`
	jwt.RegisteredClaims
}

// AuthMiddleware verifica se o token JWT é válido seguindo as práticas recomendadas
func AuthMiddleware() gin.HandlerFunc {
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
					Message: "Formato de token inválido. Use: Bearer <token>",
				},
			})
			c.Abort()
			return
		}

		// Tentar validar como token customizado primeiro
		if customClaims, err := auth.ParseCustomJWT(tokenString); err == nil {
			userCtx := &auth.UserContext{
				UserID:    customClaims.Subject,
				Email:     customClaims.Email,
				Nome:      customClaims.Nome,
				Role:      customClaims.Role,
				SessionID: "", // Não aplicável para tokens customizados
			}

			// Adicionar ao contexto da requisição
			ctx := auth.WithUserContext(c.Request.Context(), userCtx)
			c.Request = c.Request.WithContext(ctx)

			// Manter compatibilidade com código existente
			c.Set("trader_id", customClaims.TraderID)
			c.Set("trader_email", customClaims.Email)
			c.Set("trader_nome", customClaims.Nome)

			c.Next()
			return
		}

		// Tentar validar como token do Supabase
		if supabaseClaims, err := auth.ParseSupabaseJWT(tokenString); err == nil {
			userCtx := &auth.UserContext{
				UserID:    supabaseClaims.Subject,
				Email:     supabaseClaims.Email,
				Nome:      "", // Supabase não tem nome nas claims padrão
				Role:      supabaseClaims.Role,
				SessionID: supabaseClaims.SessionID,
			}

			// Adicionar ao contexto da requisição
			ctx := auth.WithUserContext(c.Request.Context(), userCtx)
			c.Request = c.Request.WithContext(ctx)

			// Manter compatibilidade com código existente
			if userID, err := uuid.Parse(supabaseClaims.Subject); err == nil {
				c.Set("trader_id", userID)
			}
			c.Set("trader_email", supabaseClaims.Email)
			c.Set("trader_nome", "")

			c.Next()
			return
		}

		// Se nenhum dos dois funcionou, token inválido
		logrus.WithField("token_prefix", tokenString[:min(10, len(tokenString))]).Warn("Token JWT inválido")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: models.APIError{
				Type:    "authentication_error",
				Message: "Token inválido ou expirado",
			},
		})
		c.Abort()
	}
}

// SupabaseAuthMiddleware middleware específico para tokens do Supabase
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

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.APIError{
					Type:    "authentication_error",
					Message: "Formato de token inválido. Use: Bearer <token>",
				},
			})
			c.Abort()
			return
		}

		claims, err := auth.ParseSupabaseJWT(tokenString)
		if err != nil {
			logrus.WithError(err).Warn("Token Supabase inválido")
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: models.APIError{
					Type:    "authentication_error",
					Message: "Token inválido ou expirado",
				},
			})
			c.Abort()
			return
		}

		userCtx := &auth.UserContext{
			UserID:    claims.Subject,
			Email:     claims.Email,
			Role:      claims.Role,
			SessionID: claims.SessionID,
		}

		// Adicionar ao contexto da requisição
		ctx := auth.WithUserContext(c.Request.Context(), userCtx)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GenerateJWT gera um token JWT para o trader
func GenerateJWT(traderID uuid.UUID, email, nome string) (string, time.Time, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", time.Time{}, fmt.Errorf("JWT_SECRET não configurado")
	}

	expirationTime := time.Now().Add(24 * time.Hour) // Token válido por 24 horas

	claims := &JWTClaims{
		TraderID: traderID,
		Email:    email,
		Nome:     nome,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mobgran-importer",
			Subject:   traderID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("erro ao assinar token: %w", err)
	}

	return tokenString, expirationTime, nil
}

// ValidateJWT valida um token JWT e retorna as claims
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET não configurado")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verifica se o método de assinatura é HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao validar token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	return claims, nil
}

// GenerateRefreshToken gera um refresh token aleatório
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("erro ao gerar refresh token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword gera um hash bcrypt da senha
func HashPassword(password string) (string, error) {
	const cost = 10 // Custo balanceado entre segurança e performance
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword verifica se a senha corresponde ao hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetTraderFromContext extrai as informações do trader do contexto Gin
func GetTraderFromContext(c *gin.Context) (uuid.UUID, string, string, error) {
	traderID, exists := c.Get("trader_id")
	if !exists {
		return uuid.Nil, "", "", fmt.Errorf("trader_id não encontrado no contexto")
	}

	email, exists := c.Get("trader_email")
	if !exists {
		return uuid.Nil, "", "", fmt.Errorf("trader_email não encontrado no contexto")
	}

	nome, exists := c.Get("trader_nome")
	if !exists {
		return uuid.Nil, "", "", fmt.Errorf("trader_nome não encontrado no contexto")
	}

	traderUUID, ok := traderID.(uuid.UUID)
	if !ok {
		return uuid.Nil, "", "", fmt.Errorf("trader_id tem tipo inválido")
	}

	traderEmail, ok := email.(string)
	if !ok {
		return uuid.Nil, "", "", fmt.Errorf("trader_email tem tipo inválido")
	}

	traderNome, ok := nome.(string)
	if !ok {
		return uuid.Nil, "", "", fmt.Errorf("trader_nome tem tipo inválido")
	}

	return traderUUID, traderEmail, traderNome, nil
}

// CORS middleware para permitir requisições cross-origin
// CORSMiddleware configurado para frontend Next.js seguindo práticas recomendadas
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Lista de origens permitidas (configurável via variáveis de ambiente)
		allowedOrigins := []string{
			"http://localhost:3000",  // Next.js dev
			"http://localhost:3001",  // Next.js dev alternativo
			"https://localhost:3000", // Next.js dev com HTTPS
			"https://localhost:3001", // Next.js dev alternativo com HTTPS
		}

		// Adicionar origens de produção se configuradas
		if prodOrigin := os.Getenv("FRONTEND_URL"); prodOrigin != "" {
			allowedOrigins = append(allowedOrigins, prodOrigin)
		}

		// Verificar se a origem está na lista permitida
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Header("Access-Control-Max-Age", "86400") // Cache preflight por 24 horas

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adiciona headers de segurança recomendados
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Previne ataques XSS
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy básico
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'")

		// Força HTTPS em produção
		if os.Getenv("GIN_MODE") == "release" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Previne referrer leakage
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Controla recursos que podem ser carregados
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// LoggerMiddleware personalizado para logging estruturado
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"user_agent":  param.Request.UserAgent(),
			"error":       param.ErrorMessage,
		}).Info("HTTP Request")
		return ""
	})
}

// RecoveryMiddleware personalizado para capturar panics
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logrus.WithFields(logrus.Fields{
			"panic":  recovered,
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		}).Error("Panic recuperado")

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.APIError{
				Type:    "internal_error",
				Message: "Erro interno do servidor",
			},
		})
	})
}