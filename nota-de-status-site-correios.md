# Notas de Deploy — Guia de Logística CDD Campos

_Atualizado em 06/09/2026_

## Status atual

**No ar em produção**, na VPS (Oracle Cloud), via Docker Compose.
Domínios: `cddcampos.devsaulo.com.br` (frontend) e
`cddcampos-api.devsaulo.com.br` (backend).

Deploy agora é via **GitHub Actions** (workflow_dispatch, disparo manual na
aba Actions do repo) — ver seção própria abaixo. Deixou de ser só
`git pull` + `docker compose up -d --build` manual na VPS.

As melhorias de interface e ordenamento descritas na atualização de 06/09
foram validadas localmente e ainda aguardam publicação em produção pelo
workflow manual.

## O que foi feito desde a última nota (10/08)

### Acesso restrito ao público
- Todo o site agora exige login — antes várias rotas (mapa, ruas,
  distritos, aniversariantes, folgas, observações, Zé Rota, estatísticas)
  eram públicas sem querer
- Só `/health` e `/swagger` continuam públicos (infraestrutura)
- Sistema tem só **2 papéis de verdade**: `admin` e `colaborador` (não 3)
- Frontend ganhou tela de bloqueio (`AcessoRestrito.jsx`) — header/logo
  continuam visíveis pra quem não logou, conteúdo fica bloqueado com
  prompt de login
- **49 logins criados em lote** pros colaboradores: login = matrícula,
  senha provisória = `*` + matrícula, `senha_provisoria = true` força troca
  no primeiro acesso (mesmo fluxo que já existia). Feito via SQL puro
  usando `pgcrypto` (`crypt()` + `gen_salt('bf', 10)`, mesmo custo do
  bcrypt do Go) — sem precisar de script externo
- `GET /distritos` agora só retorna distritos **ativos** (importante pro
  Redistritamento, ver abaixo)

### CI/CD
- O `deploy.yml` antigo (GitHub Pages, nunca ativado) foi **substituído**
  por um deploy de verdade via SSH
- Disparo **manual** (`workflow_dispatch`, botão "Run workflow" na aba
  Actions) — separa dar `push` de publicar em produção, importante agora
  que colegas usam o site de verdade
- Job `build-check` (compila o Backend em Go + builda o Frontend em React)
  roda antes do deploy — não publica se o build quebrar
- Pré-requisito de infra: usuário `ubuntu` adicionado ao grupo `docker` na
  VPS (`sudo usermod -aG docker ubuntu`), pra não precisar de `sudo`
  interativo num script rodando sozinho
- Secrets no GitHub: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`
- Cache do Go (`go.sum`) ajustado — reduziu o build-check de ~2m40s pra
  ~1m10s

### Redistritamento (feature nova)
Menu novo "Redistritamento" na sidebar (admin only), logo abaixo de
"Mapa Geral".

- **Fluxo de REDUÇÃO** (implementado por completo): admin escolhe reduzir
  a quantidade de distritos (ex: 24 → 17); o corte é sempre **global**, do
  maior código de distrito pro menor (não separado por setor
  Campos/Guarus); as ruas dos distritos extintos ficam "órfãs" e são
  realocadas **manualmente**, uma por uma ou em lote por grupo, pra um
  distrito sobrevivente
- Dois botões: **Concluir** (salva progresso, ainda editável/revisável) e
  **Aplicar** (definitivo, com modal de confirmação — muda `ruas.distrito`,
  desativa (`ativo = false`, soft-delete) os distritos extintos, grava
  histórico de auditoria)
- Botão **"Voltar ao início"** pra descartar um rascunho e recomeçar (só
  funciona antes de Aplicar — depois disso é definitivo mesmo)
- Distritos extintos só somem do seletor de chips (`601, 602...`) **depois**
  de Aplicar — até lá, carteiros continuam vendo os 24 normal, sem impacto
  na operação real durante o planejamento
- Novas tabelas: `planos_redistritamento`, `plano_redistritamento_ruas`;
  `distritos` ganhou coluna `ativo` (nunca apaga distrito de verdade, só
  desativa — permite reverter no futuro)
- **Testado com sucesso em produção**: removido o distrito 624 (120 ruas
  realocadas)
- **Fluxo de AUMENTO**: só **desenhado no banco** (`Tipo` aceita
  `"aumento"`, `Distrito.Ativo` já é soft-delete) — **ainda sem lógica de
  negócio nem tela**. Pendente pra quando a unidade precisar ganhar
  distrito de volta

### Geometria das ruas (mapa mais preciso)
No início da sessão: **442 ruas sem geometria** (de 2.115 ativas).

- **Nominatim** (geocodificador gratuito do OpenStreetMap, sem chave):
  125 ruas achadas de verdade pelo nome (as outras 317 caíram num
  fallback ruim por CEP, descartadas por imprecisão — CEP no Brasil o
  Nominatim resolve mal, 226 ruas diferentes caíam no mesmo ponto genérico)
- **Upgrade de traçado**: reconsultando o Nominatim com
  `polygon_geojson=1`, 114 das 125 (91%) conseguiram o traçado real da via
  (não só um ponto) — as outras 11 continuam como ponto
- **Overpass API** (todas as ruas nomeadas da cidade numa consulta só, via
  bounding box): mais 59 casadas, mas só **14 com confiança alta**
  (score de similaridade ≥ 0.88) foram aplicadas — as outras 45 eram falsos
  positivos perigosos (nome parecido, rua errada — ex: "Alexandre Dumas"
  casando com "Alexandre Vargas")
- Descoberta de dado: muita rua no banco tem o tipo de logradouro
  **invertido** ("Manoel P. Barbosa, Rua" em vez de "Rua Manoel P.
  Barbosa") — o script do Overpass já trata isso (`desinverter_tipo()`)
- **No snapshot anterior: 303 ruas ainda sem geometria** — prováveis loteamentos não
  mapeados no OSM, seguem pra desenho manual (ferramenta que já existe),
  sem mais atalho automático gratuito disponível
- Considerado (e descartado) usar Google Maps API: exige pré-pagamento de
  R$150 no Brasil pra ativar faturamento, mesmo pra uso dentro da cota
  gratuita — não compensou pro volume necessário

### Bugs achados e corrigidos no caminho
- `buscarDistritosGeoJSON` usava `fetch` sem token — desde a restrição de
  acesso público, sempre falhava silenciosamente e caía pro GeoJSON
  estático antigo em vez do banco. Corrigido pra usar `apiFetch`
- `OperacaoResumo` (a faixa "operação em números" embaixo do mapa) tinha o
  mesmo bug — sumia sem erro nenhum depois de `/estatisticas/operacao`
  virar rota autenticada. Corrigido
- Ícone de marcador quebrado no Leaflet pra geometria tipo `Point` — o
  ícone padrão do Leaflet depende de uma imagem que quebra com o build do
  Vite. Corrigido usando o mesmo pino (`divIcon`, HTML/CSS puro, sem
  imagem) já usado no marcador de distrito, só menor e sem texto
- Bug de flexbox no painel do Redistritamento: `overflow: hidden` no card
  de cada grupo zerava o tamanho mínimo automático, fazendo o navegador
  espremer os grupos em vez de rolar — corrigido com `flex-shrink: 0`

### Atualização local de 06/09/2026 — interface e ciclo de carga
- Ícones funcionais foram padronizados em SVG de linha, sem emojis.
- O cabeçalho passou a diferenciar `admin` (escudo) e `colaborador`
  (identificação de usuário), com rótulos próprios.
- O marcador do ponto de partida agora representa uma unidade postal e
  mantém a sigla `CDD`; o título ao lado mostra `Campos dos Goytacazes`.
- A marca da lateral foi reorganizada com o logotipo dos Correios, a
  identificação da unidade operacional e a descrição do painel.
- Ao limpar uma carga, objetos e paradas são removidos em transação e o
  horário do ordenamento é reiniciado para a nova carga; histórico e memória
  pessoal permanecem intactos.
- O rodapé oferece instalação do PWA somente em layout móvel. O botão também
  orienta a instalação pelo menu do navegador quando não há prompt nativo e
  fica oculto depois que o app é instalado.
- A busca e a entrada por CEP continuam cobertas pelos testes anteriores;
  nenhuma permissão foi ampliada nesta atualização.

### Validação local da busca — 06/09/2026

Uma consulta somente leitura ao PostgreSQL local encontrou 2.115 ruas em 24
distritos, com 1.810 geometrias e 305 ruas sem geometria. Os casos `Silva
Tavares`, `sergio`, `Santa Cecilia` e o CEP `28010-562` retornaram as opções
esperadas. A busca `Av. Sete Setembro` também foi conferida: a interface
remove a falsa coincidência de `Rua Vinte e Oito de Setembro`, causada pelo
prefixo `Sete` dentro de `Setembro`, sem perder a digitação progressiva.

Esse número é um retrato da base local nesta data e não substitui a conferência
da produção no momento da publicação.

### Scripts novos (fora do Docker, rodam local no PC)
- `preencher_geometria_nominatim.py` — geocodifica por nome via Nominatim
- `upgradar_tracado_nominatim.py` — tenta upgradar ponto pra traçado real
- `casar_ruas_overpass.py` — baixa todas as ruas da cidade via Overpass e
  casa por similaridade de nome (com correção de tipo invertido)
- `scripts/criar_logins_colaboradores.sql` — cria login em lote a partir
  da tabela `colaboradores` (SQL puro, usa `pgcrypto`)

### Ambiente local (PC do Saulo)
- Backend local usa PostgreSQL 17 nativo do Windows (porta 5432, via
  pgAdmin) — **não** o "Docker CDD" (outro servidor Postgres, sobra de um
  teste antigo com o `docker-compose.yml` local, mesmo nome de banco por
  dentro do container)
- Fluxo de sincronizar local com produção: `pg_dump -F c` na VPS →
  `scp` pro PC → Restore no pgAdmin com **"Clean before restore"**
  ligado (essencial, senão dá erro de "já existe")

## Pendências conhecidas / combinadas pra próxima sessão

- **305 ruas sem geometria no snapshot local de 06/09/2026** — desenho manual,
  ritmo próprio, sem mais
  atalho automático gratuito disponível
- **Fluxo de AUMENTO do Redistritamento** — banco já preparado, falta
  lógica de negócio (service/handler) e tela
- **Plano de ordenamento por rua**: planilha com a sequência real de
  numeração de entrega ainda não chegou — vai virar tabela própria
  (`rua_id`, `ordem`, `número`)
- **Zé Rota — próxima fase**: sugestão de rota pra múltiplas encomendas,
  depende do ordenamento acima
- **Precisão do ordenamento**: validar coordenadas ausentes e sequência
  sugerida com listas operacionais reais antes da publicação
- **AGC (Agência Comunitária)**: áreas sem entrega domiciliária, ainda não
  modelado no banco
- **Botão "Contribuir" no sidebar**: alimentar o Zé Rota com
  características gerais de distrito
- Gap de segurança aceito por ora: resetar/bloquear usuário não invalida
  um token JWT já emitido (fica válido até expirar sozinho em 8h)
- Botão "Por Carteiro" na tabela de ruas não filtra de verdade, só
  reordena (combinado deixar parado, sem prioridade)

## Atualizações de rotina

No PC:
```bash
git add .
git commit -m "..."
git push
```

Publicar (agora via CI/CD, não é mais automático nem manual na VPS):
1. Vai em `github.com/Saulorangel87/rotacddcampos` → aba **Actions**
2. Clica em **"Deploy Produção (VPS)"** → **Run workflow**

Se precisar rodar algo direto na VPS (scripts SQL avulsos, backup manual):
```bash
ssh -i "C:\Users\saulo\Documents\Chave VM 12RAM\ssh-key-2026-06-23.key" ubuntu@157.151.24.49
```

## Comandos únicos disponíveis (`Backend/cmd/`)

Todos rodam dentro do container já buildado:
```bash
docker exec -it rotas_api ./nome-do-binario
```
- `seed-admin` — cria o primeiro usuário admin
- `criar-usuario` — cria usuário adicional
- `importar-folgas-iniciais [-confirmar]` — já rodado; idempotente

## Scripts Python avulsos (fora do Docker)

Os scripts de geometria (Nominatim/Overpass) rodam **local no PC**, contra
um CSV exportado do banco de produção — não precisam de rede especial nem
container Python, só `pip install requests`. Geram um CSV de revisão pra
conferência humana antes de qualquer SQL ser aplicado em produção.

Scripts mais antigos (`casar_ruas_osm.py`, `aplicar_revisao_osm.py`,
`listar_ruas_sem_match.py`, `aplicar_geometria_manual.py`) continuam como
estavam, descritos na nota anterior.
