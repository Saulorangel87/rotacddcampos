# Notas de Deploy — Guia de Logística CDD Campos

_Atualizado em 04/08/2026_

## Status atual

Ainda **não publicado** na VPS. Rodando localmente (dev), com toda a
autenticação, PWA e revisão de dados já concluídos. Caminho de deploy
(Docker) preparado, faltando execução do primeiro deploy.

## O que já está pronto pro deploy

- `Backend/Dockerfile` — build multi-stage, sem segredo gravado na imagem
- `Frontend/Dockerfile` + `nginx.conf` — build Vite + serve estático,
  com `sw.js`/`manifest.json`/`index.html` sempre revalidados (nunca em
  cache agressivo, senão o Service Worker não atualiza sozinho)
- `docker-compose.yml` (raiz) — junta postgres + api + frontend
- `.env.example` (raiz) — variáveis que o compose de produção espera
- `.dockerignore` nos dois lados (não copia `.env`/`node_modules` pro build)
- CORS configurável por `CORS_ORIGINS` (não fica mais travado em localhost)
- Rate limit de login por IP (10/min) além do bloqueio por conta (5/15min)

## Pendente antes do primeiro deploy

1. **Escolher os subdomínios definitivos** — sugestão: `rotas.devsaulo.com.br`
   (frontend) e `rotas-api.devsaulo.com.br` (backend). Ajustar no `.env` da
   VPS e no Cloudflare/Nginx Proxy Manager.
2. **Migrar os dados reais** — o Postgres do `docker-compose.yml` sobe
   vazio. As ~2122 ruas, colaboradores e distritos estão no Postgres local.
   Precisa `pg_dump` local → `pg_restore`/`psql` na VPS.
3. **Gerar segredos de produção** — `JWT_SECRET` novo (`openssl rand -hex 32`)
   e senha forte pro Postgres, diferentes dos usados em dev.
4. **Configurar Cloudflare Tunnel + Nginx Proxy Manager** apontando os dois
   subdomínios pras portas do host (`8081` api, `8082` frontend, conforme
   o compose atual — ajustável).
5. **Criar o primeiro admin dentro do container** depois do primeiro
   `docker compose up`:
```bash
   sudo docker compose run --rm api go run cmd/seed-admin/main.go -matricula SUA_MATRICULA -senha "SenhaTemporaria123"
```
6. Considerar apagar o `.github/workflows/deploy.yml` (workflow antigo pro
   GitHub Pages, nunca ativado — Pages Source está em "None") já que o
   deploy real vai ser via VPS/Docker, pra não confundir depois.

## Deploy inicial (primeira vez)

```bash
cd ~/apps
git clone <url-do-repo> site-correios
cd site-correios
cp .env.example .env   # preencher com valores reais de produção
sudo docker compose up -d --build
```

## Atualizações de rotina (depois do primeiro deploy)

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

## Segurança — pontos já revisados

- `.env` nunca foi commitado no histórico do Git (conferido — zero
  ocorrências em `git log --all --full-history`)
- Nenhum segredo real (senha, JWT_SECRET) encontrado em nenhum commit
- Dockerfile do backend não grava mais `.env` dentro da imagem
- Gap conhecido, aceito por ora: resetar senha de um usuário não invalida
  token JWT já emitido (fica válido até completar 8h sozinho)