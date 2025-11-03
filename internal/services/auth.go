package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	
	"mobgran-importer-go/internal/auth"
	"mobgran-importer-go/internal/models"
	"mobgran-importer-go/pkg/database"
)

// AuthService gerencia operações de autenticação
type AuthService struct {
	db *database.PostgresClient
}

// NewAuthService cria uma nova instância do serviço de autenticação
func NewAuthService(db *database.PostgresClient) *AuthService {
	return &AuthService{db: db}
}

// RegistrarTrader registra um novo trader no sistema
func (s *AuthService) RegistrarTrader(ctx context.Context, registro *models.TraderRegistro) (*models.Trader, error) {
	// Verifica se o email já existe
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM traders WHERE email = $1)", registro.Email).Scan(&exists)
	if err != nil {
		return nil, models.NewInternalError("Erro interno do servidor")
	}

	if exists {
		return nil, models.NewConflictError("Email já está em uso")
	}

	// Gera hash da senha
	senhaHash, err := bcrypt.GenerateFromPassword([]byte(registro.Senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, models.NewInternalError("Erro interno do servidor")
	}

	// Insere o trader
	query := `
		INSERT INTO traders (nome, email, senha_hash, telefone, empresa, ativo)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, nome, email, telefone, empresa, ativo, created_at, updated_at
	`

	trader := &models.Trader{}
	err = s.db.QueryRow(
		query,
		registro.Nome,
		registro.Email,
		string(senhaHash),
		registro.Telefone,
		registro.Empresa,
	).Scan(
		&trader.ID,
		&trader.Nome,
		&trader.Email,
		&trader.Telefone,
		&trader.Empresa,
		&trader.Ativo,
		&trader.CreatedAt,
		&trader.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, models.NewConflictError("Email já está em uso")
		}
		return nil, models.NewInternalError("Erro interno do servidor")
	}

	return trader, nil
}

// LoginWithToken autentica um trader e retorna AuthResponse com JWT token
func (s *AuthService) LoginWithToken(ctx context.Context, login *models.TraderLogin) (*models.AuthResponse, error) {
	// Primeiro autentica o trader
	trader, err := s.Login(ctx, login)
	if err != nil {
		return nil, err
	}

	// Gera o JWT token
	token, expiresAt, err := auth.GenerateCustomJWT(trader.ID, trader.Email, trader.Nome)
	if err != nil {
		return nil, models.NewInternalError("Erro ao gerar token de autenticação")
	}

	// Cria a resposta de autenticação
	authResponse := &models.AuthResponse{
		Token:        token,
		RefreshToken: "", // TODO: Implementar refresh token se necessário
		ExpiresAt:    expiresAt,
		Trader:       trader.ToResponse(),
	}

	return authResponse, nil
}

// Login autentica um trader e retorna os dados
func (s *AuthService) Login(ctx context.Context, login *models.TraderLogin) (*models.Trader, error) {
	query := `
		SELECT id, nome, email, senha_hash, telefone, empresa, ativo, created_at, updated_at
		FROM traders
		WHERE email = $1 AND ativo = true
	`

	trader := &models.Trader{}
	var senhaHash string

	err := s.db.QueryRow(query, login.Email).Scan(
		&trader.ID,
		&trader.Nome,
		&trader.Email,
		&senhaHash,
		&trader.Telefone,
		&trader.Empresa,
		&trader.Ativo,
		&trader.CreatedAt,
		&trader.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.NewAuthenticationError("Email ou senha incorretos")
	}
	if err != nil {
		return nil, models.NewInternalError("Erro interno do servidor")
	}

	// Verifica a senha
	if err := bcrypt.CompareHashAndPassword([]byte(senhaHash), []byte(login.Senha)); err != nil {
		return nil, models.NewAuthenticationError("Email ou senha incorretos")
	}

	return trader, nil
}

// BuscarTraderPorID busca um trader pelo ID
func (s *AuthService) BuscarTraderPorID(ctx context.Context, traderID string) (*models.Trader, error) {
	query := `
		SELECT id, nome, email, telefone, empresa, ativo, created_at, updated_at
		FROM traders
		WHERE id = $1 AND ativo = true
	`

	trader := &models.Trader{}
	err := s.db.QueryRow(query, traderID).Scan(
		&trader.ID,
		&trader.Nome,
		&trader.Email,
		&trader.Telefone,
		&trader.Empresa,
		&trader.Ativo,
		&trader.CreatedAt,
		&trader.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.NewNotFoundError("Trader não encontrado")
	}
	if err != nil {
		return nil, models.NewInternalError("Erro interno do servidor")
	}

	return trader, nil
}

// RefreshToken gera um novo token para o trader (implementação básica)
func (s *AuthService) RefreshToken(ctx context.Context, traderID string) (*models.Trader, error) {
	// Por enquanto, apenas retorna os dados do trader
	// Em uma implementação real, você geraria um novo JWT token aqui
	return s.BuscarTraderPorID(ctx, traderID)
}

// Logout realiza o logout do trader (implementação básica)
func (s *AuthService) Logout(ctx context.Context, traderID string) error {
	// Por enquanto, apenas valida se o trader existe
	// Em uma implementação real, você invalidaria o token aqui
	_, err := s.BuscarTraderPorID(ctx, traderID)
	if err != nil {
		return err
	}
	return nil
}

// BuscarTrader é um alias para BuscarTraderPorID para compatibilidade
func (s *AuthService) BuscarTrader(ctx context.Context, traderID string) (*models.Trader, error) {
	return s.BuscarTraderPorID(ctx, traderID)
}

// AtualizarTrader atualiza os dados de um trader
func (s *AuthService) AtualizarTrader(ctx context.Context, traderID string, dados *models.TraderAtualizar) (*models.Trader, error) {
	// Monta query dinâmica baseada nos campos fornecidos
	updates := []string{}
	args := []interface{}{}
	argCount := 1

	if dados.Nome != nil {
		updates = append(updates, fmt.Sprintf("nome = $%d", argCount))
		args = append(args, *dados.Nome)
		argCount++
	}

	if dados.Telefone != nil {
		updates = append(updates, fmt.Sprintf("telefone = $%d", argCount))
		args = append(args, *dados.Telefone)
		argCount++
	}

	if dados.Empresa != nil {
		updates = append(updates, fmt.Sprintf("empresa = $%d", argCount))
		args = append(args, *dados.Empresa)
		argCount++
	}

	if len(updates) == 0 {
		return nil, models.NewBadRequestError("Nenhum campo para atualizar", "")
	}

	// Adiciona updated_at
	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

	// Adiciona WHERE clause
	args = append(args, traderID)
	whereClause := fmt.Sprintf("WHERE id = $%d", argCount)

	query := fmt.Sprintf(`
		UPDATE traders 
		SET %s 
		%s
		RETURNING id, nome, email, telefone, empresa, ativo, created_at, updated_at
	`, strings.Join(updates, ", "), whereClause)

	trader := &models.Trader{}
	err := s.db.QueryRow(query, args...).Scan(
		&trader.ID,
		&trader.Nome,
		&trader.Email,
		&trader.Telefone,
		&trader.Empresa,
		&trader.Ativo,
		&trader.CreatedAt,
		&trader.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.NewNotFoundError("Trader não encontrado")
	}
	if err != nil {
		return nil, models.NewInternalError("Erro interno do servidor")
	}

	return trader, nil
}

// AlterarSenha altera a senha de um trader
func (s *AuthService) AlterarSenha(ctx context.Context, traderID string, senhaAtual, novaSenha string) error {
	// Busca a senha atual
	var senhaHash string
	err := s.db.QueryRow("SELECT senha_hash FROM traders WHERE id = $1 AND ativo = true", traderID).Scan(&senhaHash)
	if err == sql.ErrNoRows {
		return models.NewNotFoundError("Trader não encontrado")
	}
	if err != nil {
		return models.NewInternalError("Erro interno do servidor")
	}

	// Verifica a senha atual
	if err := bcrypt.CompareHashAndPassword([]byte(senhaHash), []byte(senhaAtual)); err != nil {
		return models.NewAuthenticationError("Senha atual incorreta")
	}

	// Gera hash da nova senha
	novoHash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		return models.NewInternalError("Erro interno do servidor")
	}

	// Atualiza a senha
	_, err = s.db.Exec(
		"UPDATE traders SET senha_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		string(novoHash), traderID,
	)
	if err != nil {
		return models.NewInternalError("Erro interno do servidor")
	}

	return nil
}

// DesativarTrader desativa um trader
func (s *AuthService) DesativarTrader(ctx context.Context, traderID string) error {
	result, err := s.db.Exec(
		"UPDATE traders SET ativo = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND ativo = true",
		traderID,
	)
	if err != nil {
		return models.NewInternalError("Erro interno do servidor")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewInternalError("Erro interno do servidor")
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("Trader não encontrado")
	}

	return nil
}

// ListarTraders lista todos os traders ativos com paginação
func (s *AuthService) ListarTraders(ctx context.Context, limite, offset int) ([]*models.Trader, int, error) {
	// Conta total de traders ativos
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM traders WHERE ativo = true").Scan(&total)
	if err != nil {
		return nil, 0, models.NewInternalError("Erro interno do servidor")
	}

	// Busca traders com paginação
	query := `
		SELECT id, nome, email, telefone, empresa, ativo, created_at, updated_at
		FROM traders
		WHERE ativo = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limite, offset)
	if err != nil {
		return nil, 0, models.NewInternalError("Erro interno do servidor")
	}
	defer rows.Close()

	var traders []*models.Trader
	for rows.Next() {
		trader := &models.Trader{}
		err := rows.Scan(
			&trader.ID,
			&trader.Nome,
			&trader.Email,
			&trader.Telefone,
			&trader.Empresa,
			&trader.Ativo,
			&trader.CreatedAt,
			&trader.UpdatedAt,
		)
		if err != nil {
			return nil, 0, models.NewInternalError("Erro interno do servidor")
		}
		traders = append(traders, trader)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, models.NewInternalError("Erro interno do servidor")
	}

	return traders, total, nil
}

// BuscarTraderPorEmail busca um trader pelo email
func (s *AuthService) BuscarTraderPorEmail(ctx context.Context, email string) (*models.Trader, error) {
	query := `
		SELECT id, nome, email, telefone, empresa, ativo, created_at, updated_at
		FROM traders
		WHERE email = $1 AND ativo = true
	`

	trader := &models.Trader{}
	err := s.db.QueryRow(query, email).Scan(
		&trader.ID,
		&trader.Nome,
		&trader.Email,
		&trader.Telefone,
		&trader.Empresa,
		&trader.Ativo,
		&trader.CreatedAt,
		&trader.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.NewNotFoundError("Trader não encontrado")
	}
	if err != nil {
		return nil, models.NewInternalError("Erro interno do servidor")
	}

	return trader, nil
}