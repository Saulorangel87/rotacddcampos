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
workflow manual. A próxima entrega está identificada como versão `1.4.0` e
inclui o ordenamento estrito por proximidade passo a passo.

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

### Auditoria de coordenadas e ordenamento — 06/09/2026

As 1.810 geometrias preenchidas na base local passaram por validação de JSON,
tipo, pontos válidos, limites municipais amplos e degeneração: não foram
encontradas geometrias inválidas. Existem 305 ruas sem geometria, e o campo
`rota` está vazio em todas as ruas; por isso, não há como inferir uma ordem
operacional a partir do cadastro. A próxima alteração do motor deve ser medida
contra uma sequência real de paradas ou uma planilha de ordem por rua. Até lá,
o algoritmo mantém apenas a proteção objetiva de começar pela parada mais
próxima do CDD, sem incorporar preferências de rota não medidas.

A planilha ou sequência operacional real está pendente, com previsão de obtenção
em até dois dias. Enquanto ela não chega, a próxima frente é revisar permissões
e preparar a validação final para produção.

### Auditoria do motor de proximidade — 06/09/2026

O motor foi testado com casos determinísticos e uma simulação de 1.000 conjuntos
de ruas da base local. Em cada passo, a saída escolhe a menor distância entre a
posição atual e as paradas restantes, começando no CDD. Não há mais uma
reordenação global posterior. O resultado continua sendo uma sugestão em linha
reta; vias, trânsito e ordem dos números ainda dependem de dados operacionais
reais.

No caso corrigido com Araújo Silva e Advaldo Maciel, o cadastro local indica
aproximadamente 244 m entre os centros usados pelo motor e 113 m entre os
pontos mais próximos das geometrias. A sequência estrita atende essa
proximidade. Alcides Vieira Maciel era a rua citada por engano e fica em outra
posição do cadastro. Nenhuma coordenada foi modificada.

### Auditoria do traçado usado no cálculo — 07/09/2026

Uma nova inspeção separou a regra de ordenamento da qualidade da coordenada.
O laço do motor já inicia no CDD e escolhe a menor distância disponível em cada
transição. O problema encontrado estava na entrada: `coordenadaDasGeometrias`
calculava um centro médio com todos os vértices. Como o importador do OSM reúne
ways de um mesmo nome antes de gravar a rua, alguns registros têm trechos
desconectados em bairros distintos; o centro médio pode ficar quilômetros fora
de qualquer trecho real.

O cálculo foi ajustado para preservar os vértices e, durante cada transição,
medir a menor distância entre a posição atual e os pontos do traçado. A
posição atual passa a ser o ponto escolhido do trecho, mantendo a regra de
vizinho mais próximo e o ponto inicial fixo do CDD. Fallbacks externos que
possuem apenas uma coordenada continuam funcionando como antes.

No banco local foram auditadas 2.042 geometrias válidas no formato; 207 delas
ficam a mais de 500 m do próprio centro médio e são candidatas a revisão de
associação por distrito/CEP. Isso não é tratado como coordenada corrigida: o
motor já deixa de depender da média, mas a confirmação da rua correta ainda
deve ser feita com o cadastro operacional ou desenho manual. A comparação por
ID dos hashes JSONB das 2.115 linhas ativas encontrou zero divergências entre o
banco local e a produção; a produção também mantém 73 ruas sem geometria.

Validação: `go test ./... -count=1` no backend, `git diff --check` e build do
frontend após a atualização da versão. Nenhuma geometria ou tabela foi
alterada nesta auditoria.

### Auditoria de nomes repetidos e proteção do casamento OSM — 07/09/2026

A hipótese de coordenadas trocadas foi confirmada como risco do processo de
casamento: o importador agrupava todos os ways com a mesma chave-base de nome
antes de gravar a geometria. Em nomes repetidos por bairro, isso podia colocar
trechos de regiões distintas no mesmo cadastro.

O relatório local encontrou 97 grupos de nomes exatos repetidos, envolvendo 232
cadastros, e 127 chaves-base com contextos diferentes. `NOSSA SENHORA DA PENHA`
é um exemplo com três cadastros, dois bairros e três CEPs compartilhando o
mesmo hash de geometria. Esses números são triagem; não significam que toda
linha esteja comprovadamente errada.

Foi criado `scripts/auditar_duplicidades_geometrias.py`, somente leitura, com
relatórios por cadastro e por grupo. O casamento OSM passou a usar
`scripts/osm_agrupamento.py`: ways do mesmo nome são separados por continuidade
geográfica, correspondências com mais de um componente não são aplicadas e vão
para `scripts/revisao_matches_ambiguos.csv`. O aplicador de revisão exige
`componente` quando a escolha tiver mais de um grupo. Nenhum banco foi alterado
nesta etapa.

Também foi criado `scripts/revisar-duplicidades.html`. O operador carrega o
relatório, seleciona um cadastro, consulta os ways daquele nome no Overpass e
escolhe visualmente o componente contínuo correspondente ao bairro/CEP. As
decisões são salvas no navegador e exportadas em
`decisoes_duplicidades_osm.json`; a ferramenta não grava no banco. O aplicador
continua exigindo o campo `componente` para impedir uma associação ambígua.
Quando a revisão estiver concluída, use
`python scripts/aplicar_revisao_osm.py --arquivo decisoes_duplicidades_osm.json`;
as credenciais são lidas do ambiente ou de `Backend/.env`.
Durante a revisão foi identificado que um componente OSM pode ter o nome
correto, mas representar um trecho maior que o cadastro postal. Esses casos
ficam pendentes e deverão ser recortados por segmento antes de qualquer
aplicação.

Para esse cenário foi criado um fluxo separado: `gerar_revisao_geometrias_existentes.py`
gera a lista com as geometrias atuais, `revisar-geometrias-existentes.html`
permite desenhar o trecho corrigido e `aplicar_revisoes_geometrias_manuais.py`
valida e aplica o JSON somente quando executado com `--aplicar`.

Para os cadastros que ficaram pendentes na revisão de duplicidades, o lote deve
ser regenerado com `python scripts/gerar_lote_revisao_manual.py`. A ferramenta
`desenhar-revisoes-pendentes.html` é a versão de desenho iniciada do zero e
contém 157 pendências atuais, separadas das 73 ruas históricas sem geometria.

Em 08/09/2026, 95 desenhos foram validados e aplicados no banco local, com
backup dos registros afetados e conferência pós-aplicação sem divergências. A
produção ainda não foi alterada. O relatório atual das 73 ruas sem geometria
está em `scripts/relatorio_ruas_sem_geometria.csv`, com a lista limpa de nomes
em `scripts/relatorio_ruas_sem_geometria_nomes.txt`.

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

- **73 ruas sem geometria após a aplicação auditada de 228 desenhos em
  07/09/2026** — o lote foi aplicado localmente e na produção em transação;
  13 nunca tiveram correspondência apresentada pelo OSM e permanecem para
  revisão/desenho manual
- **Auditoria cadastral** — os bancos mantêm 2.115 ruas ativas e o mesmo hash
  de geometrias (`7f0371aa84d767a5ed0438572366cf9d`). Os IDs 721 e 722 têm
  nomes diferentes entre local e produção, embora bairro, CEP e distrito
  coincidam; nenhum deles foi alterado pelo lote de geometrias.
- **Fluxo de AUMENTO do Redistritamento** — banco já preparado, falta
  lógica de negócio (service/handler) e tela
- **Plano de ordenamento por rua**: planilha com a sequência real de
  numeração de entrega ainda não chegou — vai virar tabela própria
  (`rua_id`, `ordem`, `número`)
- **Zé Rota — próxima fase**: sugestão de rota pra múltiplas encomendas,
  depende do ordenamento acima
- **Precisão do ordenamento**: as geometrias existentes foram auditadas e estão
  válidas; ainda falta uma sequência operacional real para medir e ajustar o
  resultado antes da publicação
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
ssh -o KexAlgorithms=curve25519-sha256 -i "C:\Users\saulo\Documents\Chave VM 12RAM\ssh-key-2026-06-23.key" ubuntu@100.67.151.30
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

Os scripts `casar_ruas_osm.py` e `aplicar_revisao_osm.py` foram atualizados para
separar componentes geográficos e bloquear associações ambíguas. Os demais
scripts (`listar_ruas_sem_match.py` e `aplicar_geometria_manual.py`) continuam
com o fluxo anterior.
