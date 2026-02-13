# Guia de Testes

Este documento descreve a estratégia e organização de testes para o projeto go-bb-pix.

## Tipos de Testes

### 1. Testes Unitários

Testes rápidos que não dependem de recursos externos. Executam validação de lógica de negócio, parsing de dados e manipulação de estruturas.

**Localização**: `*_test.go` em cada pacote

**Executar**:
```bash
# Todos os testes unitários
go test ./... -short

# Com cobertura
go test ./... -short -cover

# Verbose
go test ./... -short -v

# Pacote específico
go test ./pix -short -v
```

**Características**:
- Não fazem chamadas HTTP reais
- Usam `httptest.Server` para simular API
- Baseados em fixtures do diretório `testdata/`
- Executam em milissegundos

### 2. Testes de Validação

Validam conformidade com as especificações da API (OpenAPI 2.8.6).

**Arquivos**:
- `pix/validation_test.go`: Validações de schema e formato
- `pix/error_handling_test.go`: Casos de erro

**Executar**:
```bash
go test ./pix -v -run TestValidation
go test ./pix -v -run TestAPIErrorResponses
```

**O que validam**:
- Formato de TxID (26-35 caracteres alfanuméricos)
- Formato de valores monetários (^\d{1,10}\.\d{2}$)
- CPF (11 dígitos) e CNPJ (14 dígitos)
- Datas em RFC 3339
- Códigos de status válidos
- Limites de campo (webhookUrl: 140 chars, chave: 77 chars)

### 3. Testes de Integração

Simulam fluxos completos da API usando mock servers.

**Arquivos**:
- `pix/integration_test.go`: Fluxos de QR Code, pagamentos, devoluções
- `pix/cobv_test.go`: Cobranças com vencimento
- `pix/webhook_test.go`: Webhooks e callbacks

**Executar**:
```bash
go test ./pix -v -run TestIntegration
go test ./pix -v -run TestCreateQRCodeIntegration
```

**Cenários testados**:
- Ciclo completo de QR Code (criar, consultar, atualizar, deletar)
- Listagem com paginação
- Filtros de busca (CPF, CNPJ, status, data)
- Criação e consulta de devoluções
- Configuração de webhooks
- Concorrência (múltiplas requisições simultâneas)
- Cancelamento via context

### 4. Testes End-to-End (E2E)

Testes contra a API real do Banco do Brasil (sandbox/homologação).

**Arquivo**: `pix/e2e_test.go`

**Build Tag**: `integration`

**Pré-requisitos**:

1. Criar arquivo `.env` a partir do exemplo:
   ```bash
   cp .env.example .env
   ```

2. Preencher credenciais no `.env`:
   ```env
   BB_ENVIRONMENT=sandbox
   BB_CLIENT_ID=seu-client-id
   BB_CLIENT_SECRET=seu-client-secret
   BB_DEV_APP_KEY=sua-developer-application-key
   ```

Veja [ENV_CONFIG.md](ENV_CONFIG.md) para mais detalhes sobre configuração.

**Executar**:
```bash
# Executar testes E2E
go test -v -tags=integration ./pix

# Apenas um teste específico
go test -v -tags=integration ./pix -run TestE2E_CompleteQRCodeFlow
```

**⚠️ Atenção**:
- Requer credenciais válidas do BB
- Executa operações reais na API
- Pode ter custos ou limites de taxa
- Não executam com `-short`

## Fixtures de Teste

Diretório `testdata/` contém payloads JSON baseados nas especificações oficiais:

```
testdata/
├── cob/          # Cobranças imediatas
├── cobv/         # Cobranças com vencimento
├── pix/          # Pagamentos e devoluções
├── webhook/      # Configurações e callbacks
└── errors/       # Respostas de erro (400, 403, 404, 422)
```

Ver [testdata/README.md](testdata/README.md) para mais detalhes.

## Cobertura de Testes

**Meta**: Mínimo 80% de cobertura

**Verificar cobertura**:
```bash
# Gerar relatório
go test ./... -short -coverprofile=coverage.out

# Visualizar no navegador
go tool cover -html=coverage.out

# Cobertura por pacote
go test ./... -short -cover | grep coverage
```

**Áreas críticas**:
- [ ] Parsing de requests/responses: 100%
- [ ] Validação de dados: 100%
- [ ] Error handling: 100%
- [ ] Client operations: 90%+
- [ ] Transport layer: 80%+

## Padrões de Teste

### Table-Driven Tests

Preferir table-driven tests para múltiplos cenários:

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "valid case", input: "value", want: "result", wantErr: false},
        {name: "error case", input: "bad", want: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if got != tt.want {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Mock HTTP Servers

Usar `httptest.Server` para simular respostas da API:

```go
func TestAPICall(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verificar request
        if r.Method != http.MethodPut {
            t.Errorf("Method = %s, want PUT", r.Method)
        }

        // Retornar response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        w.Write(responseData)
    }))
    defer server.Close()

    client := NewClient(&http.Client{}, server.URL)
    // Executar teste...
}
```

### Context Handling

Sempre testar cancelamento de context:

```go
func TestContextCancellation(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
    defer cancel()

    _, err := client.LongOperation(ctx)

    if err == nil {
        t.Fatal("Expected context deadline exceeded")
    }
}
```

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Unit Tests
        run: go test ./... -short -cover

      - name: Validation Tests
        run: go test ./pix -v -run TestValidation

      - name: Integration Tests
        run: go test ./pix -v -run TestIntegration

      - name: Coverage Report
        run: |
          go test ./... -short -coverprofile=coverage.out
          go tool cover -func=coverage.out
```

### Pre-commit Hook

Instalar hook local:

```bash
cp scripts/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

Ou usar manualmente:
```bash
./scripts/pre-commit.sh
```

## Debugging de Testes

### Executar teste específico

```bash
go test ./pix -v -run TestCreateQRCode
```

### Ver output detalhado

```bash
go test ./pix -v -run TestName 2>&1 | less
```

### Usar debugger (delve)

```bash
dlv test ./pix -- -test.run TestName
```

### Logs estruturados

Os testes usam `t.Log()` e `t.Logf()` para output detalhado:

```go
t.Logf("Creating QR Code with txid: %s", txid)
```

## Troubleshooting

### Testes E2E falhando

1. Verificar se `.env` existe:
   ```bash
   ls -la .env
   cat .env
   ```

2. Verificar credenciais no `.env`:
   ```bash
   grep BB_ .env
   ```

3. Recriar `.env` se necessário:
   ```bash
   cp .env.example .env
   # Editar e preencher credenciais
   ```

3. Testar autenticação:
   ```bash
   go test -v -tags=integration ./internal/auth -run TestOAuth2
   ```

### Fixtures desatualizadas

Se a API mudou:

1. Consultar versão atual: https://publicador.developers.bb.com.br/
2. Baixar nova especificação OpenAPI
3. Atualizar fixtures em `testdata/`
4. Executar: `go test ./pix -v -run TestValidation`

### Erros de rate limit

Testes E2E podem exceder rate limits:

1. Reduzir número de testes simultâneos
2. Adicionar delays: `time.Sleep(time.Second)`
3. Usar ambiente de sandbox

## Boas Práticas

✅ **Fazer**:
- Executar testes antes de commit
- Manter cobertura acima de 80%
- Usar table-driven tests
- Validar todos os campos críticos
- Testar casos de erro
- Documentar testes complexos

❌ **Evitar**:
- Testes que dependem de ordem de execução
- Hardcoded credentials em código
- Testes flaky (intermitentes)
- Ignorar erros em setup/teardown
- Compartilhar estado entre testes

## Recursos

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Go HTTP Test](https://pkg.go.dev/net/http/httptest)
- [API PIX - Documentação BB](https://www.bb.com.br/site/developers/)
- [Especificação BACEN](https://github.com/bacen/pix-api)
