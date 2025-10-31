# Mobgran Importer - Go

Uma API REST desenvolvida em Go para gerenciar traders e ofertas do Mobgran com PostgreSQL. Esta implementação oferece um sistema completo de autenticação e gerenciamento de dados com migrations automáticas.

## 🚀 Características

- **API REST** com endpoints para autenticação e gerenciamento de traders
- **PostgreSQL** como banco de dados principal com migrations automáticas
- **Sistema de Autenticação** completo com JWT e refresh tokens
- **Estrutura de Dados** robusta para traders, ofertas, cavaletes e produtos
- **Logging estruturado** com diferentes níveis
- **Containerização** com Docker e hot reload
- **Configuração flexível** via variáveis de ambiente
- **Health checks** para monitoramento

## 📋 Pré-requisitos

- Go 1.24 ou superior
- PostgreSQL 15+ (ou Docker para desenvolvimento)
- Docker e Docker Compose (recomendado para desenvolvimento)

## 🛠️ Instalação

### Desenvolvimento com Docker (Recomendado)

1. Clone o repositório:
```bash
git clone <repository-url>
cd mobgran-importer-go
```

2. Configure as variáveis de ambiente:
```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações
```

3. Execute com Docker Compose:
```bash
docker-compose up --watch
```

A aplicação estará disponível em `http://localhost:8080` com hot reload ativo.

### Desenvolvimento Local

1. Clone o repositório:
```bash
git clone <repository-url>
cd mobgran-importer-go
```

2. Instale as dependências:
```bash
go mod download
```

3. Configure PostgreSQL local e as variáveis de ambiente:
```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações do PostgreSQL
```

4. Execute a aplicação:
```bash
go run cmd/server/main.go
```
docker build -t mobgran-importer-go .
```

2. Execute o container:
```bash
docker run -p 8080:8080 --env-file .env mobgran-importer-go
```

### Docker Compose

```bash
docker-compose up -d
```

## ⚙️ Configuração

### Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|---------|
| `PORT` | Porta do servidor | `8080` |
| `LOG_LEVEL` | Nível de log (debug, info, warn, error) | `info` |
| `DB_HOST` | Host do PostgreSQL | `localhost` |
| `DB_PORT` | Porta do PostgreSQL | `5433` |
| `DB_NAME` | Nome do banco de dados | `mobgran_db` |
| `DB_USER` | Usuário do PostgreSQL | `mobgran_user` |
| `DB_PASSWORD` | Senha do PostgreSQL | **Obrigatório** |
| `DB_SSLMODE` | Modo SSL do PostgreSQL | `disable` |
| `JWT_SECRET` | Chave secreta para JWT | **Obrigatório** |
| `JWT_EXPIRATION` | Expiração do JWT em horas | `24` |
| `JWT_REFRESH_EXPIRATION` | Expiração do refresh token em horas | `168` |
| `JWT_ISSUER` | Emissor do JWT | `mobgran-api` |
| `ENVIRONMENT` | Ambiente da aplicação | `development` |
| `CORS_ALLOWED_ORIGINS` | Origens permitidas para CORS | `*` |

### Exemplo de .env

```env
# Servidor
PORT=8080
LOG_LEVEL=debug
ENVIRONMENT=development

# PostgreSQL
DB_HOST=localhost
DB_PORT=5433
DB_NAME=mobgran_db
DB_USER=mobgran_user
DB_PASSWORD=sua_senha_segura
DB_SSLMODE=disable

# JWT
JWT_SECRET=sua_chave_jwt_muito_segura_aqui
JWT_EXPIRATION=24
JWT_REFRESH_EXPIRATION=168
JWT_ISSUER=mobgran-api

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
```

## 📚 API Endpoints

### Health Check

```http
GET /health
```

Retorna o status da aplicação.

**Resposta:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Autenticação

#### Registrar Trader

```http
POST /auth/registrar
```

Registra um novo trader no sistema.

**Body:**
```json
{
  "nome": "João Silva",
  "email": "joao@exemplo.com",
  "senha": "senha123",
  "telefone": "+5511999999999",
  "cidade": "São Paulo",
  "estado": "SP"
}
```

#### Login

```http
POST /auth/login
```

Realiza login de um trader.

**Body:**
```json
{
  "email": "joao@exemplo.com",
  "senha": "senha123"
}
```

**Resposta:**
```json
{
  "access_token": "jwt-token-here",
  "refresh_token": "refresh-token-here",
  "trader": {
    "id": "uuid",
    "nome": "João Silva",
    "email": "joao@exemplo.com"
  }
}
```

#### Refresh Token

```http
POST /auth/refresh
```

Renova o token de acesso usando o refresh token.

**Body:**
```json
{
  "refresh_token": "refresh-token-here"
}
```

#### Logout

```http
POST /auth/logout
```

Realiza logout invalidando o refresh token.

**Headers:**
```
Authorization: Bearer jwt-token-here
```

### Traders (Autenticado)

#### Buscar Trader por ID

```http
GET /auth/trader/:id
```

#### Atualizar Trader

```http
PUT /auth/trader/:id
```

#### Alterar Senha

```http
PUT /auth/alterar-senha
```

#### Listar Traders

```http
GET /auth/traders
```

## 🏗️ Arquitetura

```
mobgran-importer-go/
├── cmd/
│   └── server/          # Ponto de entrada da aplicação
├── internal/
│   ├── config/          # Configurações e setup
│   ├── handlers/        # Handlers HTTP
│   ├── middleware/      # Middlewares (autenticação, CORS, etc.)
│   ├── models/          # Estruturas de dados
│   └── services/        # Lógica de negócio
├── pkg/
│   └── database/        # Cliente PostgreSQL e migrations
└── docs/                # Documentação Swagger
```

### Componentes Principais

- **Config**: Gerenciamento de configurações e variáveis de ambiente
- **Handlers**: Controladores HTTP que processam as requisições
- **Middleware**: Autenticação JWT, CORS, rate limiting e logging
- **Models**: Estruturas de dados para traders, ofertas, cavaletes e produtos
- **Services**: Lógica de autenticação e gerenciamento de dados
- **Database**: Cliente PostgreSQL com migrations automáticas

## 🔄 Fluxo de Autenticação

1. **Registro**: Trader se registra fornecendo dados pessoais
2. **Validação**: Sistema valida dados e cria hash da senha
3. **Login**: Trader faz login com email e senha
4. **JWT**: Sistema gera access token e refresh token
5. **Autorização**: Requests protegidos usam JWT no header
6. **Refresh**: Token expirado pode ser renovado com refresh token

## 🗄️ Banco de Dados

### Migrations Automáticas

O sistema executa migrations automaticamente na inicialização:

- `001_create_tables.sql`: Cria tabelas principais
- `002_create_indexes.sql`: Cria índices para performance
- `003_create_views.sql`: Cria views para consultas complexas

### Estrutura Principal

- **traders**: Dados dos traders (usuários)
- **refresh_tokens**: Tokens de refresh para autenticação
- **ofertas**: Ofertas do Mobgran
- **cavaletes**: Cavaletes disponíveis
- **produtos**: Produtos e itens
- **schema_migrations**: Controle de versão das migrations

## 🧪 Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes com cobertura
go test -cover ./...

# Executar testes de um pacote específico
go test ./internal/services
```

## 📦 Build

### Build local

```bash
go build -o mobgran-importer ./cmd/server
```

### Build para produção

```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mobgran-importer ./cmd/server
```

## 🚀 Deploy

### Docker

```bash
docker build -t mobgran-importer-go .
docker run -p 8080:8080 --env-file .env mobgran-importer-go
```

### Docker Compose

```bash
docker-compose up -d
```

## 📊 Monitoramento

### Health Check

A aplicação expõe um endpoint de health check em `/health` que retorna:

- Status da aplicação
- Timestamp atual
- Conectividade com PostgreSQL

### Logs

A aplicação usa logging estruturado com níveis configuráveis:

- `debug`: Informações detalhadas para desenvolvimento
- `info`: Informações gerais de operação
- `warn`: Avisos que não impedem a operação
- `error`: Erros que requerem atenção

### Swagger Documentation

Acesse a documentação interativa da API em:
```
http://localhost:8080/swagger/index.html
```

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 🔧 Desenvolvimento

### Estrutura de Commits

- `feat:` Nova funcionalidade
- `fix:` Correção de bug
- `docs:` Documentação
- `style:` Formatação
- `refactor:` Refatoração de código
- `test:` Testes
- `chore:` Tarefas de manutenção

### Padrões de Código

- Use `gofmt` para formatação
- Execute `go vet` para análise estática
- Mantenha cobertura de testes acima de 80%
- Documente funções públicas

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## ✨ Características da Implementação PostgreSQL

- **Migrations Automáticas**: Sistema de versionamento de banco de dados
- **Connection Pool**: Gerenciamento eficiente de conexões
- **Transações**: Operações atômicas para consistência de dados
- **Índices Otimizados**: Performance aprimorada para consultas
- **Autenticação JWT**: Sistema seguro de autenticação
- **Hot Reload**: Desenvolvimento com recarga automática via Air
- **API REST**: Interface HTTP completa (vs CLI na versão Python)

## 📞 Suporte

Para suporte e dúvidas, abra uma issue no repositório do projeto.