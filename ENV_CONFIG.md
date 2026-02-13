# Configuração de Ambiente (.env)

Este documento explica como configurar as variáveis de ambiente usando arquivo `.env`.

## Quick Start

### 1. Criar arquivo .env

Copie o arquivo de exemplo:

```bash
cp .env.example .env
```

### 2. Preencher credenciais

Edite o arquivo `.env` com suas credenciais do Banco do Brasil:

```env
BB_ENVIRONMENT=sandbox
BB_CLIENT_ID=seu-client-id-aqui
BB_CLIENT_SECRET=seu-client-secret-aqui
BB_DEV_APP_KEY=sua-developer-app-key-aqui
```

### 3. Executar testes

```bash
# Testes unitários (não precisam de credenciais)
go test ./... -short

# Testes E2E (usam credenciais do .env)
go test -v -tags=integration ./pix
```

## Variáveis Disponíveis

### Obrigatórias

| Variável | Descrição | Exemplo |
|----------|-----------|---------|
| `BB_ENVIRONMENT` | Ambiente da API | `sandbox`, `homolog`, `production` |
| `BB_CLIENT_ID` | OAuth2 Client ID | Fornecido pelo BB |
| `BB_CLIENT_SECRET` | OAuth2 Client Secret | Fornecido pelo BB |
| `BB_DEV_APP_KEY` | Developer Application Key | `gw-dev-app-key` (sandbox) ou `gw-app-key` (prod) |

### Opcionais

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `BB_TIMEOUT_SECONDS` | Timeout de requisições HTTP | `30` |
| `BB_RETRY_COUNT` | Número de tentativas em caso de falha | `3` |
| `BB_RETRY_DELAY_MS` | Delay entre tentativas (ms) | `100` |
| `BB_LOG_LEVEL` | Nível de log | `info` |
| `BB_LOG_FORMAT` | Formato do log | `text` ou `json` |
| `BB_PIX_KEY` | Chave PIX para testes | - |

## Ambientes Disponíveis

### Sandbox (Desenvolvimento)

Para testes e desenvolvimento:

```env
BB_ENVIRONMENT=sandbox
BB_DEV_APP_KEY=gw-dev-app-key
```

**URLs**:
- OAuth: `https://oauth.sandbox.bb.com.br/oauth/token`
- API: `https://api.sandbox.bb.com.br/pix-bb/v1`

**Características**:
- Sem impacto financeiro
- Pode não refletir comportamento de produção 100%
- Ideal para desenvolvimento e testes automatizados

### Homologação

Para testes de homologação:

```env
BB_ENVIRONMENT=homolog
BB_DEV_APP_KEY=gw-dev-app-key
```

**URLs**:
- OAuth: `https://oauth.hm.bb.com.br/oauth/token`
- API: `https://api.hm.bb.com.br/pix-bb/v1`

**Características**:
- Ambiente mais próximo da produção
- Usado para testes finais antes do deploy

### Produção

⚠️ **ATENÇÃO**: Use com cuidado! Ambiente real com impacto financeiro.

```env
BB_ENVIRONMENT=production
BB_DEV_APP_KEY=gw-app-key
```

**URLs**:
- OAuth: `https://oauth.bb.com.br/oauth/token`
- API: `https://api.bb.com.br/pix-bb/v1`

**Características**:
- Transações reais
- Impacto financeiro
- Requer credenciais de produção

## Como Obter Credenciais

### 1. Acessar Portal Developers BB

Visite: https://developers.bb.com.br

### 2. Criar Aplicação

1. Faça login
2. Acesse "Minhas Aplicações"
3. Clique em "Nova Aplicação"
4. Preencha os dados:
   - Nome da aplicação
   - Descrição
   - URL de callback (para OAuth)

### 3. Configurar APIs

1. Selecione a aplicação criada
2. Ative as APIs necessárias:
   - PIX
   - PIX Automático (se necessário)
3. Configure os escopos (scopes):
   - `cob.read`
   - `cob.write`
   - `pix.read`
   - `pix.write`
   - `webhook.read`
   - `webhook.write`

### 4. Obter Credenciais

Após criar a aplicação, você receberá:

- **Client ID**: Identificador único da aplicação
- **Client Secret**: Chave secreta (guarde com segurança!)
- **Developer Application Key**:
  - Sandbox: `gw-dev-app-key`
  - Produção: `gw-app-key`

## Uso em Código

### Carregar .env automaticamente

```go
import (
    "github.com/pericles-luz/go-bb-pix/internal/testutil"
)

func init() {
    // Carrega .env automaticamente
    _ = testutil.LoadEnv()
}
```

### Verificar se credenciais estão configuradas

```go
if !testutil.HasCredentials() {
    log.Fatal("Credenciais não configuradas. Crie arquivo .env")
}
```

### Obter configuração

```go
config := testutil.GetBBConfig()
fmt.Println("Ambiente:", config["environment"])
fmt.Println("Client ID:", config["client_id"])
```

### Obter variável específica

```go
// Com valor padrão
timeout := testutil.GetEnv("BB_TIMEOUT_SECONDS", "30")

// Obrigatória (retorna vazio se não existir)
clientID := testutil.GetRequiredEnv("BB_CLIENT_ID")

// Obrigatória (panic se não existir)
secret := testutil.MustGetEnv("BB_CLIENT_SECRET")
```

## Segurança

### ✅ Boas Práticas

1. **NUNCA** commite o arquivo `.env` no git
2. Mantenha `.env` no `.gitignore`
3. Use `.env.example` como template sem credenciais
4. Rotacione credenciais periodicamente
5. Use credenciais diferentes para cada ambiente
6. Restrinja acesso ao arquivo `.env` (chmod 600)

### ⚠️ O que NÃO fazer

1. ❌ Commitar `.env` no repositório
2. ❌ Compartilhar `.env` por email/chat
3. ❌ Usar mesmas credenciais em prod e dev
4. ❌ Hardcoded credentials no código
5. ❌ Expor `.env` em logs

### Proteger o arquivo .env

```bash
# Linux/Mac: Restringir permissões
chmod 600 .env

# Verificar se está no .gitignore
git check-ignore .env
# Deve retornar: .env
```

## Troubleshooting

### .env não está sendo carregado

**Problema**: Variáveis não são lidas do `.env`

**Soluções**:

1. Verificar se o arquivo existe:
   ```bash
   ls -la .env
   ```

2. Verificar se está no diretório correto:
   ```bash
   pwd
   cat .env
   ```

3. Verificar se LoadEnv() foi chamado:
   ```go
   testutil.LoadEnv()
   ```

4. Verificar logs:
   ```
   Loaded environment from: /path/to/.env
   ```

### Credenciais inválidas

**Problema**: Erro de autenticação

**Soluções**:

1. Verificar Client ID e Secret no portal BB
2. Confirmar ambiente correto (sandbox vs production)
3. Verificar se a aplicação está ativa no portal
4. Confirmar scopes configurados

### Testes E2E falhando

**Problema**: `Integration test credentials not configured`

**Soluções**:

1. Criar arquivo `.env`:
   ```bash
   cp .env.example .env
   ```

2. Preencher todas as variáveis obrigatórias

3. Verificar se credenciais estão corretas:
   ```bash
   grep BB_ .env
   ```

## Exemplos

### Exemplo .env completo (Sandbox)

```env
# Ambiente
BB_ENVIRONMENT=sandbox

# Credenciais OAuth2
BB_CLIENT_ID=eyJpZCI6IjEyMzQ1Njc4OTAiLCJjb2RQdWJsaWNhZG9yI
BB_CLIENT_SECRET=eyJpZCI6ImM5ODc2NTQzMjEwIiwiY29kUHVibGljYWRvcj
BB_DEV_APP_KEY=gw-dev-app-key

# Configurações opcionais
BB_TIMEOUT_SECONDS=30
BB_RETRY_COUNT=3
BB_RETRY_DELAY_MS=100
BB_LOG_LEVEL=debug
BB_LOG_FORMAT=json

# Chave PIX para testes
BB_PIX_KEY=sua-chave-pix@email.com
```

### Exemplo .env mínimo (Produção)

```env
BB_ENVIRONMENT=production
BB_CLIENT_ID=seu-client-id-producao
BB_CLIENT_SECRET=seu-client-secret-producao
BB_DEV_APP_KEY=gw-app-key
```

## CI/CD

### GitHub Actions

```yaml
env:
  BB_ENVIRONMENT: ${{ secrets.BB_ENVIRONMENT }}
  BB_CLIENT_ID: ${{ secrets.BB_CLIENT_ID }}
  BB_CLIENT_SECRET: ${{ secrets.BB_CLIENT_SECRET }}
  BB_DEV_APP_KEY: ${{ secrets.BB_DEV_APP_KEY }}

steps:
  - name: Run E2E Tests
    run: go test -v -tags=integration ./pix
```

### GitLab CI

```yaml
variables:
  BB_ENVIRONMENT: $BB_ENVIRONMENT
  BB_CLIENT_ID: $BB_CLIENT_ID
  BB_CLIENT_SECRET: $BB_CLIENT_SECRET
  BB_DEV_APP_KEY: $BB_DEV_APP_KEY

test:e2e:
  script:
    - go test -v -tags=integration ./pix
```

## Referências

- [Portal Developers BB](https://developers.bb.com.br)
- [Documentação API PIX](https://www.bb.com.br/site/developers/bb-como-servico/api-pix/)
- [godotenv (biblioteca Go)](https://github.com/joho/godotenv)
