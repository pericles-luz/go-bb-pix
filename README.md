# go-bb-pix

[![Go Reference](https://pkg.go.dev/badge/github.com/pericles-luz/go-bb-pix.svg)](https://pkg.go.dev/github.com/pericles-luz/go-bb-pix)
[![Go Report Card](https://goreportcard.com/badge/github.com/pericles-luz/go-bb-pix)](https://goreportcard.com/report/github.com/pericles-luz/go-bb-pix)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Cliente Go completo para as APIs de PIX e PIX Automático do Banco do Brasil.

## Características

- ✅ **Zero dependências externas** - Usa apenas a standard library do Go
- ✅ **TDD** - Desenvolvido com Test-Driven Development
- ✅ **Resiliência** - Retry automático, circuit breaker e timeout configurável
- ✅ **Observabilidade** - Logging estruturado com slog
- ✅ **OAuth2** - Gerenciamento automático de tokens com cache
- ✅ **Context-aware** - Suporte completo a context para cancelamento e timeout
- ✅ **Type-safe** - Tipos fortemente tipados para todas as operações

## Instalação

```bash
go get github.com/pericles-luz/go-bb-pix
```

Requisitos: Go 1.21+

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/pericles-luz/go-bb-pix/bbpix"
)

func main() {
    // Configurar cliente
    config := bbpix.Config{
        Environment:     bbpix.EnvironmentSandbox,
        ClientID:        "seu-client-id",
        ClientSecret:    "seu-client-secret",
        DeveloperAppKey: "sua-app-key",
    }

    // Criar cliente
    client, err := bbpix.New(config)
    if err != nil {
        log.Fatal(err)
    }

    // Criar QR Code PIX
    pixClient := client.PIX()
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    qrCode, err := pixClient.CreateQRCode(ctx, pix.CreateQRCodeRequest{
        TxID:       "unique-transaction-id",
        Value:      100.50,
        Expiration: 3600, // 1 hora
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("QR Code criado: %s", qrCode.QRCode)
}
```

## Configuração

### Variáveis de Ambiente

Você pode configurar o cliente usando variáveis de ambiente:

```bash
export BB_ENVIRONMENT=sandbox      # sandbox, homologacao, producao
export BB_CLIENT_ID=seu-client-id
export BB_CLIENT_SECRET=seu-client-secret
export BB_DEV_APP_KEY=sua-app-key
```

```go
config, err := bbpix.LoadConfigFromEnv()
if err != nil {
    log.Fatal(err)
}

client, err := bbpix.New(config)
```

### Options

Customize o comportamento do cliente usando functional options:

```go
import "log/slog"

client, err := bbpix.New(config,
    bbpix.WithLogger(slog.Default()),
    bbpix.WithTimeout(30*time.Second),
    bbpix.WithRetry(3, 100*time.Millisecond),
    bbpix.WithCircuitBreaker(10, 5*time.Second),
)
```

## Operações Suportadas

### PIX

#### QR Code

```go
pixClient := client.PIX()

// Criar QR Code
qrCode, err := pixClient.CreateQRCode(ctx, pix.CreateQRCodeRequest{
    TxID:       "txid-123",
    Value:      100.00,
    Expiration: 3600,
})

// Consultar QR Code
qrCode, err := pixClient.GetQRCode(ctx, "txid-123")

// Atualizar QR Code
qrCode, err := pixClient.UpdateQRCode(ctx, "txid-123", pix.UpdateQRCodeRequest{
    Value: 150.00,
})

// Listar QR Codes
list, err := pixClient.ListQRCodes(ctx, pix.ListQRCodesParams{
    StartDate: time.Now().Add(-24 * time.Hour),
    EndDate:   time.Now(),
})
```

#### Pagamentos

```go
// Consultar pagamento
payment, err := pixClient.GetPayment(ctx, "e2e-id")

// Listar pagamentos
payments, err := pixClient.ListPayments(ctx, pix.ListPaymentsParams{
    StartDate: time.Now().Add(-24 * time.Hour),
    EndDate:   time.Now(),
})
```

#### Devoluções

```go
// Criar devolução
refund, err := pixClient.CreateRefund(ctx, "e2e-id", pix.CreateRefundRequest{
    Value: 50.00,
})

// Consultar devolução
refund, err := pixClient.GetRefund(ctx, "e2e-id", "refund-id")
```

### PIX Automático

#### Cobranças Recorrentes

```go
pixAutoClient := client.PIXAuto()

// Criar cobrança recorrente
recurring, err := pixAutoClient.CreateRecurring(ctx, pixauto.CreateRecurringChargeRequest{
    Value:     100.00,
    Frequency: "MONTHLY",
})

// Consultar
recurring, err := pixAutoClient.GetRecurring(ctx, "recurring-id")

// Atualizar
recurring, err := pixAutoClient.UpdateRecurring(ctx, "recurring-id", pixauto.UpdateRecurringChargeRequest{
    Value: 120.00,
})

// Cancelar
err := pixAutoClient.CancelRecurring(ctx, "recurring-id")
```

#### Cobranças Agendadas

```go
// Criar cobrança agendada (até 90 dias)
scheduled, err := pixAutoClient.CreateScheduled(ctx, pixauto.ScheduledChargeRequest{
    Value:         100.00,
    ScheduledDate: time.Now().Add(30 * 24 * time.Hour),
})

// Consultar
scheduled, err := pixAutoClient.GetScheduled(ctx, "scheduled-id")

// Cancelar
err := pixAutoClient.CancelScheduled(ctx, "scheduled-id")
```

#### Acordos de Débito

```go
// Criar acordo
agreement, err := pixAutoClient.CreateAgreement(ctx, pixauto.CreateAgreementRequest{
    PayerKey: "chave-pix-pagador",
    MaxValue: 500.00,
})

// Consultar
agreement, err := pixAutoClient.GetAgreement(ctx, "agreement-id")

// Atualizar
agreement, err := pixAutoClient.UpdateAgreement(ctx, "agreement-id", pixauto.UpdateAgreementRequest{
    MaxValue: 1000.00,
})

// Cancelar
err := pixAutoClient.CancelAgreement(ctx, "agreement-id")
```

### Webhooks

```go
import "github.com/pericles-luz/go-bb-pix/webhook"

// Criar handler
handler := webhook.NewHandler(
    webhook.WithSecret("seu-webhook-secret"),
)

// Registrar handler para evento
handler.On(webhook.EventTypePaymentReceived, func(event webhook.Event) error {
    log.Printf("Pagamento recebido: %s", event.TxID)
    return nil
})

// Usar como HTTP handler
http.Handle("/webhook", handler)
http.ListenAndServe(":8080", nil)
```

## Ambientes

O pacote suporta três ambientes:

```go
bbpix.EnvironmentSandbox      // Ambiente de testes
bbpix.EnvironmentHomologacao  // Ambiente de homologação
bbpix.EnvironmentProducao     // Ambiente de produção
```

## Tratamento de Erros

```go
qrCode, err := pixClient.CreateQRCode(ctx, request)
if err != nil {
    // Verificar se é erro da API
    var apiErr *bbpix.APIError
    if errors.As(err, &apiErr) {
        log.Printf("API Error: code=%d, message=%s",
            apiErr.StatusCode, apiErr.Message)
        for _, detail := range apiErr.Details {
            log.Printf("  - %s: %s", detail.Field, detail.Message)
        }
        return
    }

    // Outro tipo de erro
    log.Printf("Error: %v", err)
}
```

## Testes

### Testes Unitários

```bash
go test ./... -short -cover
```

### Testes de Integração

Os testes de integração requerem credenciais do Banco do Brasil:

```bash
export BB_ENVIRONMENT=sandbox
export BB_CLIENT_ID=seu-client-id
export BB_CLIENT_SECRET=seu-client-secret
export BB_DEV_APP_KEY=sua-app-key

go test -v -tags=integration ./...
```

## Exemplos

Veja a pasta `examples/` para exemplos completos:

- [PIX QR Code](examples/pix_qrcode/main.go)
- [PIX Automático Recorrente](examples/pixauto_recurring/main.go)
- [Webhook Server](examples/webhook_server/main.go)

## Resiliência

O cliente implementa várias estratégias de resiliência:

### Retry Automático

- Retry em erros transitórios (429, 502, 503, 504)
- Exponential backoff com jitter
- Apenas para métodos idempotentes (GET, PUT, DELETE)
- Configurável via `WithRetry()`

### Circuit Breaker

- Proteção contra cascata de falhas
- Fail-fast quando API está indisponível
- Estados: Closed → Open → Half-Open
- Configurável via `WithCircuitBreaker()`

### Timeout

- Timeout global configurável
- Context-aware para timeout por operação
- Configurável via `WithTimeout()`

## Logging

O pacote usa `log/slog` para logging estruturado:

```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

client, err := bbpix.New(config, bbpix.WithLogger(logger))
```

## Contribuindo

Contribuições são bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch para sua feature
3. Escreva testes primeiro (TDD)
4. Implemente a feature
5. Execute os testes e pre-commit checks
6. Abra um Pull Request

## Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.

## Recursos

- [Documentação oficial da API PIX BB](https://www.bb.com.br/site/developers/bb-como-servico/api-pix/)
- [Especificação PIX do Banco Central](https://github.com/bacen/pix-api)
- [Documentação completa (CLAUDE.md)](CLAUDE.md)

## Suporte

Para bugs e feature requests, [abra uma issue](https://github.com/pericles-luz/go-bb-pix/issues).
