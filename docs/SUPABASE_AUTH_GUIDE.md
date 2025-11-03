# Guia de Autenticação Supabase

Este documento fornece um guia completo sobre a implementação de autenticação Supabase no projeto Mobgran Importer.

## Índice

1. [Visão Geral](#visão-geral)
2. [Configuração](#configuração)
3. [Endpoints de Autenticação](#endpoints-de-autenticação)
4. [Middleware de Autenticação](#middleware-de-autenticação)
5. [Estrutura de Dados](#estrutura-de-dados)
6. [Exemplos de Uso](#exemplos-de-uso)
7. [Tratamento de Erros](#tratamento-de-erros)
8. [Segurança](#segurança)

## Visão Geral

O sistema de autenticação utiliza o Supabase Auth como provedor de identidade, oferecendo:

- ✅ Criação de usuários administrativos pré-confirmados
- ✅ Registro de usuários regulares
- ✅ Login com email e senha
- ✅ Renovação de tokens JWT
- ✅ Logout seguro
- ✅ Middleware de proteção de rotas
- ✅ Validação de tokens JWT

## Configuração

### Variáveis de Ambiente

Configure as seguintes variáveis no arquivo `.env`:

```env
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-role-key
SUPABASE_JWT_SECRET=your-jwt-secret
```

### Estrutura de Arquivos

```
internal/
├── handlers/
│   └── supabase_auth.go      # Handlers de autenticação
├── services/
│   └── supabase_auth.go      # Lógica de negócio
├── middleware/
│   └── auth.go               # Middleware de autenticação
└── models/
    └── auth.go               # Modelos de dados

pkg/
└── supabase/
    ├── client.go             # Cliente Supabase
    └── auth.go               # Funções de autenticação
```

## Endpoints de Autenticação

### 1. Criar Usuário Admin

**POST** `/supabase/auth/admin/create`

Cria um usuário administrativo pré-confirmado.

**Request Body:**
```json
{
  "email": "admin@exemplo.com",
  "password": "senha123456",
  "data": {
    "role": "admin",
    "name": "Administrador"
  }
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "admin@exemplo.com",
    "email_confirmed_at": "2023-11-03T19:00:00Z",
    "user_metadata": {
      "role": "admin",
      "name": "Administrador"
    }
  },
  "session": {
    "access_token": "jwt-token",
    "refresh_token": "refresh-token",
    "expires_in": 3600,
    "token_type": "bearer"
  }
}
```

### 2. Registrar Usuário

**POST** `/supabase/auth/register`

Registra um novo usuário regular (requer confirmação por email).

**Request Body:**
```json
{
  "email": "usuario@exemplo.com",
  "password": "senha123456",
  "data": {
    "name": "Nome do Usuário"
  }
}
```

### 3. Login

**POST** `/supabase/auth/login`

Autentica um usuário existente.

**Request Body:**
```json
{
  "email": "usuario@exemplo.com",
  "password": "senha123456"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "usuario@exemplo.com",
    "email_confirmed_at": "2023-11-03T19:00:00Z"
  },
  "session": {
    "access_token": "jwt-token",
    "refresh_token": "refresh-token",
    "expires_in": 3600,
    "token_type": "bearer"
  }
}
```

### 4. Obter Usuário

**GET** `/supabase/auth/user`

Retorna informações do usuário autenticado.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "usuario@exemplo.com",
    "email_confirmed_at": "2023-11-03T19:00:00Z",
    "user_metadata": {
      "name": "Nome do Usuário"
    }
  }
}
```

### 5. Renovar Token

**POST** `/supabase/auth/refresh`

Renova o token de acesso usando o refresh token.

**Request Body:**
```json
{
  "refresh_token": "refresh-token"
}
```

### 6. Logout

**POST** `/supabase/auth/logout`

Invalida a sessão do usuário.

**Headers:**
```
Authorization: Bearer <access_token>
```

## Middleware de Autenticação

### SupabaseAuthMiddleware

O middleware `SupabaseAuthMiddleware()` protege rotas que requerem autenticação:

```go
// Exemplo de uso
produtos := router.Group("/produtos")
{
    produtos.GET("/", middleware.SupabaseAuthMiddleware(), handler.ListarProdutos)
    produtos.POST("/", middleware.SupabaseAuthMiddleware(), handler.CriarProduto)
}
```

### Funcionamento

1. **Extração do Token**: Extrai o token do header `Authorization: Bearer <token>`
2. **Validação JWT**: Valida a assinatura usando `SUPABASE_JWT_SECRET`
3. **Contexto**: Adiciona informações do usuário ao contexto da requisição:
   - `user_id`: ID do usuário
   - `user_email`: Email do usuário
   - `user_role`: Role do usuário

### Acessando Dados do Usuário

```go
func (h *Handler) MinhaFuncao(c *gin.Context) {
    userID := c.GetString("user_id")
    userEmail := c.GetString("user_email")
    userRole := c.GetString("user_role")
    
    // Usar os dados...
}
```

## Estrutura de Dados

### SupabaseAuthResponse

```go
type SupabaseAuthResponse struct {
    User    *SupabaseUser    `json:"user"`
    Session *SupabaseSession `json:"session"`
}
```

### SupabaseUser

```go
type SupabaseUser struct {
    ID                string                 `json:"id"`
    Email             string                 `json:"email"`
    EmailConfirmedAt  *time.Time            `json:"email_confirmed_at"`
    UserMetadata      map[string]interface{} `json:"user_metadata"`
    AppMetadata       map[string]interface{} `json:"app_metadata"`
    CreatedAt         time.Time             `json:"created_at"`
    UpdatedAt         time.Time             `json:"updated_at"`
}
```

### SupabaseSession

```go
type SupabaseSession struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    TokenType    string `json:"token_type"`
}
```

## Exemplos de Uso

### Frontend - Login

```javascript
// Login
const response = await fetch('/supabase/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    email: 'usuario@exemplo.com',
    password: 'senha123456'
  })
});

const data = await response.json();
const accessToken = data.session.access_token;

// Armazenar token
localStorage.setItem('access_token', accessToken);
```

### Frontend - Requisições Autenticadas

```javascript
// Fazer requisição autenticada
const token = localStorage.getItem('access_token');

const response = await fetch('/produtos', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});
```

### cURL - Criar Admin

```bash
curl -X POST http://localhost:8080/supabase/auth/admin/create \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@exemplo.com",
    "password": "senha123456",
    "data": {
      "role": "admin",
      "name": "Administrador"
    }
  }'
```

### cURL - Login

```bash
curl -X POST http://localhost:8080/supabase/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@exemplo.com",
    "password": "senha123456"
  }'
```

## Tratamento de Erros

### Códigos de Status HTTP

- **200**: Sucesso
- **201**: Criado com sucesso
- **400**: Dados inválidos
- **401**: Não autorizado
- **409**: Conflito (email já existe)
- **500**: Erro interno do servidor

### Formato de Erro

```json
{
  "error": {
    "type": "authentication_error",
    "message": "Token inválido ou expirado",
    "details": "Detalhes adicionais do erro"
  }
}
```

### Tipos de Erro

- `validation_error`: Dados de entrada inválidos
- `authentication_error`: Problemas de autenticação
- `authorization_error`: Falta de permissões
- `conflict_error`: Recurso já existe
- `not_found_error`: Recurso não encontrado
- `internal_error`: Erro interno do servidor

## Segurança

### Boas Práticas Implementadas

1. **Tokens JWT**: Assinados com chave secreta forte
2. **HTTPS**: Recomendado para produção
3. **Headers de Segurança**: CSP, XSS Protection, etc.
4. **CORS**: Configurado para origens específicas
5. **Validação**: Entrada validada em todos os endpoints
6. **Logs**: Eventos de autenticação registrados

### Configurações de Segurança

```go
// Headers de segurança aplicados automaticamente
c.Header("X-Content-Type-Options", "nosniff")
c.Header("X-Frame-Options", "DENY")
c.Header("X-XSS-Protection", "1; mode=block")
c.Header("Content-Security-Policy", "default-src 'self'")
```

### Recomendações

1. **Rotação de Chaves**: Altere `SUPABASE_JWT_SECRET` periodicamente
2. **Monitoramento**: Monitore tentativas de login falhadas
3. **Rate Limiting**: Implemente limitação de taxa para endpoints de auth
4. **Auditoria**: Registre eventos de autenticação importantes
5. **Backup**: Mantenha backup das configurações de autenticação

## Migração do Sistema Anterior

### Mudanças Principais

- ❌ **Removido**: Sistema de autenticação tradicional (`/auth/*`)
- ❌ **Removido**: `AuthMiddleware()` tradicional
- ❌ **Removido**: Handlers e serviços de auth tradicionais
- ✅ **Adicionado**: Integração completa com Supabase
- ✅ **Adicionado**: `SupabaseAuthMiddleware()`
- ✅ **Atualizado**: Todas as rotas protegidas usam novo middleware

### Checklist de Migração

- [x] Configurar variáveis de ambiente do Supabase
- [x] Remover código de autenticação tradicional
- [x] Atualizar middleware das rotas protegidas
- [x] Testar todos os endpoints de autenticação
- [x] Atualizar documentação Swagger
- [x] Criar documentação de migração

## Troubleshooting

### Problemas Comuns

1. **Token inválido**: Verifique `SUPABASE_JWT_SECRET`
2. **CORS errors**: Adicione origem à lista permitida
3. **Email não confirmado**: Use endpoint admin para criar usuários pré-confirmados
4. **Refresh token expirado**: Implemente renovação automática no frontend

### Logs Úteis

```bash
# Verificar logs do servidor
tail -f logs/app.log | grep "authentication"

# Verificar configuração
env | grep SUPABASE
```

---

**Última atualização**: 03/11/2023  
**Versão**: 1.0.0  
**Autor**: Sistema de Autenticação Supabase