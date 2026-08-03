# Nota de status — Site Correios (CDD Campos dos Goytacazes)
_Atualizado em 02/08/2026_

## Visão geral

Ferramenta interna da unidade CDD Campos dos Goytacazes: mapa interativo dos 24 distritos postais (601–624), consulta de ruas/CEP, cadastro de colaboradores, e ajuste de rotas (mover rua entre distritos).

- **Frontend**: React + Vite, mapa real com **React Leaflet + OpenStreetMap**
- **Backend**: Go (Fiber + GORM), PostgreSQL 17 local (`rotas_db`)
- **Banco**: 2122 ruas ativas, 49 colaboradores, 24 distritos com polígono real

## Modelo de acesso combinado (ainda não implementado — é o próximo passo)

- **Visitante / usuário comum (qualquer pessoa, sem login)**: acesso total de **consulta** — mapa, busca de trecho/rota, e **imprimir a listagem de ruas** (ex: carteiro novo confere sua rota e imprime). Nada disso pode ficar atrás de login.
- **Administrador (login obrigatório)**: único nível com permissão de editar/cadastrar — mover rua de distrito, criar/excluir rua, criar/excluir colaborador, etc. Por enquanto só **2 pessoas**: Saulo e o gerente.
- Login simples: matrícula + senha (bcrypt), JWT (~8h de validade, carrega id/matrícula/papel), guardado em `localStorage` + header `Authorization`. Trava real sempre no backend (middleware confere o papel antes de qualquer escrita); o front só esconde/mostra botão por UX.
- Diagrama de arquitetura de acesso já desenhado e aprovado nesta conversa (visitante/consulta → só leitura; admin → leitura + escrita).

## O que já está funcionando

### Backend (Go)
- Sobe com `go run main.go`, porta 8080, CORS liberado pro front (`localhost:5173`)
- Tabelas: `ruas`, `colaboradores`, `distritos`, `historico_alteracoes` (+ `carteiros` e `rotas`, criadas à toa em algum momento, vazias, sem uso)
- `ruas`: CRUD completo (`GET/POST/PUT/DELETE /ruas`), campo `geometria` (GeoJSON do OpenStreetMap) e `ativo` (soft-delete)
- `colaboradores`: CRUD completo, incluindo `?carteiro=true`, `/aniversariantes-hoje?data=DD/MM`
- `distritos`: `GET /distritos` e `/distritos/:codigo`, cada um com `geo_json` (polígono real, chave = código tipo "601")
- `historico_alteracoes`: grava **sozinho**, automaticamente, toda vez que `PUT /ruas/:id` muda o distrito de uma rua de verdade (não precisa chamar endpoint separado)
- `GET /historico?pagina=&limite=` — paginado
- `GET /estatisticas/operacao` — números da unidade (distritos, colaboradores, motos/carros/ciclistas/interno/administrativo/OTT/OT/supervisor/gerente), tudo calculado ao vivo do banco, nunca fixo no código

### Frontend (React)
- Header com logo real, 24 distritos com cores únicas, **nenhum distrito selecionado por padrão**
- **Mapa real (Leaflet + OpenStreetMap)**: os 24 polígonos de distrito vêm do banco (`GET /distritos`), com fallback pro arquivo estático se a API cair. Clique no polígono ou no botão do distrito: destaque forte no contorno + zoom automático + alfinete no centro
- **Camada opcional "Ruas reais (OSM)"**: liga/desliga por cima do mapa de distritos, mostra o traçado real das ruas do distrito selecionado (dado do OpenStreetMap, casado por nome — ver seção abaixo)
- **Ajustes de Rotas**: assistente de 3 passos (selecionar ruas → escolher distrito → confirmar), paginado, altura travada na tela
- **Ruas**: tabela com abas, busca (corrigida: distrito exato / CEP por prefixo / texto livre), paginação 30, exportar CSV, **imprimir** (tabela própria com bordas, mesmo filtro do exportar), **+ Nova rua** e excluir por linha (com modal de confirmação estilizado)
- **Colaboradores**: modal com busca, **+ Novo** (função por seletor fixo, não texto livre — evita quebrar a contagem de estatísticas), excluir (mesmo modal de confirmação)
- **🎂 Aniversariante do dia**: ícone no header, popover ao passar o mouse, busca de `/colaboradores/aniversariantes-hoje`
- **CEP**: busca por nome de rua → mostra CEP/bairro/distrito
- **Relatórios**: paginado de 10, mostra mudanças da **sessão atual do navegador** (ainda não persistente na tela — o backend já grava no `historico_alteracoes`, só falta o front puxar de lá em vez do estado local)
- **Operação em números**: faixa embaixo do mapa com todos os totais, atualiza sozinha quando colaborador é criado/excluído (sem precisar F5)

## Higienização de dados (concluída)

- Banco de 2264 → **2122 ruas ativas** (duplicidade removida direto pela interface)
- Planilha higienizada importada e casada por CEP (script SQL com staging table), sem perder histórico
- Restaram 11 CEPs com nomes ligeiramente diferentes → **3 resolvidos manualmente** (Pedro Maciel Neto, Estrada Antônio Moacir Batista, Rua Saldanha Marinho), os outros 8 eram legítimos (rua cruzando fronteira de 2 distritos, não é duplicidade)

## Ruas ↔ OpenStreetMap (concluído, com pendência de revisão)

- Script `casar_ruas_osm.py` (Overpass API + fuzzy matching): **1252 de 2115 ruas ativas** casaram com confiança alta (≥85%) e já têm geometria real gravada em `ruas.geometria`
- **751 ficaram em `revisao_matches_baixos.csv`**, confiança 60-84%, não aplicadas automaticamente — pendente de revisão manual quando houver tempo
- Detalhe técnico que já foi resolvido, se precisar rodar de novo: Windows + `psycopg2` tem bug de encoding conhecido, script já corrige forçando `PGCLIENTENCODING=LATIN1` antes de conectar

## Pendências conhecidas

1. **Autenticação** — é o próximo passo combinado, modelo já definido acima, falta implementar (tabela `usuarios`, bcrypt, JWT, middleware, telas de login)
2. **Relatórios persistente** — trocar o front pra puxar de `GET /historico` em vez do estado local da sessão
3. **751 ruas** em `revisao_matches_baixos.csv` aguardando revisão manual de nome
4. **Botão "Por Carteiro"** na tabela de ruas não filtra de verdade (só reordena) — combinamos deixar parado, sem prioridade
5. Ajuste visual pequeno: texto "Correios" um pouco desalinhado do ícone no selo do header
6. **Tabelas `carteiros` e `rotas`** existem no banco vazias, sem uso — foram criadas sem querer em algum momento, podem ser ignoradas ou removidas depois
7. Modelo relacional completo (carteiros→rotas→ruas com FK) da documentação técnica original nunca foi implementado — decidimos reaproveitar `colaboradores` em vez disso, então essa pendência está efetivamente substituída/resolvida por outro caminho

## Próxima ação combinada

Implementar autenticação: tabela `usuarios` (matrícula, senha com hash, papel: admin), login com JWT, middleware no backend travando `POST/PUT/DELETE` pra admin, e no front esconder os botões de editar/cadastrar/excluir de quem não estiver logado como admin. Consulta, busca e impressão continuam livres pra qualquer pessoa, sem login.
