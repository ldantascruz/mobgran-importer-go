package models

import (
	"time"
)

// MobgranResponse representa a resposta completa da API do Mobgran
type MobgranResponse struct {
	Situacao    string     `json:"situacao"`
	NomeEmpresa string     `json:"nomeEmpresa"`
	URLLogo     string     `json:"urlLogo"`
	Cavaletes   []Cavalete `json:"cavaletes"`
	Blocos      []Bloco    `json:"blocos"`
	BlocosComChapas []BlocoComChapa `json:"blocosComChapas"`
	Chapas      []Chapa    `json:"chapas"`
	BlocosMarcados []BlocoMarcado `json:"blocosMarcados"`
}

// Cavalete representa um cavalete no sistema Mobgran
type Cavalete struct {
	NomeMaterial     string          `json:"nomeMaterial"`
	NomeEspessura    string          `json:"nomeEspessura"`
	Comprimento      float64         `json:"comprimento"`
	Altura           float64         `json:"altura"`
	ImagemPrincipal  *ImagemPrincipal `json:"imagemPrincipal,omitempty"`
	Codigo           string          `json:"codigo"`
	Bloco            string          `json:"bloco"`
	Metragem         float64         `json:"metragem"`
	Itens            []Item          `json:"itens"`
}

// Item representa um item dentro de um cavalete
type Item struct {
	NomeEspessura      string  `json:"nomeEspessura"`
	NomeClassificacao  string  `json:"nomeClassificacao"`
	Comprimento        float64 `json:"comprimento"`
	Altura             float64 `json:"altura"`
	Codigo             string  `json:"codigo"`
	Bloco              string  `json:"bloco"`
	Metragem           float64 `json:"metragem"`
}

// ImagemPrincipal representa as informações de imagem
type ImagemPrincipal struct {
	Nome   string `json:"nome"`
	URL    string `json:"url"`
	URLMin string `json:"urlMin"`
}

// Bloco representa um bloco no sistema
type Bloco struct {
	// Estrutura a ser definida conforme necessário
}

// BlocoComChapa representa um bloco com chapa
type BlocoComChapa struct {
	// Estrutura a ser definida conforme necessário
}

// Chapa representa uma chapa
type Chapa struct {
	// Estrutura a ser definida conforme necessário
}

// BlocoMarcado representa um bloco marcado
type BlocoMarcado struct {
	// Estrutura a ser definida conforme necessário
}

// Oferta representa uma oferta no banco de dados
type Oferta struct {
	ID             string                 `json:"id" db:"id"`
	UUIDLink       string                 `json:"uuid_link" db:"uuid_link"`
	Situacao       string                 `json:"situacao" db:"situacao"`
	NomeEmpresa    string                 `json:"nome_empresa" db:"nome_empresa"`
	URLLogo        string                 `json:"url_logo" db:"url_logo"`
	DadosCompletos map[string]interface{} `json:"dados_completos" db:"dados_completos"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// CavaleteDB representa um cavalete no banco de dados
type CavaleteDB struct {
	ID                string    `json:"id" db:"id"`
	OfertaID          string    `json:"oferta_id" db:"oferta_id"`
	Codigo            string    `json:"codigo" db:"codigo"`
	Bloco             string    `json:"bloco" db:"bloco"`
	NomeMaterial      string    `json:"nome_material" db:"nome_material"`
	NomeEspessura     string    `json:"nome_espessura" db:"nome_espessura"`
	NomeClassificacao string    `json:"nome_classificacao" db:"nome_classificacao"`
	NomeAcabamento    *string   `json:"nome_acabamento" db:"nome_acabamento"`
	Comprimento       *float64  `json:"comprimento" db:"comprimento"`
	Altura            *float64  `json:"altura" db:"altura"`
	Largura           *float64  `json:"largura" db:"largura"`
	Metragem          *float64  `json:"metragem" db:"metragem"`
	Peso              *float64  `json:"peso" db:"peso"`
	TipoMetragem      *string   `json:"tipo_metragem" db:"tipo_metragem"`
	Aprovado          bool      `json:"aprovado" db:"aprovado"`
	Importado         bool      `json:"importado" db:"importado"`
	DescricaoChapas   *string   `json:"descricao_chapas" db:"descricao_chapas"`
	QuantidadeItens   *int      `json:"quantidade_itens" db:"quantidade_itens"`
	Valor             *float64  `json:"valor" db:"valor"`
	Observacao        *string   `json:"observacao" db:"observacao"`
	ObservacaoConferencia *string `json:"observacao_conferencia" db:"observacao_conferencia"`
	ProdutoCliente    *string   `json:"produto_cliente" db:"produto_cliente"`
	EspessuraCliente  *string   `json:"espessura_cliente" db:"espessura_cliente"`
	ImagemPrincipal   map[string]interface{} `json:"imagem_principal" db:"imagem_principal"`
	ImagensAdicionais map[string]interface{} `json:"imagens_adicionais" db:"imagens_adicionais"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// ItemDB representa um item no banco de dados
type ItemDB struct {
	ID                    string     `json:"id" db:"id"`
	CavaleteID            string     `json:"cavalete_id" db:"cavalete_id"`
	Codigo                string     `json:"codigo" db:"codigo"`
	Bloco                 string     `json:"bloco" db:"bloco"`
	NomeEspessura         string     `json:"nome_espessura" db:"nome_espessura"`
	NomeClassificacao     string     `json:"nome_classificacao" db:"nome_classificacao"`
	NomeAcabamento        *string    `json:"nome_acabamento" db:"nome_acabamento"`
	Comprimento           *float64   `json:"comprimento" db:"comprimento"`
	Altura                *float64   `json:"altura" db:"altura"`
	Largura               *float64   `json:"largura" db:"largura"`
	Metragem              *float64   `json:"metragem" db:"metragem"`
	Peso                  *float64   `json:"peso" db:"peso"`
	TipoMetragem          *string    `json:"tipo_metragem" db:"tipo_metragem"`
	Aprovado              bool       `json:"aprovado" db:"aprovado"`
	Importado             bool       `json:"importado" db:"importado"`
	Valor                 *float64   `json:"valor" db:"valor"`
	Observacao            *string    `json:"observacao" db:"observacao"`
	ObservacaoConferencia *string    `json:"observacao_conferencia" db:"observacao_conferencia"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// ImportRequest representa uma requisição de importação
type ImportRequest struct {
	URL                string `json:"url" binding:"required"`
	AtualizarExistente bool   `json:"atualizar_existente"`
}

// ImportResponse representa a resposta de uma operação de importação
type ImportResponse struct {
	Sucesso   bool   `json:"sucesso"`
	Mensagem  string `json:"mensagem"`
	OfertaID  string `json:"oferta_id,omitempty"`
	UUIDLink  string `json:"uuid_link,omitempty"`
}