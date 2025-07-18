# Desafio Frete Rápido

API para simulação de cotação de frete e métricas desenvolvida em Go usando Fiber framework.

## 📋 Índice

- [Pré-requisitos](#pré-requisitos)
- [Configuração](#configuração)
- [Execução com Docker](#execução-com-docker)
- [Execução Local](#execução-local)
- [Endpoints da API](#endpoints-da-api)
- [Exemplos de Uso](#exemplos-de-uso)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Testes](#testes)
- [Comandos Úteis](#comandos-úteis)

## 🛠 Pré-requisitos

- Docker
- Docker Compose
- Go 1.24.4+ (para desenvolvimento local)

## ⚙️ Configuração

### Variáveis de Ambiente

Configure o arquivo `.env.local` na raiz do projeto com as seguintes variáveis:

```env
# Configurações da Aplicação
APP_PORT=8080
LOGGING_JSON_FORMAT=true
LOGGING_LEVEL=info

# Configurações do Banco de Dados
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=desafio_frete_rapido

# Configurações CORS
HTTP_CORS_ALLOWED_HEADERS=*
HTTP_CORS_ALLOWED_METHODS=GET,POST
HTTP_CORS_ALLOWED_ORIGINS=*

# Configurações da API do Frete Rápido
FASTDELIVERY_API_BASE_URL=https://baseurl.com/api/v3
FASTDELIVERY_API_TOKEN=your_api_token_here
FASTDELIVERY_API_PLATFORM_CODE=your_platform_code_here
FASTDELIVERY_API_SENDER_CNPJ=your_sender_cnpj_here
FASTDELIVERY_API_ZIP_CODE=your_zip_code_here
```

**⚠️ Importante:** Substitua os valores das variáveis da API do Frete Rápido pelos valores corretos fornecidos pela plataforma.

## 🐳 Execução com Docker

### 1. Construir a imagem da aplicação

```bash
docker build . -t frete_rapido_app
```

### 2. Iniciar os serviços

```bash
docker compose up -d
```

Este comando irá:
- Iniciar o container PostgreSQL na porta 5432
- Iniciar a aplicação na porta 8080

### 3. Verificar se os serviços estão rodando

```bash
docker compose ps
```

### 4. Parar os serviços

```bash
docker compose down -v
```

## 💻 Execução Local

### 1. Instalar dependências

```bash
go mod tidy
```

### 2. Configurar banco de dados

Certifique-se de que o PostgreSQL está rodando e configure as variáveis de ambiente no `.env.local` com os valores corretos para sua instância local.

### 3. Executar aplicação

```bash
make run
# ou
go run ./cmd/main.go
```

## 🔗 Endpoints da API

### Base URL
```
http://localhost:8080/v1
```

### 1. Simulação de Cotação de Frete

**POST** `/v1/quote`

Simula cotação de frete com base nos dados fornecidos.

#### Payload de Exemplo:

```json
{
  "recipient": {
    "address": {
      "zipcode": "01310100"
    }
  },
  "volumes": [
    {
      "category": 7,
      "amount": 1,
      "unitary_weight": 5.0,
      "price": 349.00,
      "sku": "ABC123",
      "height": 2.0,
      "width": 11.0,
      "length": 16.0
    }
  ]
}
```

#### Resposta de Exemplo:

```json
{
  "carriers": [
    {
      "name": "CORREIOS",
      "service": "PAC",
      "deadline": 5,
      "price": 15.50
    },
    {
      "name": "CORREIOS", 
      "service": "SEDEX",
      "deadline": 2,
      "price": 25.80
    }
  ]
}
```

### 2. Métricas de Cotações

**GET** `/v1/metrics?last_quotes=10`

Retorna métricas das últimas cotações realizadas.

#### Parâmetros de Query:
- `last_quotes` (opcional): Número de cotações a considerar (padrão: 10)

#### Resposta de Exemplo:

```json
{
  "carrier_quotes": [
    {
      "carrier_name": "CORREIOS",
      "total_quotes": 15,
      "total_price": 387.50,
      "average_price": 25.83
    }
  ],
  "cheapest_shipping": 15.50,
  "highest_shipping": 45.90
}
```

## 📝 Exemplos de Uso

### Usando curl

#### Simulação de Cotação:

```bash
curl -X POST http://localhost:8080/v1/quote \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": {
      "address": {
        "zipcode": "01310100"
      }
    },
    "volumes": [
      {
        "category": 7,
        "amount": 1,
        "unitary_weight": 5.0,
        "price": 349.00,
        "sku": "ABC123",
        "height": 2.0,
        "width": 11.0,
        "length": 16.0
      }
    ]
  }'
```

#### Consulta de Métricas:

```bash
curl -X GET "http://localhost:8080/v1/metrics?last_quotes=5"
```

### Usando httpie

#### Simulação de Cotação:

```bash
http POST localhost:8080/v1/quote \
  recipient:='{"address":{"zipcode":"01310100"}}' \
  volumes:='[{"category":7,"amount":1,"unitary_weight":5.0,"price":349.00,"sku":"ABC123","height":2.0,"width":11.0,"length":16.0}]'
```

#### Consulta de Métricas:

```bash
http GET localhost:8080/v1/metrics last_quotes==5
```

## 📁 Estrutura do Projeto

```
.
├── cmd/                     # Ponto de entrada da aplicação
│   └── main.go
├── internal/                # Código interno da aplicação
│   └── quote/              # Módulo de cotações
│       ├── controller.go   # Lógica de negócio
│       ├── entity.go       # Estruturas de dados
│       ├── handler.go      # Handlers HTTP
│       └── repository.go   # Acesso a dados
├── pkg/                    # Pacotes reutilizáveis
│   ├── config/            # Configurações
│   ├── database/          # Conexão e queries do banco
│   ├── fastdelivery_api/  # Cliente da API externa
│   ├── logger/            # Sistema de logs
│   └── server/            # Configuração do servidor HTTP
├── docker-compose.yaml    # Configuração dos serviços
├── Dockerfile            # Imagem da aplicação
├── Makefile             # Comandos automatizados
└── README.md           # Documentação
```

## 🧪 Testes

### Executar todos os testes

```bash
make test
# ou
go test ./...
```

### Executar testes com verbose

```bash
go test -v ./...
```

### Executar testes de um pacote específico

```bash
go test ./internal/quote/
```

## 🔧 Comandos Úteis

### Makefile

O projeto inclui um Makefile com comandos úteis:

```bash
# Subir os serviços do Docker Compose
make up

# Parar os serviços do Docker Compose
make down

# Executar testes
make test

# Executar aplicação localmente
make run
```

### Docker

```bash
# Ver logs da aplicação
docker compose logs -f frete_rapido

# Ver logs do banco de dados
docker compose logs -f postgres

# Executar comandos no container da aplicação
docker compose exec frete_rapido sh

# Reconstruir e reiniciar os serviços
docker compose up -d --build
```

### Banco de Dados

```bash
# Conectar ao PostgreSQL
docker compose exec postgres psql -U postgres -d desafio_frete_rapido

# Ver tabelas
docker compose exec postgres psql -U postgres -d desafio_frete_rapido -c "\dt"
```

## 📊 Health Check

Para verificar se a aplicação está funcionando corretamente:

```bash
# Verificar se a aplicação responde
curl http://localhost:8080/v1/metrics

# Status dos containers
docker compose ps

# Logs em tempo real
docker compose logs -f
```

## 🐛 Solução de Problemas

### Erro de conexão com banco de dados
- Verifique se o PostgreSQL está rodando: `docker compose ps`
- Verifique as variáveis de ambiente no `.env.local`
- Aguarde o banco ficar pronto (health check pode levar alguns segundos)

### Erro na API externa
- Verifique as credenciais da API do Frete Rápido no `.env.local`
- Teste a conectividade com a API externa

### Porta já em uso
- Verifique se a porta 8080 não está sendo usada por outro processo
- Altere a variável `APP_PORT` no `.env.local` se necessário

---

**Desenvolvido por:** Jean Carlos  
**Tecnologias:** Go, Fiber, PostgreSQL, Docker