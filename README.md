# Guia de Logística — CDD Campos dos Goytacazes

Ferramenta interna da unidade CDD Campos dos Goytacazes (Correios): mapa
interativo dos 24 distritos postais (601–624), consulta de ruas/CEP,
cadastro de colaboradores e ajuste de rotas (mover rua entre distritos).

Em produção: **https://cddcampos.devsaulo.com.br** — versão atual: **v1.2.0**

## Stack

- **Frontend**: React 19 + Vite, mapa real com React Leaflet + OpenStreetMap
- **Backend**: Go (Fiber + GORM), PostgreSQL 17 (PostGIS)
- **Autenticação**: JWT + bcrypt, 3 níveis de acesso (público / colaborador / admin)
- **PWA**: instalável no celular, com Service Worker (só cacheia estático, nunca dado de API)
- **Deploy**: Docker Compose (postgres + api + frontend), atrás de Cloudflare Tunnel

## Estrutura

```
Backend/
  cmd/              # comandos standalone: seed-admin, criar-usuario
  config/           # variáveis de ambiente
  database/         # conexão + migrations (GORM AutoMigrate)
  handlers/         # camada HTTP (Fiber)
  services/         # regras de negócio
  repositories/      # acesso ao banco
  models/           # structs GORM
  middlewares/      # auth (JWT), error handler
  routes/           # todas as rotas e o que é público/autenticado/admin

Frontend/
  src/
    api/            # chamadas para o backend (client.js centraliza token + erro)
    context/        # AuthContext (sessão, login/logout)
    components/     # UI (Header, Sidebar, MapPanel/LeafletMap, RuasTable, modais...)
  public/           # favicon, ícones PWA, manifest.json, sw.js
```

## Modelo de acesso

| Nível | O que faz | Precisa de login? |
|---|---|---|
| Público | mapa, busca de rua/CEP, imprimir listagem, aniversariante do dia | Não |
| Colaborador | + ver colaboradores, relatórios/histórico | Sim (matrícula + senha) |
| Admin | + criar/editar/excluir ruas e colaboradores, gerenciar usuários | Sim, papel admin |

Login é matrícula + senha (bcrypt), JWT de 8h. Conta trava por 15min após 5
tentativas erradas (por conta) + limite de 10 tentativas/minuto por IP.
Senha sempre nasce provisória — troca obrigatória no primeiro login.

## Rodando localmente

**Backend:**
```bash
cd Backend
cp .env.example .env    # ajuste DB_PASSWORD, gere um JWT_SECRET novo
go mod tidy
go run main.go
```
Sobe em `localhost:8080`. As tabelas são criadas sozinhas (AutoMigrate).

Criar o primeiro admin:
```bash
go run cmd/seed-admin/main.go -matricula SUA_MATRICULA -senha "SenhaTemporaria123"
```

**Frontend:**
```bash
cd Frontend
cp .env.example .env    # VITE_API_URL=http://localhost:8080
npm install
npm run dev
```
Abre em `localhost:5173`.

## Deploy em produção

Ver `nota-de-status-site-correios.md` para o passo a passo completo, status atual e pendências.

## Funcionalidades

- Mapa real (Leaflet + OpenStreetMap) com os 24 distritos, camada opcional
  de traçado real das ruas
- Busca de rua por nome/CEP/distrito, exportar CSV, imprimir
- Ajustes de Rotas: mover rua entre distritos (admin), com histórico
  persistido no banco (quem moveu, quando, de onde pra onde)
- Colaboradores: cadastro, busca, exclusão (admin), aniversariante do dia
  em destaque (público, sem login)
- Consulta de Folgas: saldo por matrícula (livro-razão de créditos/débitos),
  consulta pública sem listar ninguém, lançar/excluir restrito a admin,
  com auditoria de quem lançou
- Observações de rua: conhecimento de campo dos carteiros (acesso difícil,
  numeração fora de ordem, mais de um nome, segurança), cadastro restrito a
  admin, leitura pública
- Zé Rota: assistente em chat (texto e voz) — busca rua/distrito/CEP no
  cadastro real (nunca inventa endereço), consulta o tempo em Campos dos
  Goytacazes (Open-Meteo) e sugere um link do Google Maps só quando não
  encontra nada no cadastro interno, deixando sempre claro que não é dado
  oficial
- Gerenciar usuários: criar conta, definir papel, resetar senha — tudo
  pela interface, sem precisar de terminal
- PWA: instalável no celular, funciona offline pro casco estático

## Pendências conhecidas

- Ajuste visual pequeno: texto "Correios" levemente desalinhado do ícone
  no selo do header (baixa prioridade)
- Resetar/bloquear um usuário não invalida um token JWT já emitido (fica
  válido até expirar sozinho em 8h) — aceitável pro tamanho da equipe hoje
- Botão "Por Carteiro" na tabela de ruas não filtra de verdade, só reordena
  (combinado deixar parado, sem prioridade)
