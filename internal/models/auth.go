package models

import (
	"time"

	"github.com/google/uuid"
)

// Trader representa um usuário do sistema
type Trader struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Nome            string     `json:"nome" db:"nome" binding:"required,min=2,max=255"`
	Email           string     `json:"email" db:"email" binding:"required,email"`
	SenhaHash       string     `json:"-" db:"senha_hash"`
	Telefone        *string    `json:"telefone,omitempty" db:"telefone"`
	Empresa         *string    `json:"empresa,omitempty" db:"empresa"`
	Ativo           bool       `json:"ativo" db:"ativo"`
	EmailVerificado bool       `json:"email_verificado" db:"email_verificado"`
	UltimoLogin     *time.Time `json:"ultimo_login,omitempty" db:"ultimo_login"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// TraderRegistro representa os dados para registro de um novo trader
type TraderRegistro struct {
	Nome     string  `json:"nome" binding:"required,min=2,max=255"`
	Email    string  `json:"email" binding:"required,email"`
	Senha    string  `json:"senha" binding:"required,min=6,max=100"`
	Telefone *string `json:"telefone,omitempty"`
	Empresa  *string `json:"empresa,omitempty"`
}

// TraderLogin representa os dados para login
type TraderLogin struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required"`
}

// TraderAtualizar representa os dados para atualização do perfil
type TraderAtualizar struct {
	Nome     *string `json:"nome,omitempty" binding:"omitempty,min=2,max=255"`
	Telefone *string `json:"telefone,omitempty"`
	Empresa  *string `json:"empresa,omitempty"`
}

// TraderResponse representa a resposta com dados do trader (sem senha)
type TraderResponse struct {
	ID              uuid.UUID  `json:"id"`
	Nome            string     `json:"nome"`
	Email           string     `json:"email"`
	Telefone        *string    `json:"telefone,omitempty"`
	Empresa         *string    `json:"empresa,omitempty"`
	Ativo           bool       `json:"ativo"`
	EmailVerificado bool       `json:"email_verificado"`
	UltimoLogin     *time.Time `json:"ultimo_login,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ProdutoAprovado representa um produto na vitrine do trader
type ProdutoAprovado struct {
	ID              uuid.UUID `json:"id" db:"id"`
	TraderID        uuid.UUID `json:"trader_id" db:"trader_id"`
	CavaleteID      uuid.UUID `json:"cavalete_id" db:"cavalete_id"`
	NomeCustomizado string    `json:"nome_customizado" db:"nome_customizado" binding:"required,min=1,max=255"`
	PrecoVenda      float64   `json:"preco_venda" db:"preco_venda" binding:"required,gt=0"`
	Descricao       *string   `json:"descricao,omitempty" db:"descricao"`
	Visivel         bool      `json:"visivel" db:"visivel"`
	Destaque        bool      `json:"destaque" db:"destaque"`
	OrdemExibicao   int       `json:"ordem_exibicao" db:"ordem_exibicao"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ProdutoAprovarRequest representa os dados para aprovar um produto
type ProdutoAprovarRequest struct {
	CavaleteID      uuid.UUID `json:"cavalete_id" binding:"required"`
	NomeCustomizado string    `json:"nome_customizado" binding:"required,min=1,max=255"`
	PrecoVenda      float64   `json:"preco_venda" binding:"required,gt=0"`
	Descricao       *string   `json:"descricao,omitempty"`
	Visivel         *bool     `json:"visivel,omitempty"`
	Destaque        *bool     `json:"destaque,omitempty"`
}

// ProdutoAtualizarRequest representa os dados para atualizar um produto
type ProdutoAtualizarRequest struct {
	NomeCustomizado *string  `json:"nome_customizado,omitempty" binding:"omitempty,min=1,max=255"`
	PrecoVenda      *float64 `json:"preco_venda,omitempty" binding:"omitempty,gt=0"`
	Descricao       *string  `json:"descricao,omitempty"`
	Visivel         *bool    `json:"visivel,omitempty"`
	Destaque        *bool    `json:"destaque,omitempty"`
	OrdemExibicao   *int     `json:"ordem_exibicao,omitempty"`
}

// CavaleteDisponivel representa um cavalete disponível para aprovação
type CavaleteDisponivel struct {
	ID                string      `json:"id" db:"id"`
	OfertaID          string      `json:"oferta_id" db:"oferta_id"`
	Codigo            string      `json:"codigo" db:"codigo"`
	Bloco             string      `json:"bloco" db:"bloco"`
	NomeMaterial      string      `json:"nome_material" db:"nome_material"`
	NomeEspessura     string      `json:"nome_espessura" db:"nome_espessura"`
	NomeClassificacao string      `json:"nome_classificacao" db:"nome_classificacao"`
	NomeAcabamento    *string     `json:"nome_acabamento" db:"nome_acabamento"`
	Comprimento       *float64    `json:"comprimento" db:"comprimento"`
	Altura            *float64    `json:"altura" db:"altura"`
	Largura           *float64    `json:"largura" db:"largura"`
	Metragem          *float64    `json:"metragem" db:"metragem"`
	Peso              *float64    `json:"peso" db:"peso"`
	TipoMetragem      *string     `json:"tipo_metragem" db:"tipo_metragem"`
	ImagemPrincipal   interface{} `json:"imagem_principal" db:"imagem_principal"`
	ImagensAdicionais interface{} `json:"imagens_adicionais" db:"imagens_adicionais"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
	TraderID          uuid.UUID   `json:"trader_id" db:"trader_id"`
	NomeEmpresa       string      `json:"nome_empresa" db:"nome_empresa"`
	JaAprovado        bool        `json:"ja_aprovado" db:"ja_aprovado"`
}

// VitrinePublica representa um produto na vitrine pública
type VitrinePublica struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	TraderID        uuid.UUID   `json:"trader_id" db:"trader_id"`
	NomeCustomizado string      `json:"nome_customizado" db:"nome_customizado"`
	PrecoVenda      float64     `json:"preco_venda" db:"preco_venda"`
	Descricao       *string     `json:"descricao,omitempty" db:"descricao"`
	Destaque        bool        `json:"destaque" db:"destaque"`
	OrdemExibicao   int         `json:"ordem_exibicao" db:"ordem_exibicao"`
	Codigo          string      `json:"codigo" db:"codigo"`
	Bloco           string      `json:"bloco" db:"bloco"`
	NomeMaterial    string      `json:"nome_material" db:"nome_material"`
	NomeEspessura   string      `json:"nome_espessura" db:"nome_espessura"`
	NomeClassificacao *string   `json:"nome_classificacao,omitempty" db:"nome_classificacao"`
	NomeAcabamento  *string     `json:"nome_acabamento,omitempty" db:"nome_acabamento"`
	Comprimento     *float64    `json:"comprimento,omitempty" db:"comprimento"`
	Altura          *float64    `json:"altura,omitempty" db:"altura"`
	Largura         *float64    `json:"largura,omitempty" db:"largura"`
	Metragem        *float64    `json:"metragem,omitempty" db:"metragem"`
	Peso            *float64    `json:"peso,omitempty" db:"peso"`
	TipoMetragem    *string     `json:"tipo_metragem,omitempty" db:"tipo_metragem"`
	ImagemPrincipal interface{} `json:"imagem_principal,omitempty" db:"imagem_principal"`
	ImagensAdicionais interface{} `json:"imagens_adicionais,omitempty" db:"imagens_adicionais"`
	TraderNome      string      `json:"trader_nome" db:"trader_nome"`
	TraderEmpresa   *string     `json:"trader_empresa,omitempty" db:"trader_empresa"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
}

// RefreshToken representa um token de refresh JWT
type RefreshToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TraderID  uuid.UUID `json:"trader_id" db:"trader_id"`
	TokenHash string    `json:"-" db:"token_hash"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Revogado  bool      `json:"revogado" db:"revogado"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuthResponse representa a resposta de autenticação
type AuthResponse struct {
	Trader       TraderResponse `json:"trader"`
	Token        string         `json:"token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresAt    time.Time      `json:"expires_at"`
}

// TokenClaims representa as claims do JWT
type TokenClaims struct {
	TraderID uuid.UUID `json:"trader_id"`
	Email    string    `json:"email"`
	Nome     string    `json:"nome"`
}

// EstatisticasProdutos representa estatísticas dos produtos do trader
type EstatisticasProdutos struct {
	TotalProdutos     int `json:"total_produtos"`
	ProdutosVisiveis  int `json:"produtos_visiveis"`
	ProdutosDestaque  int `json:"produtos_destaque"`
	CavaletesDisponiveis int `json:"cavaletes_disponiveis"`
}

// ToResponse converte Trader para TraderResponse
func (t *Trader) ToResponse() TraderResponse {
	return TraderResponse{
		ID:              t.ID,
		Nome:            t.Nome,
		Email:           t.Email,
		Telefone:        t.Telefone,
		Empresa:         t.Empresa,
		Ativo:           t.Ativo,
		EmailVerificado: t.EmailVerificado,
		UltimoLogin:     t.UltimoLogin,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}