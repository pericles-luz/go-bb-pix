# Guia de Contribuição

Obrigado por considerar contribuir com o go-bb-pix! Este documento fornece diretrizes para ajudar você a contribuir com o projeto.

## Índice

- [Código de Conduta](#código-de-conduta)
- [Como Posso Contribuir?](#como-posso-contribuir)
  - [Reportando Bugs](#reportando-bugs)
  - [Sugerindo Melhorias](#sugerindo-melhorias)
  - [Contribuindo com Código](#contribuindo-com-código)
- [Processo de Pull Request](#processo-de-pull-request)
- [Padrões de Código](#padrões-de-código)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Executando Testes](#executando-testes)

## Código de Conduta

Este projeto e todos os participantes estão comprometidos em manter um ambiente respeitoso e acolhedor. Esperamos que todos:

- Usem linguagem acolhedora e inclusiva
- Respeitem pontos de vista e experiências diferentes
- Aceitem críticas construtivas de forma profissional
- Foquem no que é melhor para a comunidade

## Como Posso Contribuir?

### Reportando Bugs

Bugs são rastreados como [GitHub Issues](https://github.com/pericles-luz/go-bb-pix/issues). Antes de criar um bug report:

1. **Verifique se o bug já foi reportado** - procure nas issues existentes
2. **Verifique se você está usando a versão mais recente** - o bug pode já ter sido corrigido

Ao criar um bug report, inclua:

- **Título claro e descritivo**
- **Descrição detalhada** do problema
- **Passos para reproduzir** o comportamento
- **Comportamento esperado** vs **comportamento atual**
- **Versão do Go** (`go version`)
- **Sistema operacional** e versão
- **Logs ou mensagens de erro** relevantes
- **Código de exemplo** que reproduz o problema (se possível)

**Exemplo de bug report:**

```markdown
## Descrição
Ao tentar criar um QR Code dinâmico no ambiente sandbox, recebo erro 401.

## Passos para Reproduzir
1. Configurar cliente com credenciais válidas do sandbox
2. Chamar `pixClient.CreateQRCode(ctx, request)` com valor > 0
3. Observar erro de autenticação

## Comportamento Esperado
QR Code deveria ser criado com sucesso

## Comportamento Atual
Erro: `401 Unauthorized`

## Ambiente
- Go version: 1.21.5
- OS: Ubuntu 22.04
- go-bb-pix version: v0.1.0

## Código de Exemplo
[código aqui]
```

### Sugerindo Melhorias

Sugestões de melhorias também são rastreadas como GitHub Issues. Ao sugerir uma melhoria:

- **Use um título claro e descritivo**
- **Descreva o comportamento atual** e **por que ele é inadequado**
- **Descreva o comportamento proposto** e por que seria útil
- **Liste casos de uso** que se beneficiariam da melhoria
- **Inclua exemplos de código** mostrando como a API seria usada

**Exemplo de sugestão:**

```markdown
## Proposta
Adicionar suporte para timeout customizado por requisição

## Motivação
Algumas operações (como consultas) podem necessitar timeouts menores,
enquanto operações de criação podem precisar de mais tempo.

## Solução Proposta
```go
qrCode, err := pixClient.CreateQRCode(ctx, request,
    pix.WithTimeout(60*time.Second),
)
```

## Alternativas Consideradas
- Usar contexto com deadline (atual)
- Configuração global por tipo de operação
```

### Contribuindo com Código

Contribuições de código são muito bem-vindas! Siga estes passos:

1. **Fork o repositório**
2. **Clone seu fork**
3. **Crie uma branch** para sua feature/fix
4. **Faça suas alterações** seguindo os padrões de código
5. **Escreva/atualize testes**
6. **Execute os testes** e garanta que passam
7. **Commit suas mudanças**
8. **Push para seu fork**
9. **Abra um Pull Request**

## Processo de Pull Request

### Antes de Submeter

1. **Verifique se existe uma issue** relacionada à sua mudança
   - Se não existir, considere criar uma para discussão
   - Para mudanças pequenas (typos, docs), pode ir direto ao PR

2. **Garanta que seu código segue os padrões** do projeto

3. **Execute todos os testes**:
   ```bash
   # Testes unitários
   go test ./... -short -cover

   # Testes de integração (se aplicável)
   go test -v -tags=integration ./...
   ```

4. **Execute as verificações de qualidade**:
   ```bash
   ./scripts/pre-commit.sh
   ```

5. **Atualize a documentação** se necessário

### Submetendo o Pull Request

1. **Título descritivo** usando um dos prefixos:
   - `feat:` - Nova funcionalidade
   - `fix:` - Correção de bug
   - `docs:` - Apenas documentação
   - `refactor:` - Refatoração sem mudança de comportamento
   - `test:` - Adiciona ou corrige testes
   - `chore:` - Mudanças em build, CI, etc

2. **Descrição completa** incluindo:
   - O que foi mudado e por quê
   - Link para issue relacionada (se houver)
   - Resultados de testes
   - Screenshots/exemplos (se aplicável)

3. **Commits organizados**:
   - Commits atômicos e com mensagens claras
   - Cada commit deve compilar e passar nos testes

**Exemplo de PR:**

```markdown
## Descrição
Adiciona suporte para QR Code dinâmico com vencimento customizado.

Fixes #42

## Mudanças
- Adiciona campo `ExpirationDate` em `CreateQRCodeRequest`
- Valida que data de vencimento está no futuro
- Atualiza documentação e exemplos

## Testes
- ✅ Testes unitários: `go test ./pix -v`
- ✅ Testes de integração: Testado no sandbox
- ✅ Cobertura: 95% (mantida)

## Checklist
- [x] Código segue os padrões do projeto
- [x] Testes adicionados/atualizados
- [x] Documentação atualizada
- [x] Todos os testes passando
- [x] Pre-commit checks executados
```

### Revisão de Código

- Mantenha a discussão focada no código, não na pessoa
- Seja receptivo a feedback
- Responda a todos os comentários
- Faça as alterações solicitadas ou explique por que não devem ser feitas

## Padrões de Código

### 1. Test-Driven Development (TDD)

Sempre escreva testes antes do código:

```go
// 1. Escreva o teste primeiro
func TestCreateQRCode_Success(t *testing.T) {
    // arrange, act, assert
}

// 2. Implemente o código
func (c *Client) CreateQRCode(ctx context.Context, req CreateQRCodeRequest) (*QRCodeResponse, error) {
    // implementação
}
```

### 2. Context-First

Todos os métodos públicos devem aceitar `context.Context`:

```go
func (c *Client) CreateQRCode(ctx context.Context, req CreateQRCodeRequest) (*QRCodeResponse, error)
```

### 3. Error Wrapping

Sempre adicione contexto aos erros:

```go
if err != nil {
    return nil, fmt.Errorf("failed to create qr code for txid %s: %w", req.TxID, err)
}
```

### 4. Structured Logging

Use `slog` para logging estruturado:

```go
logger.InfoContext(ctx, "creating qr code",
    slog.String("txid", req.TxID),
    slog.Float64("value", req.Value),
)
```

### 5. Table-Driven Tests

Prefira table-driven tests para múltiplos cenários:

```go
tests := []struct {
    name    string
    input   CreateQRCodeRequest
    want    *QRCodeResponse
    wantErr bool
}{
    {
        name: "success",
        input: CreateQRCodeRequest{TxID: "123", Value: 10.0},
        want: &QRCodeResponse{QRCode: "..."},
        wantErr: false,
    },
    {
        name: "invalid txid",
        input: CreateQRCodeRequest{TxID: "", Value: 10.0},
        want: nil,
        wantErr: true,
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := client.CreateQRCode(ctx, tt.input)
        if (err != nil) != tt.wantErr {
            t.Errorf("CreateQRCode() error = %v, wantErr %v", err, tt.wantErr)
            return
        }
        // assert got == tt.want
    })
}
```

### 6. Nomenclatura

- **Pacotes**: lowercase, uma palavra (evite underscores)
- **Interfaces**: nomes terminam em `-er` quando possível (`TokenProvider`)
- **Constantes**: CamelCase ou SCREAMING_SNAKE_CASE para exported
- **Variáveis**: camelCase para privadas, CamelCase para públicas

### 7. Documentação

- Todo símbolo público deve ter um comentário
- Comentários devem começar com o nome do símbolo
- Use exemplos executáveis quando apropriado

```go
// CreateQRCode cria um novo QR Code PIX no Banco do Brasil.
// O contexto pode ser usado para cancelamento e timeout.
//
// Exemplo:
//
//	qrCode, err := client.CreateQRCode(ctx, CreateQRCodeRequest{
//	    TxID:  "123",
//	    Value: 10.50,
//	})
func (c *Client) CreateQRCode(ctx context.Context, req CreateQRCodeRequest) (*QRCodeResponse, error)
```

## Estrutura do Projeto

```
go-bb-pix/
├── bbpix/          # Cliente principal e configuração
├── pix/            # Operações PIX (QR Code, pagamentos, devoluções)
├── pixauto/        # PIX Automático (recorrência, agendamento)
├── internal/       # Pacotes internos (não exportados)
│   ├── auth/       # OAuth2
│   ├── http/       # Cliente HTTP base
│   └── transport/  # Middlewares (retry, circuit breaker, logging)
├── examples/       # Exemplos de uso
└── testdata/       # Fixtures para testes
```

### Onde Adicionar Seu Código

- **Nova operação PIX**: adicione em `pix/`
- **Nova operação PIX Automático**: adicione em `pixauto/`
- **Mudança no cliente base**: modifique `bbpix/`
- **Novo middleware de transporte**: adicione em `internal/transport/`
- **Utilitários internos**: adicione em `internal/`

## Executando Testes

### Testes Unitários (rápidos)

```bash
go test ./... -short -cover
```

### Testes de Integração (requer credenciais)

```bash
# Configure variáveis de ambiente
export BB_DEV_APP_KEY=sua_chave
export BB_CLIENT_ID=seu_client_id
export BB_CLIENT_SECRET=seu_secret
export BB_ENVIRONMENT=sandbox

# Execute testes de integração
go test -v -tags=integration ./...
```

### Verificações de Qualidade

```bash
# Execute todas as verificações
./scripts/pre-commit.sh

# Ou manualmente:
go test ./... -short          # Testes
go mod tidy                   # Organizar dependências
go vet ./...                  # Análise estática
staticcheck ./...             # Linter avançado
go build ./...                # Verificar compilação
```

### Cobertura de Testes

```bash
# Gerar relatório de cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Mantenha a cobertura acima de 80%.

## Perguntas?

Se tiver dúvidas que não foram respondidas neste guia:

1. Procure nas [issues existentes](https://github.com/pericles-luz/go-bb-pix/issues)
2. Abra uma nova issue com a tag `question`
3. Entre em contato com os mantenedores

## Licença

Ao contribuir, você concorda que suas contribuições serão licenciadas sob a mesma licença do projeto (MIT License).

---

**Obrigado por contribuir! 🚀**
