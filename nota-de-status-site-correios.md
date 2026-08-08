# Notas de Deploy — Guia de Logística CDD Campos

_Atualizado em 08/08/2026_

## Status atual

**No ar em produção**, na VPS (Oracle Cloud), via Docker Compose.
Domínios: `cddcampos.devsaulo.com.br` (frontend) e
`cddcampos-api.devsaulo.com.br` (backend).

## O que foi feito desde a última nota (04/08)

- **Primeiro deploy concluído**: dados reais migrados (pg_dump local →
  restore na VPS, corrigido problema de versão do Postgres — igualado
  pra `postgres:17-alpine` nos dois lados)
- **Repaginação visual completa**: ícones de linha (SVG próprio, sem lib
  nova) substituindo emoji no Sidebar e nos cards de estatística, cards
  de "Operação em números" com badge de ícone, Sidebar com cartão de
  marca fixo logo abaixo do menu (não mais grudado no fundo da página —
  bug corrigido), pills de distrito voltaram ao estilo original a pedido
- **Nova funcionalidade: Consulta de Folgas**
  - Modelo de livro-razão (`folgas_lancamentos`): créditos e débitos
    justificados, nunca sobrescritos
  - Consulta pública por matrícula exata (`GET /folgas/saldo`) — não
    lista ninguém, precisa digitar a matrícula certa
  - Lançar/excluir restrito a admin, sempre com auditoria (matrícula de
    quem fez vem do token, nunca do que o front manda)
  - Toda movimentação de folga também entra em `historico_alteracoes`
    (`tipo: "folga"`) e aparece junto com as movimentações de rua na aba
    Relatórios
  - **Importação do saldo real concluída**: 61 lançamentos, 35
    colaboradores, 121 folgas em aberto, importados a partir da
    planilha oficial (`cmd/importar-folgas-iniciais`) — conferido
    matrícula por matrícula, 100% confiável
- **Bug de fuso horário corrigido**: aniversariante aparecia com ~3h de
  antecedência (container Docker roda em UTC, Brasil é UTC-3). Fix:
  `time.FixedZone("America/Sao_Paulo", -3*60*60)` em vez de
  `time.Now()` puro, sem precisar instalar `tzdata` na imagem (Brasil
  não tem mais horário de verão desde 2019)
- `Backend/Dockerfile` agora compila dois binários (`main` +
  `importar-folgas-iniciais`), reutilizável pra futuros comandos
  `cmd/` sem precisar de outro Dockerfile

## Pendências conhecidas

- **Revisão do restante das ruas** — próximo passo combinado agora.
  Contexto: das ~2115 ruas ativas, 1252 já casaram com geometria real
  do OpenStreetMap (alta confiança) e 573 grupos de baixa confiança já
  foram revisados manualmente (306 aceitas). Ainda sobra revisar o
  resto que ficou de fora dessas duas levas.
- Alguns ajustes pontuais de digitação identificados na planilha de
  folgas durante a conferência final (nada estrutural, só correções
  pontuais de matrícula/valor — a se resolver direto pelo painel admin
  quando aparecer)
- Considerar apagar `.github/workflows/deploy.yml` (workflow antigo pro
  GitHub Pages, nunca ativado — o deploy real é VPS/Docker)
- Gap de segurança aceito por ora: resetar senha de um usuário não
  invalida token JWT já emitido (fica válido até expirar sozinho, hoje
  configurado pra 3h)

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

- `seed-admin` — cria o primeiro usuário admin
- `criar-usuario` — cria usuário adicional
- `importar-folgas-iniciais [-confirmar]` — importação de saldo (já
  rodado; idempotente, seguro rodar de novo sem duplicar)

Todos rodam dentro do container já buildado:
```bash
docker compose exec api ./nome-do-binario
```
