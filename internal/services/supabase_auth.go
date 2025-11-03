package services

import (
	"github.com/sirupsen/logrus"
	"mobgran-importer-go/internal/config"
	"mobgran-importer-go/internal/models"
	"mobgran-importer-go/pkg/supabase"
)

type SupabaseAuthService struct {
	authClient *supabase.AuthClient
	logger     *logrus.Logger
	config     *config.Config
}

func NewSupabaseAuthService(cfg *config.Config, logger *logrus.Logger) *SupabaseAuthService {
	authClient := supabase.NewAuthClient(cfg.SupabaseURL, cfg.SupabaseKey)
	
	return &SupabaseAuthService{
		authClient: authClient,
		logger:     logger,
		config:     cfg,
	}
}

func (s *SupabaseAuthService) CriarUsuarioAdmin(email, password string, userData map[string]interface{}) (*models.SupabaseAuthResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"email": email,
	}).Info("Criando usuário admin no Supabase")

	// Criar cliente com service key para privilégios administrativos
	adminClient := supabase.NewAuthClientWithServiceKey(s.config.SupabaseURL, s.config.SupabaseKey, s.config.SupabaseServiceKey)

	// Usar AdminCreateUser com email_confirm = true para criar usuário já confirmado
	resp, err := adminClient.AdminCreateUser(email, password, userData, true)
	if err != nil {
		s.logger.WithError(err).Error("Erro ao criar usuário admin no Supabase")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": resp.User.ID,
		"email":   resp.User.Email,
	}).Info("Usuário admin criado com sucesso no Supabase")

	return resp, nil
}

func (s *SupabaseAuthService) RegistrarUsuario(email, password string, userData map[string]interface{}) (*models.SupabaseAuthResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"email": email,
	}).Info("Registrando usuário no Supabase")

	resp, err := s.authClient.SignUp(email, password, userData)
	if err != nil {
		s.logger.WithError(err).Error("Erro ao registrar usuário no Supabase")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": resp.User.ID,
		"email":   resp.User.Email,
	}).Info("Usuário registrado com sucesso no Supabase")

	return resp, nil
}

func (s *SupabaseAuthService) FazerLogin(email, password string) (*models.SupabaseAuthResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"email": email,
	}).Info("Fazendo login no Supabase")

	resp, err := s.authClient.SignIn(email, password)
	if err != nil {
		s.logger.WithError(err).Error("Erro ao fazer login no Supabase")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": resp.User.ID,
		"email":   resp.User.Email,
	}).Info("Login realizado com sucesso no Supabase")

	return resp, nil
}

func (s *SupabaseAuthService) ObterUsuario(token string) (*models.SupabaseUser, error) {
	s.logger.Info("Obtendo usuário do Supabase")

	user, err := s.authClient.GetUser(token)
	if err != nil {
		s.logger.WithError(err).Error("Erro ao obter usuário do Supabase")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Usuário obtido com sucesso do Supabase")

	return user, nil
}

func (s *SupabaseAuthService) RenovarToken(refreshToken string) (*models.SupabaseSession, error) {
	s.logger.Info("Renovando token no Supabase")

	session, err := s.authClient.RefreshToken(refreshToken)
	if err != nil {
		s.logger.WithError(err).Error("Erro ao renovar token no Supabase")
		return nil, err
	}

	s.logger.Info("Token renovado com sucesso no Supabase")
	return session, nil
}

func (s *SupabaseAuthService) FazerLogout(token string) error {
	s.logger.Info("Fazendo logout no Supabase")

	err := s.authClient.SignOut(token)
	if err != nil {
		s.logger.WithError(err).Error("Erro ao fazer logout no Supabase")
		return err
	}

	s.logger.Info("Logout realizado com sucesso no Supabase")
	return nil
}