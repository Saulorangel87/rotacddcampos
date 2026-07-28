# CDD Campos — Ajustes de Rotas (Frontend React)

Refatoração do site estático (`Frontend/index.html` etc.) para React + Vite,
seguindo o mockup da interface "Ajustes de Rotas".

## Rodando

```bash
npm install
cp .env.example .env    # ajuste VITE_API_URL se a API Go não estiver em localhost:8080
npm run dev
```

Abre em `http://localhost:5173`.

## Como está organizado

```
src/
  api/ruas.js           # chamadas para a API Go (Backend/) — cai para dados de exemplo se a API não responder
  data/distritos.js      # cores e layout dos 9 distritos (mesmas cores da legenda atual)
  data/mockRuas.js        # dados de exemplo, no formato do model Rua do backend
  components/
    Header.jsx            # topo azul com busca
    DistrictNav.jsx        # faixa amarela com os botões 601-609
    Sidebar.jsx            # menu lateral (Mapa Geral, Ajustes de Rotas, Ruas, CEP, Carteiros, Relatórios)
    DistrictMap.jsx        # mini-mapa esquemático em SVG (não é geográfico, é um diagrama de blocos)
    MapPanel.jsx           # painel principal com o mapa + legenda
    AjustesRotasPanel/     # assistente de 3 passos: selecionar ruas → escolher distrito → confirmar
    RecentChanges.jsx      # log de alterações (hoje só em memória, na sessão)
    RuasTable.jsx          # tabela com abas (todas/distrito/carteiro/cep), busca e exportar CSV
  App.jsx                  # junta tudo
```

## O que já conversa com o backend Go

- `GET /ruas` (com filtro por `distrito`) — usado para listar as ruas no painel de ajustes e na tabela.
- `PUT /ruas/:id` — usado (uma vez por rua) para gravar o novo distrito quando você confirma uma mudança.

## O que ainda é só front (precisa de trabalho no backend depois)

- **Mover em lote com histórico**: hoje `moverRuasEmLote` faz um `PUT` por rua, um de cada vez. O ideal
  é criar `POST /ruas/mover-lote` no Go que já grave um registro de histórico (rua, distrito de origem,
  distrito de destino, carteiro, motivo, usuário, data) — daí o "Alterações recentes" deixa de ser só
  da sessão e passa a vir do banco.
- **Carteiro responsável**: o model `Rua` no backend não tem esse campo ainda; o mock usa `rota` para
  simular. Quando adicionar a coluna, é só ajustar `api/ruas.js`.
- **Login/usuário**: o nome "Saulo" no cabeçalho e no histórico está fixo — sem autenticação ainda.
- **Mapa geográfico real**: o `DistrictMap` é um diagrama esquemático (SVG), não usa Google Maps/PostGIS.
  Serve para orientar visualmente qual distrito está em foco; se quiser o mapa real depois, dá pra trocar
  esse componente por um mapa com Leaflet ou Google Maps JS API usando as coordenadas do PostGIS.

## Próximos passos sugeridos

1. Terminar a importação das ~2 mil ruas no PostgreSQL.
2. Adicionar `POST /ruas/mover-lote` + tabela `historico_alteracoes` no backend Go.
3. Trocar o `DistrictMap` esquemático pelo mapa real quando o PostGIS estiver populado.
