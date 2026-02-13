# Test Fixtures

Este diretório contém fixtures JSON para testes da biblioteca go-bb-pix, baseadas nas especificações oficiais da API PIX do Banco do Brasil.

## Estrutura

```
testdata/
├── cob/                    # Cobranças imediatas
│   ├── create_request.json
│   ├── create_response.json
│   ├── patch_request.json
│   └── list_response.json
├── cobv/                   # Cobranças com vencimento
│   ├── create_request.json
│   └── create_response.json
├── pix/                    # Pagamentos recebidos
│   ├── list_response.json
│   ├── get_response.json
│   ├── refund_request.json
│   └── refund_response.json
├── webhook/                # Webhooks
│   ├── config_request.json
│   ├── config_response.json
│   └── callback_payload.json
└── errors/                 # Respostas de erro
    ├── 400_bad_request.json
    ├── 403_forbidden.json
    ├── 404_not_found.json
    └── 422_unprocessable.json
```

## Fontes das Especificações

As fixtures foram criadas com base em:

1. **OpenAPI Specification**:
   - URL: https://publicador.developers.bb.com.br/bucket/Open_API_Pix_2_8_6_c42e4af8f7_a2ca111fb2.json
   - Versão: 2.8.6

2. **Insomnia Collection**:
   - URL: https://raw.githubusercontent.com/portaldevelopers/PIX/main/PIX.json
   - Portal Developers BB

3. **Especificação BACEN**:
   - Padrão PIX do Banco Central do Brasil
   - Versão: 2.8.2

## Uso nos Testes

### Testes Unitários

Os testes unitários usam estas fixtures para validar:

- Parsing de respostas JSON
- Validação de schema
- Formatação de requests
- Handling de erros

Exemplo:
```go
func TestQRCodeResponse(t *testing.T) {
    data, _ := os.ReadFile("testdata/cob/create_response.json")

    var response QRCodeResponse
    if err := json.Unmarshal(data, &response); err != nil {
        t.Fatalf("Failed to unmarshal: %v", err)
    }

    // Validações...
}
```

### Testes de Integração

Os testes de integração simulam respostas da API usando estas fixtures em servidores HTTP de teste:

```go
func TestIntegration(t *testing.T) {
    responseData, _ := os.ReadFile("testdata/cob/create_response.json")

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write(responseData)
    }))
    defer server.Close()

    client := NewClient(&http.Client{}, server.URL)
    // Testes...
}
```

## Validações Implementadas

### Formato de Dados

- **TxID**: 26-35 caracteres alfanuméricos
- **EndToEndID**: Formato E[0-9a-zA-Z]+ com mínimo 32 caracteres
- **Valor**: Padrão `^\d{1,10}\.\d{2}$`
- **CPF**: 11 dígitos
- **CNPJ**: 14 dígitos
- **Data/Hora**: RFC 3339
- **Chave PIX**: Máximo 77 caracteres
- **Webhook URL**: Máximo 140 caracteres, HTTPS obrigatório

### Status Válidos

**Cobranças (Cob/CobV)**:
- `ATIVA`
- `CONCLUIDA`
- `REMOVIDA_PELO_USUARIO_RECEBEDOR`
- `REMOVIDA_PELO_PSP`

**Devoluções**:
- `DEVOLVIDO`
- `EM_PROCESSAMENTO`
- `NAO_REALIZADO`

**Cobranças Recorrentes**:
- `CRIADA`
- `ATIVA`
- `CONCLUIDA`
- `EXPIRADA`
- `REJEITADA`
- `CANCELADA`

### Códigos de Erro HTTP

- **400**: Bad Request - Validação de schema
- **403**: Forbidden - Acesso negado
- **404**: Not Found - Recurso não encontrado
- **422**: Unprocessable Entity - Regra de negócio
- **429**: Too Many Requests - Rate limit
- **500**: Internal Server Error
- **502**: Bad Gateway
- **503**: Service Unavailable
- **504**: Gateway Timeout

## Manutenção

Ao atualizar as fixtures:

1. Verifique a versão mais recente da especificação OpenAPI
2. Compare com a documentação oficial do BB
3. Execute todos os testes: `go test ./...`
4. Valide schema com: `go test -v -run TestValidation`
5. Atualize este README se houver mudanças estruturais

## Referências

- [API PIX - Banco do Brasil](https://www.bb.com.br/site/developers/bb-como-servico/api-pix/)
- [Portal Developers BB](https://apoio.developers.bb.com.br)
- [Especificação PIX - BACEN](https://github.com/bacen/pix-api)
- [OpenAPI Specification](https://spec.openapis.org/oas/latest.html)
