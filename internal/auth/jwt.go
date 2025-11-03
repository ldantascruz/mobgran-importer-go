package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SupabaseClaims representa as claims do JWT do Supabase
type SupabaseClaims struct {
	Email     string `json:"email"`
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// CustomClaims representa as claims customizadas do nosso sistema
type CustomClaims struct {
	TraderID uuid.UUID `json:"trader_id"`
	Email    string    `json:"email"`
	Nome     string    `json:"nome"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// UserContext representa o contexto do usuário autenticado
type UserContext struct {
	UserID    string
	Email     string
	Nome      string
	Role      string
	SessionID string
}

type contextKey string

const UserContextKey contextKey = "user"

// ParseSupabaseJWT valida um token JWT do Supabase
func ParseSupabaseJWT(tokenString string) (*SupabaseClaims, error) {
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET não configurado")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&SupabaseClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verificar método de assinatura HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao validar token: %w", err)
	}

	if claims, ok := token.Claims.(*SupabaseClaims); ok && token.Valid {
		// Verificar se o token não expirou
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("token expirado")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("token inválido")
}

// ParseCustomJWT valida um token JWT customizado do nosso sistema
func ParseCustomJWT(tokenString string) (*CustomClaims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET não configurado")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verificar método de assinatura HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao validar token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// Verificar se o token não expirou
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("token expirado")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("token inválido")
}

// GenerateCustomJWT gera um token JWT customizado
func GenerateCustomJWT(traderID uuid.UUID, email, nome string) (string, time.Time, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", time.Time{}, fmt.Errorf("JWT_SECRET não configurado")
	}

	// Token válido por 1 hora (recomendação do documento)
	expirationTime := time.Now().Add(1 * time.Hour)

	claims := &CustomClaims{
		TraderID: traderID,
		Email:    email,
		Nome:     nome,
		Role:     "authenticated",
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

// GetUserFromContext extrai o contexto do usuário do contexto da requisição
func GetUserFromContext(ctx context.Context) (*UserContext, error) {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	if !ok {
		return nil, fmt.Errorf("contexto do usuário não encontrado")
	}
	return user, nil
}

// WithUserContext adiciona o contexto do usuário ao contexto da requisição
func WithUserContext(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}