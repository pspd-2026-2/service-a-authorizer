# Módulo A: Serviço de Autorização (Authorizer)

Este microsserviço é o **Módulo A** da arquitetura distribuída, atuando como o autorizador principal de transações de cartão de crédito. Desenvolvido em **Go**, ele expõe simultaneamente endpoints **gRPC** e **REST HTTP/1.1** para permitir testes comparativos de desempenho e integração flexível.

## Escopo e Regras de Negócio

O serviço recebe tentativas de transação e avalia a elegibilidade do cartão com base em regras estritas:
* **Validação de Status:** Verifica se o cartão consultado no repositório em memória está `active`, `blocked` ou `expired`.
* **Checagem de Limite:** Garante que o limite disponível (`availableLimit`) seja maior ou igual ao montante solicitado (`amount`).
* **Geração de Autorização:** Transações aprovadas recebem um `TransactionID` único gerado via bytes aleatórios combinados ao timestamp, além de um `AuthCode` em formato hexadecimal.
* **Simulação de Latência (Benchmark):** Cada requisição possui um atraso artificial fixo de `80ms` (`time.Sleep`) na camada de domínio para simular chamadas externas a emissores reais ou I/O de rede, auxiliando nos testes de estresse (gRPC vs REST).

---

## Estrutura do Projeto

A organização das pastas reflete a arquitetura atual do microsserviço:

```text
.
|-- bin/                      # Diretório de saída para binários compilados
|-- cmd/                      
|   `-- server/               
|       `-- main.go           # Ponto de entrada. Injeta dependências e sobe os servidores HTTP e gRPC
|-- config/                   
|   `-- config.go             # Leitura de variáveis de ambiente e configurações
|-- internal/                 # Código privado do microsserviço
|   |-- authorizer/           # Regras de negócio, handlers e persistência (repositório em memória)
|   |   |-- authorizer.go     # Core da lógica de autorização e aprovação de limites
|   |   |-- grpc_handler.go   # Implementação da interface gRPC
|   |   |-- handler.go        # Implementação dos endpoints REST e Health Check
|   |   `-- repository.go     # Acesso e busca de dados dos cartões mockados
|   |-- models/               
|   |   `-- models.go         # Entidades puras do negócio (Request, Response, Output)
|   `-- pb/                   # Código Go auto-gerado pelo compilador do Protocol Buffers
|       |-- models.go         
|       `-- service.go        
|-- proto/                    
|   `-- authorization.proto   # Definição dos contratos de comunicação gRPC
|-- .env.example              # Template de variáveis de ambiente
|-- Dockerfile                # Configuração para gerar a imagem do contêiner
|-- go.mod                    # Gerenciamento de dependências
`-- go.sum                    # Checksums das dependências
```

### Decisões Arquiteturais Relevantes

1. **Banco de Dados em Memória Persistente:** Utilizamos sync.RWMutex para suportar milhares de requisições concorrentes.
2. **Key Stretching (Gargalo de CPU):** Para o benchmark gRPC vs REST, a aplicação implementa um loop de 50.000 iterações de sha256 para gerar a Assinatura de Segurança da transação. Isso simula o processamento pesado e evidencia o poder do HTTP/2 no gRPC em ambientes de alto estresse.



---

## Variáveis de Ambiente

O projeto utiliza o arquivo `.env` para configuração local (veja o `.env.example`). No ambiente Kubernetes, essas variáveis devem ser injetadas nos *Pods*.

| Variável | Valor Padrão | Descrição |
| --- | --- | --- |
| `REST_PORT` | `8081` | Porta em que o servidor REST/JSON irá escutar. |
| `GRPC_PORT` | `50052` | Porta em que o servidor gRPC (HTTP/2) irá escutar. |
| `SERVICE_NAME` |  | Nome do serviço utilizando logs de inicialização |

---

## Como Executar

### 1. Rodando Localmente (Para Desenvolvimento)

Certifique-se de ter o Go 1.22+ instalado.

```bash
# Baixe as dependências
go mod tidy

# Execute o servidor a partir da raiz do projeto
go run ./cmd/server/main.go

```

*Os servidores subirão nas portas `8081` (REST) e `50052` (gRPC) e carregarão a seed de dados do JSON.*

### 2. Rodando via Docker (Para Kubernetes / Minikube)

O Dockerfile está posicionado na raiz do repositório.

```bash
# Gerar a imagem estática
docker build -t authorizer-service:latest .

# Rodar o contêiner mapeando ambas as portas
docker run -p 8081:8081 -p 50052:50052 authorizer-service:latest

```

---

## Como Testar e Gerar a Massa de Dados

Você pode testar a aplicação utilizando o **Postman** (que tem suporte nativo a gRPC).

### API REST (Porta 8081)

**POST** `http://localhost:8081/health`

```json
{
  "status": "ok",
  "service": "authorization-service"
}

```

**POST** `http://localhost:8081/authorize`

```json
{
  "userId": "user-888",
  "carNumber": "1234-5678-9012-3456",
  "amount": 250.00,
  "ipAddress": "192.168.0.10"
}

```

### API gRPC (Porta 50052)

Importe o arquivo proto/authorization.proto no seu cliente gRPC (Postman, grpcurl, etc). O serviço expõe o método Unary Authorize.

```json
{
  "UserId": "user-999", 
  "Amount": 1500.00,
  "CardNumber": "4321-8765-2109-6543",
  "IpAddress": "172.16.254.1"
}

```
