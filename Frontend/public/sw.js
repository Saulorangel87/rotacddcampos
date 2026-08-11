// Service Worker do PWA — cobre só o "casco" estático do site (JS, CSS,
// ícones, o HTML de entrada). NUNCA intercepta nem guarda em cache nada que
// pareça requisição da API — login, colaboradores, ruas, histórico, etc.
// sempre vão direto pra rede, sem passar pelo cache. Isso evita tanto servir
// dado desatualizado quanto guardar informação sensível no dispositivo.

const CACHE_NAME = "cdd-campos-v3";

const CAMINHOS_API = [
  "/auth",
  "/ruas",
  "/colaboradores",
  "/distritos",
  "/historico",
  "/estatisticas",
  "/health",
];

function ehRequisicaoDeApi(url) {
  return CAMINHOS_API.some((prefixo) => url.pathname.startsWith(prefixo));
}

function ehAssetEstatico(url) {
  return /\.(js|css|png|jpg|jpeg|svg|webp|ico|woff2?|json)$/i.test(
    url.pathname,
  );
}

self.addEventListener("install", (evento) => {
  self.skipWaiting();
});

self.addEventListener("activate", (evento) => {
  evento.waitUntil(
    caches
      .keys()
      .then((chaves) =>
        Promise.all(
          chaves.filter((c) => c !== CACHE_NAME).map((c) => caches.delete(c)),
        ),
      ),
  );
  self.clients.claim();
});

self.addEventListener("fetch", (evento) => {
  const req = evento.request;
  if (req.method !== "GET") return;

  const url = new URL(req.url);

  // Requisição pra API (mesmo em outro domínio/porta, tipo localhost:8080) —
  // nunca intercepta. Deixa passar direto pra rede, do jeito que o navegador
  // faria sem Service Worker nenhum.
  if (url.origin !== self.location.origin || ehRequisicaoDeApi(url)) {
    return;
  }

  // Navegação (abrir/recarregar a página): tenta rede primeiro, cai pro
  // cache só se estiver offline — garante que a versão mais nova sempre
  // aparece primeiro quando há internet.
  if (req.mode === "navigate") {
    evento.respondWith(
      fetch(req)
        .then((resposta) => {
          caches
            .open(CACHE_NAME)
            .then((cache) => cache.put("/index.html", resposta.clone()));
          return resposta;
        })
        .catch(() => caches.match("/index.html")),
    );
    return;
  }

  // Assets estáticos do próprio site: cache primeiro (mais rápido), busca na
  // rede e atualiza o cache em paralelo.
  if (ehAssetEstatico(url)) {
    evento.respondWith(
      caches.match(req).then((cacheado) => {
        const buscaRede = fetch(req)
          .then((resposta) => {
            if (resposta.ok) {
              caches
                .open(CACHE_NAME)
                .then((cache) => cache.put(req, resposta.clone()));
            }
            return resposta;
          })
          .catch(() => cacheado);
        return cacheado || buscaRede;
      }),
    );
  }
});
