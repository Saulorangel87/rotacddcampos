import { apiFetch } from './client.js'

/**
 * Manda o histórico completo da conversa (POST /ze-rota/conversar — rota
 * pública). O backend não guarda nada entre chamadas, então o front é quem
 * mantém o estado da conversa e reenvia tudo a cada mensagem nova.
 */
export async function conversarComZeRota(mensagens) {
  const res = await apiFetch('/ze-rota/conversar', {
    method: 'POST',
    body: JSON.stringify({ mensagens }),
  })
  const dados = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(dados.error || `Não consegui falar com o Zé Rota agora (HTTP ${res.status})`)
  }
  return dados.resposta
}
