package supabase

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go"
	"github.com/supabase-community/gotrue-go/types"
	"mobgran-importer-go/internal/models"
)

type AuthClient struct {
	client gotrue.Client
}

func NewAuthClient(supabaseURL, apiKey string) *AuthClient {
	// Extrair project reference da URL
	projectRef := extractProjectRef(supabaseURL)
	client := gotrue.New(projectRef, apiKey)
	return &AuthClient{
		client: client,
	}
}

func NewAuthClientWithServiceKey(supabaseURL, anonKey, serviceKey string) *AuthClient {
	// Extrair project reference da URL
	projectRef := extractProjectRef(supabaseURL)
	client := gotrue.New(projectRef, anonKey).WithToken(serviceKey)
	return &AuthClient{
		client: client,
	}
}

// extractProjectRef extrai o project reference da URL do Supabase
func extractProjectRef(supabaseURL string) string {
	// URL format: https://project-ref.supabase.co
	// Extrair apenas o project-ref
	if len(supabaseURL) == 0 {
		return ""
	}
	
	// Remover https://
	url := supabaseURL
	if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	
	// Extrair a parte antes de .supabase.co
	parts := strings.Split(url, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	
	return url
}

func (a *AuthClient) SignUp(email, password string, userData map[string]interface{}) (*models.SupabaseAuthResponse, error) {
	req := types.SignupRequest{
		Email:    email,
		Password: password,
		Data:     userData,
	}

	resp, err := a.client.Signup(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar usuário: %w", err)
	}

	// Verificar se o usuário foi criado (User é um struct, não um ponteiro)
	if resp.User.ID == uuid.Nil {
		return nil, fmt.Errorf("usuário não foi criado")
	}

	result := &models.SupabaseAuthResponse{
		User: &models.SupabaseUser{
			ID:    resp.User.ID.String(),
			Email: resp.User.Email,
		},
	}

	// Se há um access token, incluir na resposta
	if resp.AccessToken != "" {
		result.Session = &models.SupabaseSession{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			ExpiresAt:    resp.ExpiresAt,
		}
	}

	return result, nil
}

func (a *AuthClient) AdminCreateUser(email, password string, userData map[string]interface{}, emailConfirm bool) (*models.SupabaseAuthResponse, error) {
	req := types.AdminCreateUserRequest{
		Email:        email,
		Password:     &password,
		UserMetadata: userData,
		EmailConfirm: emailConfirm,
	}

	resp, err := a.client.AdminCreateUser(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar usuário admin: %w", err)
	}

	// Verificar se o usuário foi criado
	if resp.User.ID == uuid.Nil {
		return nil, fmt.Errorf("falha ao criar usuário admin")
	}

	result := &models.SupabaseAuthResponse{
		User: &models.SupabaseUser{
			ID:    resp.User.ID.String(),
			Email: resp.User.Email,
		},
	}

	// AdminCreateUser não retorna sessão, apenas o usuário
	return result, nil
}

func (a *AuthClient) ConfirmUser(email, token string) error {
	req := types.VerifyForUserRequest{
		Type:  "signup",
		Email: email,
		Token: token,
	}

	_, err := a.client.VerifyForUser(req)
	if err != nil {
		return fmt.Errorf("erro ao confirmar usuário: %w", err)
	}

	return nil
}

func (a *AuthClient) ConfirmUserByAdmin(userID uuid.UUID) error {
	// Usar a API administrativa para confirmar o usuário
	// Isso requer que o cliente tenha sido inicializado com service key
	req := types.UpdateUserRequest{
		AppData: map[string]interface{}{
			"email_confirmed": true,
		},
	}

	_, err := a.client.UpdateUser(req)
	if err != nil {
		return fmt.Errorf("erro ao confirmar usuário via admin: %w", err)
	}

	return nil
}

func (a *AuthClient) SignIn(email, password string) (*models.SupabaseAuthResponse, error) {
	resp, err := a.client.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer login: %w", err)
	}

	// Verificar se o usuário foi autenticado (User é um struct, não um ponteiro)
	if resp.User.ID == uuid.Nil {
		return nil, fmt.Errorf("falha na autenticação")
	}

	return &models.SupabaseAuthResponse{
		User: &models.SupabaseUser{
			ID:    resp.User.ID.String(),
			Email: resp.User.Email,
		},
		Session: &models.SupabaseSession{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			ExpiresAt:    resp.ExpiresAt,
		},
	}, nil
}

func (a *AuthClient) GetUser(token string) (*models.SupabaseUser, error) {
	clientWithToken := a.client.WithToken(token)
	resp, err := clientWithToken.GetUser()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter usuário: %w", err)
	}

	return &models.SupabaseUser{
		ID:    resp.ID.String(),
		Email: resp.Email,
	}, nil
}

func (a *AuthClient) RefreshToken(refreshToken string) (*models.SupabaseSession, error) {
	resp, err := a.client.RefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao renovar token: %w", err)
	}

	return &models.SupabaseSession{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    resp.ExpiresAt,
	}, nil
}

func (a *AuthClient) SignOut(token string) error {
	clientWithToken := a.client.WithToken(token)
	err := clientWithToken.Logout()
	if err != nil {
		return fmt.Errorf("erro ao fazer logout: %w", err)
	}
	return nil
}

// Função auxiliar para converter int64 para time.Time
func int64ToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}