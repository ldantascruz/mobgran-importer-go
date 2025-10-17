# 🚀 Mobgran Data Importer - Golang Edition

Uma API REST robusta e eficiente em Go para automatizar a extração e importação de dados da plataforma Mobgran para o Supabase.

## 📋 Pré-requisitos

- Go 1.21 ou superior
- Conta no Supabase com banco configurado
- Docker (opcional)
- Acesso aos links de conferência do Mobgran

## ⚡ Características

- 🎯 **API REST completa** com endpoints bem documentados
- 🚄 **Alta performance** com Gin framework
- 🔄 **Importação em lote** para múltiplas URLs
- 📊 **Estatísticas em tempo real**
- 🐳 **Docker ready** para deploy fácil
- 🛡️ **Tratamento robusto de erros**
- 📝 **Logs detalhados** para debugging
- ⚙️ **Configuração via variáveis de ambiente**

## 🏗️ Estrutura do Projeto

```
mobgran-importer-go/
├── main.go                 # Servidor principal e rotas
├── handlers/
│   └── mobgran_handler.go  # Controladores HTTP
├── services/
│   └── mobgran_importer.go # Lógica de negócio
├── models/
│   └── models.go           # Estruturas de dados
├── utils/
│   └── supabase.go        # Cliente Supabase
├── go.mod                  # Dependências Go
├── go.sum                  # Lock de dependências
├── .env.example           # Exemplo de configuração
├── Dockerfile             # Imagem Docker
├── docker-compose.yml     # Orquestração Docker
└── README.md              # Esta documentação
```

## 🚀 Instalação Rápida

### Opção 1: Instalação Local

1. **Clone ou crie o projeto:**
```bash
mkdir mobgran-importer-go
cd mobgran-importer-go
```

2. **Crie a estrutura de diretórios:**
```bash
mkdir -p handlers services models utils
```

3. **Copie todos os arquivos fornecidos para suas respectivas pastas**

4. **Instale as dependências:**
```bash
go mod init mobgran-importer
go mod tidy
```

5. **Configure o ambiente:**
```bash
cp .env.example .env
# Edite .env com suas credenciais do Supabase
```

6. **Execute o servidor:**
```bash
go run main.go
```

### Opção 2: Usando Docker

1. **Configure o ambiente:**
```bash
cp .env.example .env
# Edite .env com suas credenciais
```

2. **Construa e execute com Docker Compose:**
```bash
docker-compose up --build
```

## 🔧 Configuração

### Variáveis de Ambiente

Crie um arquivo `.env` baseado no `.env.example`:

```env
# Obrigatórias
SUPABASE_URL=https://seu-projeto.supabase.co
SUPABASE_KEY=sua-chave-api-aqui

# Opcionais
PORT=8080
LOG_LEVEL=info
MAX_REQUESTS_PER_MINUTE=60
MAX_BATCH_SIZE=50
```

### Como obter as credenciais do Supabase:
1. Acesse [Supabase](https://supabase.com)
2. Entre no seu projeto
3. Vá em **Settings** → **API**
4. Copie **Project URL** e **anon/public key**

## 📡 Endpoints da API

### Base URL
```
http://localhost:8080/api/v1
```

### 🔍 Health Check
```http
GET /health
```
**Resposta:**
```json
{
  "status": "ok",
  "message": "Mobgran Importer API está funcionando"
}
```

### 📥 Importar URL Única
```http
POST /mobgran/import
```
**Body:**
```json
{
  "url": "https://www.mobgran.com/app/conferencia/?p=link&o=uuid-xyz",
  "atualizar_existente": true
}
```
**Resposta de Sucesso (201):**
```json
{
  "sucesso": true,
  "mensagem": "✅ Oferta xyz123 importada com sucesso",
  "dados": {
    "uuid": "xyz123",
    "url": "https://...",
    "cavaletes": 15,
    "itens": 45,
    "metragem_total": 234.56,
    "tempo_processamento": 2.34,
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### 📥 Importar Múltiplas URLs
```http
POST /mobgran/import/batch
```
**Body:**
```json
{
  "urls": [
    "https://www.mobgran.com/app/conferencia/?p=link&o=uuid-1",
    "https://www.mobgran.com/app/conferencia/?p=link&o=uuid-2"
  ],
  "atualizar_existente": true
}
```
**Resposta:**
```json
{
  "resumo": {
    "total_processados": 2,
    "sucessos": 2,
    "falhas": 0,
    "metragem_total": 468.90,
    "cavaletes_total": 30
  },
  "detalhes": [...]
}
```

### 🔎 Verificar Oferta
```http
GET /mobgran/oferta/{uuid}
```
**Resposta:**
```json
{
  "sucesso": true,
  "oferta": {
    "uuid": "xyz123",
    "nome_empresa": "Empresa XYZ",
    "nome_vendedor": "João Silva",
    "data_importacao": "2024-01-15T10:30:00Z",
    "dados_completos": {...}
  }
}
```

### 📋 Listar Ofertas
```http
GET /mobgran/ofertas?limit=20&offset=0
```
**Resposta:**
```json
{
  "sucesso": true,
  "dados": {
    "ofertas": [...],
    "paginacao": {
      "limit": 20,
      "offset": 0,
      "total": 15,
      "proxima_pagina": 20
    }
  }
}
```

### 📊 Estatísticas
```http
GET /mobgran/stats
```
**Resposta:**
```json
{
  "sucesso": true,
  "estatisticas": {
    "total_ofertas": 150,
    "total_cavaletes": 2340,
    "metragem_total": 45678.90,
    "material_mais_comum": "WHITE DALLAS",
    "distribuicao_materiais": {
      "WHITE DALLAS": 45,
      "CALACATTA": 32,
      "NERO MARQUINA": 28
    }
  }
}
```

## 💻 Exemplos de Uso

### cURL

```bash
# Importar uma URL
curl -X POST http://localhost:8080/api/v1/mobgran/import \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.mobgran.com/app/conferencia/?p=link&o=uuid-xyz"}'

# Listar ofertas
curl http://localhost:8080/api/v1/mobgran/ofer