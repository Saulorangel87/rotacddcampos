# Nota de status — Site Correios (CDD Campos dos Goytacazes)

_Gerado em 29/07/2026_

## Visão geral do projeto

Ferramenta interna pra substituir mapas de papel na sua unidade: consulta e gestão das ruas atendidas pelos 24 distritos postais (601–624) do CDD Campos dos Goytacazes.

- **Frontend**: React + Vite (migrado do HTML/CSS/JS puro original)
- **Backend**: Go (Fiber + GORM), PostgreSQL 17 (`rotas_db`)
- **Banco**: 2264 ruas importadas, todas ativas

## O que já está funcionando

### Backend (Go)

- API sobe local com `go run main.go`, porta 8080
- `GET /ruas` com filtros `?distrito=` e `?nome=` — testado e validado contra os dados reais
- `PUT /ruas/:id` — usado pra gravar mudança de distrito
- CORS configurado liberando `http://localhost:5173` (necessário pro front conversar com a API)
- API `GET /colaboradores` agora aceita `?carteiro=true` no backend para filtrar apenas colaboradores cujo cargo/funcao indica carteiro (motorizado/ciclista)
- API `GET /colaboradores/aniversariantes-hoje` passou a aceitar data simulada no formato `DD/MM` ou `DD-MM`
- Datas no JSON de colaboradores agora são serializadas em formato brasileiro `dd/mm/yyyy`
- Tabela `ruas` já tem colunas preparadas pro futuro: `ativo`, `latitude`, `longitude`, `rota_id`, `rota` — **mas o model Go (`models/rua.go`) ainda não expõe `ativo`/`latitude`/`longitude`/`rota_id`**, só usa os 9 campos originais

### Frontend (React)

- Layout seguindo o mockup: header com logo real dos Correios, nav de 24 distritos (cores únicas, sem repetir), sidebar, mapa esquemático em SVG
- Painel **Ajustes de Rotas**: assistente de 3 passos (selecionar ruas → escolher distrito → confirmar), com paginação (12 por página) e altura travada na tela (não empurra mais o botão "Próximo" pra fora)
- Tabela de ruas: abas (todas/distrito/carteiro/CEP), paginação (30 por página), exportar CSV
- Busca corrigida: distrito por igualdade exata, CEP por prefixo, rua/carteiro por substring (antes misturava tudo e trazia resultado errado tipo buscar "620" e vir CEPs de outros distritos)
- Bug de zoom do mapa central corrigido (tinha teto de tamanho errado, ajustado)

## Pendências conhecidas (nada urgente, só registro)

1. **Model Go desatualizado** — adicionar `ativo`, `latitude`, `longitude`, `rota_id` em `models/rua.go` pra API enxergar essas colunas que já existem no banco
2. **Latitude/longitude vazias** — 0 de 2264 ruas geocodificadas; precisa de rotina de geocodificação em lote (CEP → coordenada) quando for trabalhar nisso
3. **Endpoint de mover em lote + histórico** — hoje o "Ajustes de Rotas" faz um `PUT` por rua; falta `POST /ruas/mover-lote` + tabela `historico_alteracoes` no backend
4. **Ajuste visual pequeno** — texto "Correios" um pouco desalinhado do ícone no selo do header
5. **Modelo relacional (carteiros/rotas)** — documentação técnica já descreve `carteiros` → `rotas` → `ruas`, mas ainda não implementado (tabela `carteiros`/`rotas` não existem no banco ainda)

## Planejado pra mais à frente (só desenho, não implementar ainda)

### Autenticação simples (admin vs. carteiro)

- Sem login complexo — matrícula + senha, JWT simples
- Padrão: **somente leitura pra todos**
- Só você (Saulo) é admin no início; botão "gerenciar usuários" futuro permite conceder acesso admin a outras pessoas
- Trava sempre no backend (middleware verifica papel antes de qualquer escrita), front só esconde botão por UX

### Redistribuição automática de ruas órfãs

Pra quando o número de distritos cair (hoje 24, pode diminuir):

- Motor de sugestão por heurística (proximidade + carga de trabalho) — **não IA decidindo sozinha**
- IA entra só pra gerar a justificativa em texto de cada sugestão, e pra interpretar nomes de rua bagunçados
- Sugestões caem pré-preenchidas dentro do próprio Ajustes de Rotas, humano sempre confirma

## Próxima ação combinada

Amanhã: você traz a planilha dos carteiros (nome + matrícula) → eu preparo a importação pro banco + ajusto o model Go junto.

## Arquivos entregues

- `frontend-react-ajustes-rotas.zip` — projeto React completo (última versão, com todos os fixes desta sessão)
- `.gitignore` já orientado (cobre `node_modules`, `.env` de ambas as pastas, artefatos de build Go)
