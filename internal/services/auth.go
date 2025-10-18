package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"mobgran-importer-go/internal/middleware"
	"mobgran-importer-go/internal/models"
)

// AuthService gerencia operações de autenticação
type AuthService struct {
	db *sql.DB
}

// NewAuthService cria uma nova instância do AuthService
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// RegistrarTrader registra um novo trader no sistema
func (s *AuthService) RegistrarTrader(registro *models.TraderRegistro) (*models.TraderResponse, error) {
	// Verifica se o email já existe
	var existingID uuid.UUID
	err := s.db.QueryRow("SELECT id FROM traders WHERE email = $1", registro.Email).Scan(&existingID)
	if err == nil {
		return nil, fmt.Errorf("email já está em uso")
	} else if err != sql.ErrNoRows {
		logrus.WithError(err).Error("Erro ao verificar email existente")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	// Gera hash da senha
	hashedPassword, err := middleware.HashPassword(registro.Senha)
	if err != nil {
		logrus.WithError(err).Error("Erro ao gerar hash da senha")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	// Cria o trader
	trader := &models.Trader{
		ID:        uuid.New(),
		Nome:      registro.Nome,
		Email:     registro.Email,
		SenhaHash: hashedPassword,
		Telefone:  registro.Telefone,
		Empresa:   registro.Empresa,
		Ativo:     true,
	}

	query := `
		INSERT INTO traders (id, nome, email, senha_hash, telefone, empresa, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`

	_, err = s.db.Exec(query, trader.ID, trader.Nome, trader.Email, trader.SenhaHash,
		trader.Telefone, trader.Empresa, trader.Ativo)
	if err != nil {
		logrus.WithError(err).Error("Erro ao inserir trader no banco")
		return nil, fmt.Errorf("erro ao criar trader")
	}

	logrus.WithFields(logrus.Fields{
		"trader_id": trader.ID,
		"email":     trader.Email,
		"nome":      trader.Nome,
	}).Info("Trader registrado com sucesso")

	response := trader.ToResponse()
	return &response, nil
}

// Login autentica um trader e retorna tokens JWT
func (s *AuthService) Login(login *models.TraderLogin) (*models.AuthResponse, error) {
	var trader models.Trader

	query := `
		SELECT id, nome, email, senha_hash, telefone, empresa, ativo, created_at, updated_at
		FROM traders
		WHERE email = $1 AND ativo = true
	`

	err := s.db.QueryRow(query, login.Email).Scan(
		&trader.ID, &trader.Nome, &trader.Email, &trader.SenhaHash,
		&trader.Telefone, &trader.Empresa, &trader.Ativo,
		&trader.CreatedAt, &trader.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("credenciais inválidas")
	} else if err != nil {
		logrus.WithError(err).Error("Erro ao buscar trader no banco")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	// Verifica a senha
	if !middleware.CheckPassword(login.Senha, trader.SenhaHash) {
		return nil, fmt.Errorf("credenciais inválidas")
	}

	// Gera tokens
	accessToken, expiresAt, err := middleware.GenerateJWT(trader.ID, trader.Email, trader.Nome)
	if err != nil {
		logrus.WithError(err).Error("Erro ao gerar access token")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		logrus.WithError(err).Error("Erro ao gerar refresh token")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	// Salva o refresh token no banco
	err = s.salvarRefreshToken(trader.ID, refreshToken)
	if err != nil {
		logrus.WithError(err).Error("Erro ao salvar refresh token")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	logrus.WithFields(logrus.Fields{
		"trader_id": trader.ID,
		"email":     trader.Email,
	}).Info("Login realizado com sucesso")

	return &models.AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Trader:       trader.ToResponse(),
	}, nil
}

// RefreshToken gera um novo access token usando o refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*models.AuthResponse, error) {
	var traderID uuid.UUID
	var createdAt time.Time

	query := `
		SELECT trader_id, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	err := s.db.QueryRow(query, refreshToken).Scan(&traderID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("refresh token inválido ou expirado")
	} else if err != nil {
		logrus.WithError(err).Error("Erro ao buscar refresh token")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	// Busca dados do trader
	var trader models.Trader
	traderQuery := `
		SELECT id, nome, email, telefone, empresa, ativo, created_at, updated_at
		FROM traders
		WHERE id = $1 AND ativo = true
	`

	err = s.db.QueryRow(traderQuery, traderID).Scan(
		&trader.ID, &trader.Nome, &trader.Email,
		&trader.Telefone, &trader.Empresa, &trader.Ativo,
		&trader.CreatedAt, &trader.UpdatedAt,
	)

	if err != nil {
		logrus.WithError(err).Error("Erro ao buscar trader para refresh")
		return nil, fmt.Errorf("trader não encontrado")
	}

	// Gera novo access token
	accessToken, expiresAt, err := middleware.GenerateJWT(trader.ID, trader.Email, trader.Nome)
	if err != nil {
		logrus.WithError(err).Error("Erro ao gerar novo access token")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	// Gera novo refresh token
	newRefreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		logrus.WithError(err).Error("Erro ao gerar novo refresh token")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	// Remove o refresh token antigo e salva o novo
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação")
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM refresh_tokens WHERE token_hash = $1", refreshToken)
	if err != nil {
		logrus.WithError(err).Error("Erro ao remover refresh token antigo")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	err = s.salvarRefreshTokenTx(tx, trader.ID, newRefreshToken)
	if err != nil {
		logrus.WithError(err).Error("Erro ao salvar novo refresh token")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	err = tx.Commit()
	if err != nil {
		logrus.WithError(err).Error("Erro ao confirmar transação de refresh")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	return &models.AuthResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		Trader:       trader.ToResponse(),
	}, nil
}

// Logout invalida o refresh token do trader
func (s *AuthService) Logout(traderID uuid.UUID, refreshToken string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE trader_id = $1 AND token_hash = $2
	`

	result, err := s.db.Exec(query, traderID, refreshToken)
	if err != nil {
		logrus.WithError(err).Error("Erro ao remover refresh token no logout")
		return fmt.Errorf("erro interno do servidor")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("Erro ao verificar linhas afetadas no logout")
		return fmt.Errorf("erro interno do servidor")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("refresh token não encontrado")
	}

	logrus.WithField("trader_id", traderID).Info("Logout realizado com sucesso")
	return nil
}

// BuscarTrader busca um trader por ID
func (s *AuthService) BuscarTrader(traderID uuid.UUID) (*models.TraderResponse, error) {
	var trader models.Trader

	query := `
		SELECT id, nome, email, telefone, empresa, ativo, created_at, updated_at
		FROM traders
		WHERE id = $1 AND ativo = true
	`

	err := s.db.QueryRow(query, traderID).Scan(
		&trader.ID, &trader.Nome, &trader.Email,
		&trader.Telefone, &trader.Empresa, &trader.Ativo,
		&trader.CreatedAt, &trader.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("trader não encontrado")
	} else if err != nil {
		logrus.WithError(err).Error("Erro ao buscar trader")
		return nil, fmt.Errorf("erro interno do servidor")
	}

	response := trader.ToResponse()
	return &response, nil
}

// AtualizarTrader atualiza os dados de um trader
func (s *AuthService) AtualizarTrader(traderID uuid.UUID, dados *models.TraderAtualizar) (*models.TraderResponse, error) {
	// Verifica se o trader existe
	_, err := s.BuscarTrader(traderID)
	if err != nil {
		return nil, err
	}

	// Constrói a query de atualização dinamicamente
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if dados.Nome != nil && *dados.Nome != "" {
		setParts = append(setParts, fmt.Sprintf("nome = $%d", argIndex))
		args = append(args, *dados.Nome)
		argIndex++
	}

	if dados.Telefone != nil {
		setParts = append(setParts, fmt.Sprintf("telefone = $%d", argIndex))
		args = append(args, dados.Telefone)
		argIndex++
	}

	if dados.Empresa != nil {
		setParts = append(setParts, fmt.Sprintf("empresa = $%d", argIndex))
		args = append(args, dados.Empresa)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("nenhum campo para atualizar")
	}

	// Adiciona updated_at e trader_id
	setParts = append(setParts, fmt.Sprintf("updated_at = NOW()"))
	args = append(args, traderID)

	query := fmt.Sprintf(`
		UPDATE traders
		SET %s
		WHERE id = $%d
	`, strings.Join(setParts, ", "), argIndex)

	_, err = s.db.Exec(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Erro ao atualizar trader")
		return nil, fmt.Errorf("erro ao atualizar trader")
	}

	// Retorna o trader atualizado
	return s.BuscarTrader(traderID)
}

// salvarRefreshToken salva um refresh token no banco
func (s *AuthService) salvarRefreshToken(traderID uuid.UUID, token string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = s.salvarRefreshTokenTx(tx, traderID, token)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// salvarRefreshTokenTx salva um refresh token usando uma transação existente
func (s *AuthService) salvarRefreshTokenTx(tx *sql.Tx, traderID uuid.UUID, token string) error {
	// Remove tokens antigos do trader (mantém apenas o mais recente)
	_, err := tx.Exec("DELETE FROM refresh_tokens WHERE trader_id = $1", traderID)
	if err != nil {
		return err
	}

	// Insere o novo token (válido por 30 dias)
	query := `
		INSERT INTO refresh_tokens (id, trader_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, NOW() + INTERVAL '30 days', NOW())
	`

	_, err = tx.Exec(query, uuid.New(), traderID, token)
	return err
}