# go-bb-pix

Pacote Go para integração com as APIs de PIX e PIX Automático do Banco do Brasil.

## Visão Geral

Este projeto implementa um cliente completo para as APIs de PIX e PIX Automático do Banco do Brasil, seguindo as melhores práticas de desenvolvimento em Go:

- **TDD (Test-Driven Development)**: Todos os componentes são desenvolvidos com testes primeiro
- **Arquitetura em camadas**: Separação clara de responsabilidades
- **Resiliência**: Retry, circuit breaker e timeout configuráveis
- **Observabilidade**: Logging estruturado com slog
- **Zero dependências**: Usa apenas a standard library do Go

## Estrutura do Projeto

```
github.com/pericles-luz/go-bb-pix/
├── bbpix/                          # Package principal
│   ├── client.go                   # Cliente principal
│   ├── config.go                   # Configuração e ambientes
│   ├── options.go                  # Functional options
│   └── errors.go                   # Tipos de erro personalizados
│
├── internal/                       # Implementações privadas
│   ├── auth/                       # OAuth2
│   ├── transport/                  # Camada HTTP com middlewares
│   ├── http/                       # Cliente HTTP base
│   └── testutil/                   # Utilitários de teste
│
├── pix/                            # Operações PIX
├── pixauto/                        # PIX Automático
├── webhook/                        # Tratamento de webhooks
├── examples/                       # Exemplos de uso
└── testdata/                       # Fixtures de teste
```

## Ambientes da API

### Sandbox
- OAuth: `https://oauth.sandbox.bb.com.br/oauth/token`
- API: `https://api.sandbox.bb.com.br/pix-bb/v1`

### Homologação
- OAuth: `https://oauth.hm.bb.com.br/oauth/token`
- API: `https://api.hm.bb.com.br/pix-bb/v1`

### Produção
- OAuth: `https://oauth.bb.com.br/oauth/token`
- API: `https://api.bb.com.br/pix-bb/v1`

## Credenciais Necessárias

- `developer_application_key`: Chave da aplicação (gw-dev-app-key para sandbox, gw-app-key para produção)
- `client_id`: ID do cliente OAuth2
- `client_secret`: Secret do cliente OAuth2

## Como Usar

### Instalação

```bash
go get github.com/pericles-luz/go-bb-pix
```

### Quick Start

```go
import (
    "context"
    "github.com/pericles-luz/go-bb-pix/bbpix"
    "log/slog"
)

// Configurar cliente
config := bbpix.Config{
    Environment:    bbpix.EnvironmentSandbox,
    ClientID:       "seu-client-id",
    ClientSecret:   "seu-client-secret",
    DeveloperAppKey: "sua-app-key",
}

// Criar cliente com options
client, err := bbpix.New(config,
    bbpix.WithLogger(slog.Default()),
    bbpix.WithTimeout(30 * time.Second),
)
if err != nil {
    log.Fatal(err)
}

// Usar PIX
pixClient := client.PIX()
qrCode, err := pixClient.CreateQRCode(context.Background(), request)
```

## Executar Testes

### Testes Unitários (rápidos)

```bash
go test ./... -short -cover
```

### Testes de Integração (requer credenciais)

```bash
# Configurar variáveis de ambiente
export BB_DEV_APP_KEY=xxx
export BB_CLIENT_ID=xxx
export BB_CLIENT_SECRET=xxx
export BB_ENVIRONMENT=sandbox

# Executar testes
go test -v -tags=integration ./...
```

### Pre-commit Checks

```bash
./scripts/pre-commit.sh
```

Executa:
- `go test ./... -short`
- `go mod tidy`
- `go vet ./...`
- `staticcheck ./...`
- `go build ./...`

## Padrões de Código

### 1. Context-First
Todos os métodos públicos aceitam `context.Context` como primeiro parâmetro:

```go
func (c *Client) CreateQRCode(ctx context.Context, req CreateQRCodeRequest) (*QRCodeResponse, error)
```

### 2. Functional Options
Configuração opcional usando o padrão functional options:

```go
client, err := bbpix.New(config,
    bbpix.WithLogger(logger),
    bbpix.WithTimeout(30*time.Second),
    bbpix.WithRetry(3, 100*time.Millisecond),
)
```

### 3. Error Wrapping
Erros são sempre wrapped com contexto:

```go
if err != nil {
    return nil, fmt.Errorf("failed to create qr code: %w", err)
}
```

### 4. Structured Logging
Usar slog para logging estruturado:

```go
logger.InfoContext(ctx, "creating qr code",
    slog.String("txid", req.TxID),
    slog.Float64("value", req.Value),
)
```

### 5. Table-Driven Tests
Preferir table-driven tests para múltiplos cenários:

```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {name: "success", input: "valid", want: "result", wantErr: false},
    {name: "error", input: "invalid", want: "", wantErr: true},
}
```

## Arquitetura de Transport

O cliente HTTP usa uma arquitetura em camadas (middleware pattern):

```
HTTP Request
    ↓
[Logging Transport] → Log request/response
    ↓
[Auth Transport] → Inject OAuth2 token
    ↓
[Retry Transport] → Retry on transient errors
    ↓
[Circuit Breaker Transport] → Fail-fast protection
    ↓
[Base HTTP Transport]
    ↓
API do Banco do Brasil
```

### Retry Strategy
- Apenas métodos idempotentes (GET, PUT, DELETE, HEAD, OPTIONS)
- Exponential backoff com jitter
- Retry em: erros de rede, 429, 502, 503, 504
- Não retry em: POST, PATCH, 4xx (exceto 429)

### Circuit Breaker
- Estados: Closed → Open → Half-Open → Closed
- Proteção contra cascata de falhas
- Fail-fast quando API está indisponível

## Operações Suportadas

### PIX
- ✅ Criar QR Code estático/dinâmico
- ✅ Consultar QR Code
- ✅ Atualizar QR Code
- ✅ Listar QR Codes
- ✅ Consultar pagamento
- ✅ Listar pagamentos
- ✅ Criar devolução
- ✅ Consultar devolução

### PIX Automático
- ✅ Criar cobrança recorrente
- ✅ Consultar cobrança recorrente
- ✅ Atualizar cobrança recorrente
- ✅ Cancelar cobrança recorrente
- ✅ Listar cobranças recorrentes
- ✅ Criar cobrança agendada (até 90 dias)
- ✅ Consultar cobrança agendada
- ✅ Cancelar cobrança agendada
- ✅ Listar cobranças agendadas
- ✅ Criar acordo de débito
- ✅ Consultar acordo de débito
- ✅ Atualizar acordo de débito
- ✅ Cancelar acordo de débito
- ✅ Listar acordos de débito

### Webhooks
- ✅ Handler HTTP
- ✅ Validação de assinatura
- ✅ Roteamento de eventos
- ✅ Registro de handlers customizados

## Exemplos

Veja a pasta `examples/` para exemplos completos:

- `examples/pix_qrcode/main.go`: Criar e consultar QR Code
- `examples/pixauto_recurring/main.go`: Criar cobrança recorrente
- `examples/webhook_server/main.go`: Servidor webhook

## Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/amazing-feature`)
3. Escreva testes primeiro (TDD)
4. Implemente a feature
5. Execute os testes (`go test ./...`)
6. Execute pre-commit checks (`./scripts/pre-commit.sh`)
7. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
8. Push para a branch (`git push origin feature/amazing-feature`)
9. Abra um Pull Request

## Licença

MIT License - veja LICENSE para detalhes

## Recursos

- [Documentação API PIX BB](https://www.bb.com.br/site/developers/bb-como-servico/api-pix/)
- [Especificação PIX - Banco Central](https://github.com/bacen/pix-api)
- [Go Client Library Best Practices](https://medium.com/@cep21/go-client-library-best-practices-83d877d604ca)

## Suporte

Para bugs e feature requests, abra uma issue no GitHub.
