# Auditoria do repositório — 09/09/2026

## Objetivo e escopo

Esta auditoria foi feita somente para inspeção. O objetivo foi verificar
duplicação de código, arquivos sem referência aparente e o estado atual do
fluxo de redistritamento antes de planejar a automação. Nenhum arquivo foi
apagado, movido ou renomeado.

Foram analisados:

- `Backend/` (código Go, rotas, serviços, repositórios e modelos);
- `Frontend/src/` (React, utilitários, testes e estilos usados pelo Vite);
- `Frontend/public/` (arquivos publicados como estáticos);
- `Frontend/assets/` (arquivos históricos estáticos);
- `scripts/` (ferramentas, entradas, relatórios, backups e fixtures);
- referências textuais no código, no build e na documentação.

## Resultado da auditoria de código

### Duplicação exata

Foi feita comparação SHA-256 dos arquivos `*.go`, `*.js`, `*.jsx`, `*.css`,
`*.py` e `*.html` em `Backend/`, `Frontend/src/` e `scripts/`. Não foi
encontrado outro arquivo com conteúdo exatamente igual dentro desse escopo.

Isso não prova que não exista lógica semanticamente parecida, mas elimina a
hipótese de cópias exatas de arquivos de código que possam ser removidas com
segurança.

### Responsabilidades parecidas que devem permanecer separadas

Há normalização de texto no backend (`Backend/services/endereco_resolver.go`)
e no frontend (`Frontend/src/utils/buscaRua.js`). São implementações em
camadas diferentes, usadas para resolver consultas em momentos diferentes.
Não são arquivos mortos nem devem ser unificados por remoção. A melhoria
recomendada é manter casos de teste de paridade quando a regra de busca mudar.

O serviço Go e o painel React de redistritamento também possuem lógica
relacionada, mas com responsabilidades distintas: o backend valida e grava o
plano; o frontend apresenta e edita o rascunho. Essa separação é esperada.

## Arquivos com uso confirmado

Os seguintes pontos foram verificados como parte do runtime ou do fluxo de
manutenção e não são candidatos a remoção:

- `Backend/services/redistritamento_service.go`,
  `Backend/repositories/redistritamento_repository.go` e os modelos do plano;
- `Frontend/src/components/RedistritamentoPanel/` e
  `Frontend/src/api/redistritamento.js`;
- `Frontend/src/components/DistrictMap.jsx`, `Frontend/src/data/funcoes.js`,
  `Frontend/src/data/mockRuas.js` e `Frontend/src/api/zeRota.js`;
- testes `Frontend/src/utils/buscaRua.test.js` e
  `Frontend/src/components/OrdenamentoPanel/ordemManual.test.js`;
- scripts de aplicação, revisão, auditoria, backup e os relatórios de
  geometrias. Os arquivos históricos documentam alterações já aplicadas e não
  devem ser tratados como código morto automaticamente.

## Candidatos a validação manual

Estes itens não aparecem referenciados pelo entrypoint atual do Vite ou por
outro código pesquisado. Eles foram preservados porque podem servir como
material histórico, fixture de teste manual ou recurso aberto diretamente no
navegador:

### Legado estático

`Frontend/assets/` contém a antiga árvore estática:

- `css/chat.css`
- `css/style.css`
- `css/styledistritos.css`
- `images/cabecalho.jpg`
- `images/iconcorreios.webp`
- `images/image.png`
- `images/image.webp`
- `js/chat.js`

O build atual começa em `Frontend/index.html` e carrega `/src/main.jsx`;
nenhuma referência a `Frontend/assets/` foi encontrada no código atual. Antes
de remover esse diretório, confirmar se alguém ainda abre esses arquivos
diretamente ou os usa fora do Vite.

### Imagens sem referência atual

Não foram encontradas referências no código atual para:

- `Frontend/public/images/iconcorreios.png`
- `Frontend/public/images/iconcorreios.webp`
- `Frontend/public/images/zerotatransparente.png`

`zerotatransparente.png` é byte a byte igual a
`ze-rota-avatar.png` e `ze-rota-corpo.png`. A imagem `iconcorreios.webp` em
`Frontend/assets/images/` também é igual à cópia em `Frontend/public/images/`.
Essas duplicações são candidatas a consolidação depois de uma confirmação
visual e de uma busca no histórico de deploy.

`logocorreios.png` e `logocorreios.webp` também são byte a byte iguais, mas
ambas têm uso declarado hoje: a PNG é carregada pelos componentes React e a
WEBP é usada no preload do `index.html`. Não remover sem primeiro ajustar
essas referências.

### Artefatos de revisão

`scripts/revisoes_comparacao_geometrias_manuais (1).json` parece ser uma cópia
gerada pelo navegador, não é referenciada por scripts ou documentação atuais e
tem um sufixo de duplicidade. Ainda assim, é um registro de revisão e deve ser
mantido até que o operador confirme que o arquivo não é necessário.

`scripts/etiqueta-teste-cep.html` e
`scripts/etiqueta-teste-correios.html` não são importados pelo app. São
fixtures para teste manual do scanner e devem permanecer enquanto o teste de
etiquetas continuar fazendo parte do processo.

Os arquivos datados em `scripts/` (backups, CSV, JSON, SQL, TSV e TXT) são
evidências de auditoria ou entradas de ferramentas. Data no nome não é, por
si só, sinal de arquivo morto.

### Arquivos gerados fora do controle de versão

`Backend/tmp/`, `Frontend/dist/` e `Backend/backend.exe` são saídas locais
ignoradas pelo Git (`.gitignore`). Não fazem parte do pacote fonte e não foram
alterados nesta auditoria. Podem ser limpos em uma manutenção específica de
ambiente, mas isso não faz parte desta tarefa.

O arquivo `nota-de-status-site-correios.md` foi preservado integralmente por
ser o documento de contexto e operação do projeto.

## Auditoria do redistritamento atual

O fluxo existente confirma que:

1. a redução escolhe os maiores códigos de distrito ativos para extinção;
2. as ruas órfãs são listadas em um plano;
3. o administrador escolhe o destino de cada rua, individualmente ou por
   grupo de origem;
4. somente depois de todas as escolhas o plano pode ser concluído e aplicado
   em transação;
5. não existe ainda recomendação automática baseada em geometria, capacidade,
   tempo de percurso ou equilíbrio de carga.

Portanto, a automação planejada na Fase 9 é uma evolução do fluxo atual. Ela
deve começar como simulação e recomendação, mantendo a aprovação explícita do
administrador. A aplicação definitiva continuará protegida por backup,
transação e auditoria.

## Conclusão

Não há duplicação exata de código que justifique remoção imediata. Existem
arquivos legados e duplicatas de assets que merecem uma limpeza posterior,
mas todos permanecem preservados nesta etapa. O próximo trabalho de código é
implementar a Fase 9 do planejamento, depois de validar os dados de geometria,
capacidade dos distritos e a planilha de tempos de percurso.
