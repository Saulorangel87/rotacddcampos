# Notas de Deploy — Guia de Logística CDD Campos

_Atualizado em 10/08/2026_

## Status atual

**No ar em produção**, na VPS (Oracle Cloud), via Docker Compose.
Domínios: `cddcampos.devsaulo.com.br` (frontend) e
`cddcampos-api.devsaulo.com.br` (backend).

## O que foi feito desde a última nota (08/08)

### Ruas / OpenStreetMap
- **Segunda leva de casamento com OSM rodada**: das 557 ruas sem geometria,
  30 ganharam geometria automática/revisada. Total hoje: **~1.671 de 2.115
  ruas ativas com geometria real** no mapa
- **414 ruas ficaram conscientemente sem geometria** (revisão manual
  rejeitou o nome sugerido pelo OSM — decisão certa, não pendência)
- **Ferramenta de desenho manual** (`desenhar-ruas-manual.html` +
  `scripts/aplicar_geometria_manual.py`) criada pras ruas que nunca bateram
  com nada no OSM (loteamentos internos, travessões). **83 de 112
  desenhadas e aplicadas**; as 29 restantes viraram planilha
  (`ruas_sem_localizacao.xlsx`) pra equipe de campo preencher ponto de
  referência
- **Bug de encoding corrigido**: os scripts Python (`casar_ruas_osm.py`,
  `aplicar_revisao_osm.py`) tinham um `client_encoding` manual que dobrava
  acentuação no CSV — removido
- **Segredo removido do Git**: `scripts/` tinha senha do Postgres hardcoded,
  corrigido pra ler de variável de ambiente; pasta tirada do `.gitignore`
  (não tem mais segredo, pode ser versionada normal)
- **Editar rua pelo site** (admin): botão ✏️ na tabela de Ruas, corrige
  nome/CEP/distrito/bairro direto, sem precisar mexer no banco
- **Busca tolerante a acento**: extensão `unaccent` do Postgres ativada;
  corrige de uma vez a Consulta de CEP, a tabela de Ruas e o Zé Rota

### Consulta de Folgas
- **Importação do saldo real concluída**: 61 lançamentos individuais (motivo
  + data reais, não um total resumido), 35 colaboradores, 121 folgas em
  aberto — importados via `cmd/importar-folgas-iniciais` (compilado junto
  do binário `main` no Dockerfile, roda com
  `docker compose exec api ./importar-folgas-iniciais -confirmar`)
- Conferido manualmente matrícula por matrícula — 100% confiável

### Observações de rua (conhecimento de campo)
- Nova tabela `rua_observacoes`: categoria fixa (Acesso / Segurança /
  Numeração irregular / Vários nomes / Outros) + texto livre
- Só admin cadastra/exclui; leitura pública. Botão 📝 na tabela de Ruas

### Zé Rota (assistente em chat)
- Chat com avatar/mascote (personagem "carteiro" gerado por IA, fundo
  removido de verdade), botão flutuante recolhível (avatar redondo) e
  **arrastável** (posição salva no navegador)
- Entrada por texto ou voz (reaproveita o reconhecimento de fala do
  CepLookup, extraído pro hook `useReconhecimentoDeVoz`)
- Backend Go com **function calling** — a IA nunca inventa nome de rua,
  sempre confirma via ferramenta `buscar_rua` batendo no banco real (que já
  inclui as observações de campo acima)
- **Roda de graça**: usa um Cloudflare Worker que o usuário já tinha
  (`flat-rice-6724.sauloleonardo1987.workers.dev`), proxy pro **Groq**
  (`llama-3.3-70b-versatile`, formato OpenAI-compatible) — não usa API paga
  da Anthropic
- Variável de ambiente: `ZE_ROTA_WORKER_URL` no `.env` de produção
- Retry automático no backend pra instabilidade conhecida do Groq/Llama
  ("Failed to call a function" intermitente)

### Bug corrigido: fuso horário do aniversariante
- `time.Now()` no container Docker roda em UTC; corrigido pra
  `time.FixedZone("America/Sao_Paulo", -3*60*60)` — sem precisar instalar
  `tzdata` na imagem (Brasil não tem mais horário de verão)

### Microfone em produção
- Bloqueado pelo Cloudflare (`Permissions-Policy: microphone=()` aplicado
  globalmente via Transform Rule pra todos os domínios da conta) — corrigido
  pra `microphone=(self)` direto no painel do Cloudflare

## Pendências conhecidas / combinadas pra próxima sessão

- **Plano de ordenamento por rua**: planilha com a sequência real de
  numeração de entrega (a caminhada do carteiro, não ordem numérica simples)
  ainda não chegou — vai virar tabela própria (`rua_id`, `ordem`, `número`)
- **Zé Rota — próxima fase**: sugestão de rota pra múltiplas encomendas
  (algoritmo geométrico com a geometria real + ordenamento, IA só explica o
  resultado em português — não decide a ordem sozinha) depende do
  ordenamento acima
- **AGC (Agência Comunitária)**: áreas sem entrega domiciliária — quer que
  o Zé Rota saiba qual endereço vai pra qual AGC. Ainda não modelado no
  banco, combinado planejar juntos
- **Botão "Contribuir" no sidebar**: pra alimentar o Zé Rota com
  características gerais de distrito (não só por rua) — combinado planejar
  juntos
- 29 ruas ainda sem geometria (planilha com a equipe de campo, aguardando
  retorno)
- Considerar apagar `.github/workflows/deploy.yml` (workflow antigo pro
  GitHub Pages, nunca ativado)
- Gap de segurança aceito por ora: resetar senha de um usuário não invalida
  token JWT já emitido (fica válido até expirar sozinho, hoje 3h)

## Atualizações de rotina

No PC:
```bash
git add .
git commit -m "..."
git push
```

Na VPS:
```bash
cd ~/apps/site-correios
git pull
sudo docker compose up -d --build
```

## Comandos únicos disponíveis (`Backend/cmd/`)

Todos rodam dentro do container já buildado:
```bash
docker compose exec api ./nome-do-binario
```
- `seed-admin` — cria o primeiro usuário admin
- `criar-usuario` — cria usuário adicional
- `importar-folgas-iniciais [-confirmar]` — já rodado; idempotente

## Scripts Python avulsos (`scripts/`, fora do Docker)

Rodam contra produção via container Python descartável ligado na rede do
Compose:
```bash
sudo -E docker run --rm -it --network site-correios_rotas_network \
  -v $(pwd)/scripts:/scripts -e DB_HOST=postgres -e DB_PASSWORD="$DB_PASSWORD" \
  python:3.12-slim bash -c "pip install -q requests psycopg2-binary && python /scripts/NOME.py"
```
- `casar_ruas_osm.py` — casamento automático com OpenStreetMap
- `aplicar_revisao_osm.py` — aplica decisões da revisão manual
- `listar_ruas_sem_match.py` — lista ruas que nunca bateram com nada
- `aplicar_geometria_manual.py` — aplica desenho manual
  (`desenhar-ruas-manual.html`, standalone, fora do Docker)

Todos já corrigidos pra ler senha de variável de ambiente (nunca hardcoded)
e escrever arquivos com caminho absoluto (não dependem da pasta em que
foram chamados).
