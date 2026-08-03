// Cliente HTTP central da API. Todo módulo em src/api/ deve chamar apiFetch
// em vez de fetch direto, por dois motivos:
//   1. Injeta automaticamente o header Authorization quando há token salvo.
//   2. Garante que erro (401, 403, 500...) sempre estoura como exceção —
//      nenhuma chamada de escrita deve "engolir" um erro e deixar a tela
//      achar que salvou quando na verdade a API recusou.
export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const CHAVE_TOKEN = 'rotas_token'

export function getToken() {
  return localStorage.getItem(CHAVE_TOKEN)
}

export function salvarToken(token) {
  localStorage.setItem(CHAVE_TOKEN, token)
}

export function limparToken() {
  localStorage.removeItem(CHAVE_TOKEN)
}

// Disparado quando a API responde 401 com um token presente — sinal de que o
// token expirou ou foi invalidado. O AuthContext escuta esse evento pra
// deslogar e mostrar a tela de login de novo, sem precisar prop-drilling.
const EVENTO_SESSAO_EXPIRADA = 'auth:sessao-expirada'

export async function apiFetch(caminho, opcoes = {}) {
  const token = getToken()
  const headers = new Headers(opcoes.headers || {})

  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  if (opcoes.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const resposta = await fetch(`${API_URL}${caminho}`, { ...opcoes, headers })

  if (resposta.status === 401 && token) {
    limparToken()
    window.dispatchEvent(new Event(EVENTO_SESSAO_EXPIRADA))
  }

  return resposta
}

export function aoExpirarSessao(callback) {
  window.addEventListener(EVENTO_SESSAO_EXPIRADA, callback)
  return () => window.removeEventListener(EVENTO_SESSAO_EXPIRADA, callback)
}

// Helper pra chamadas que esperam JSON e devem lançar erro com a mensagem
// da API (dados.error) em caso de falha — mesmo padrão que criarRua/criarColaborador
// já usavam, agora centralizado pra toda escrita seguir por aqui.
export async function apiFetchJson(caminho, opcoes = {}) {
  const resposta = await apiFetch(caminho, opcoes)
  const dados = await resposta.json().catch(() => ({}))
  if (!resposta.ok) {
    throw new Error(dados.error || `Não foi possível completar a operação (HTTP ${resposta.status})`)
  }
  return dados
}
