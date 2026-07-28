# API Rotas de Entrega

Backend em Go (Fiber + GORM + PostgreSQL/PostGIS) para gestão de ruas e rotas de entrega.

## Estrutura

```
backend/
├── main.go
├── config/         # Configurações e variáveis de ambiente
├── database/       # Conexão e migrations
├── docs/           # Documentação Swagger
├── handlers/       # Controllers HTTP
├── middlewares/    # Middlewares compartilhados
├── models/         # Entidades do domínio
├── repositories/   # Acesso a dados (Repository Pattern)
├── routes/         # Definição de rotas
└── services/       # Regras de negócio
```

## Rodando local (Docker)

```bash
docker-compose up --build
```

A API estará disponível em `http://localhost:8080`.

## Documentação Swagger

Acesse: `http://localhost:8080/swagger/index.html`

Para regenerar a documentação (requer swaggo instalado):

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init
```

## Endpoints

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | /ruas | Listar ruas (filtros: ?nome=&cep=&distrito=) |
| GET | /ruas/:id | Buscar rua por ID |
| POST | /ruas | Cadastrar rua |
| PUT | /ruas/:id | Atualizar rua |
| DELETE | /ruas/:id | Remover rua |
| GET | /health | Health check |

## PostGIS Futuro

A imagem Docker já utiliza `postgis/postgis`. Para adicionar campos geométricos no futuro:

1. Ative a extensão via migration:
   ```sql
   CREATE EXTENSION IF NOT EXISTS postgis;
   ```
2. Adicione colunas ao model `Rua`:
   ```go
   Coordenadas geom.Point `gorm:"type:geometry(Point,4326)"`
   Area    geom.Polygon `gorm:"type:geometry(Polygon,4326)"`
   ```
   (usando bibliotecas como `github.com/twpayne/go-geom` ou `gorm.io/datatypes`)

## Variáveis de Ambiente

Copie `.env.example` para `.env` e ajuste conforme necessário.
