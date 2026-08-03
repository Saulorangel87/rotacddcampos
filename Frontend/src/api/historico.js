import { apiFetch } from './client.js'

/**
 * Histórico persistido de mudanças de distrito (GET /historico — exige login,
 * qualquer papel). Paginado pelo backend. Lança erro com mensagem amigável
 * se não estiver autenticado, pra tela decidir o que mostrar.
 */
export async function listarHistorico({ pagina = 1, limite = 10 } = {}) {
  const params = new URLSearchParams({ pagina: String(pagina), limite: String(limite) })
  const res = await apiFetch(`/historico?${params.toString()}`)
  if (!res.ok) {
    throw new Error(res.status === 401 ? 'Faça login pra ver o histórico de movimentações.' : `API respondeu ${res.status}`)
  }
  return await res.json()
}
