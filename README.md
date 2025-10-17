# Mobgran Importer - Go

Uma API REST desenvolvida em Go para importar dados de ofertas do Mobgran para o Supabase. Esta é uma reimplementação em Go do projeto Python original, oferecendo melhor performance e facilidade de deploy.

## 🚀 Características

- **API REST** com endpoints para importação e validação
- **Cliente Supabase** integrado para persistência de dados
- **Validação de URLs** do Mobgran
- **Logging estruturado** com diferentes níveis
- **Containerização** com Docker
- **Configuração flexível** via variáveis de ambiente
- **Health checks** para monitoramento

## 📋 Pré-requisitos

- Go 1.21 ou superior
- Conta no Supabase configurada
- Docker (opcional, para containerização)

## 🛠️ Instalação

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

3. Configure as variáveis de ambiente:
```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações
```

4. Execute a aplicação:
```bash
go run cmd/server/main.go
```

### Docker

1. Build da imagem:
```bash
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
| `SUPABASE_URL` | URL do projeto Supabase | **Obrigatório** |
| `SUPABASE_KEY` | Chave de API do Supabase | **Obrigatório** |
| `LOG_LEVEL` | Nível de log (debug, info, warn, error) | `info` |
| `MOBGRAN_API_URL` | URL base da API Mobgran | `https://api.mobgran.com.br/api/v1/ofertas/` |

### Exemplo de .env

```env
PORT=8080
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-supabase-anon-key
LOG_LEVEL=info
MOBGRAN_API_URL=https://api.mobgran.com.br/api/v1/ofertas/
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
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "mobgran-importer-go",
  "version": "1.0.0"
}
```

### Importar Oferta

```http
POST /api/importar
```

Importa uma oferta do Mobgran para o Supabase.

**Body:**
```json
{
  "url": "https://www.mobgran.com/app/conferencia/?p=link&o=uuid-here",
  "atualizar_existente": false
}
```

**Resposta:**
```json
{
  "sucesso": true,
  "mensagem": "Oferta importada com sucesso",
  "uuid": "uuid-da-oferta"
}
```

### Validar URL

```http
POST /api/validar-url
```

Valida se uma URL é um link válido do Mobgran.

**Body:**
```json
{
  "url": "https://www.mobgran.com/app/conferencia/?p=link&o=uuid-here"
}
```

**Resposta:**
```json
{
  "valida": true,
  "mensagem": "URL válida",
  "uuid": "uuid-extraido"
}
```

### Extrair UUID

```http
POST /api/extrair-uuid
```

Extrai o UUID de uma URL do Mobgran.

**Body:**
```json
{
  "url": "https://www.mobgran.com/app/conferencia/?p=link&o=uuid-here"
}
```

**Resposta:**
```json
{
  "sucesso": true,
  "uuid": "uuid-extraido"
}
```

## 🏗️ Arquitetura

```
mobgran-importer-go/
├── cmd/
│   └── server/          # Ponto de entrada da aplicação
├── internal/
│   ├── config/          # Configurações e setup
│   ├── handlers/        # Handlers HTTP
│   ├── models/          # Estruturas de dados
│   └── services/        # Lógica de negócio
├── pkg/
│   └── supabase/        # Cliente Supabase
└── docs/                # Documentação
```

### Componentes Principais

- **Config**: Gerenciamento de configurações e variáveis de ambiente
- **Handlers**: Controladores HTTP que processam as requisições
- **Models**: Estruturas de dados para representar ofertas, cavaletes e itens
- **Services**: Lógica de importação e processamento de dados
- **Supabase Client**: Interface para comunicação com o Supabase

## 🔄 Fluxo de Importação

1. **Validação da URL**: Verifica se a URL é um link válido do Mobgran
2. **Extração do UUID**: Extrai o identificador único da oferta
3. **Busca de dados**: Faz requisição para a API do Mobgran
4. **Verificação de existência**: Checa se a oferta já existe no Supabase
5. **Processamento**: Salva ou atualiza a oferta, cavaletes e itens
6. **Resposta**: Retorna o resultado da operação

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

A aplicação expõe um endpoint de health check em `/health` que pode ser usado para:

- Load balancers
- Kubernetes health checks
- Monitoramento de uptime
- CI/CD pipelines

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 🆚 Diferenças da Versão Python

- **Performance**: Melhor performance devido à natureza compilada do Go
- **Concorrência**: Melhor suporte nativo para operações concorrentes
- **Deploy**: Binário único, sem dependências externas
- **Memória**: Menor uso de memória em produção
- **API REST**: Interface HTTP completa (vs CLI na versão Python)

## 📞 Suporte

Para suporte e dúvidas, abra uma issue no repositório do projeto.