import { apiFetch, apiFetchJson } from './client.js'

/**
 * Busca o plano de redistritamento em andamento (rascunho ou concluído).
 * Retorna null se não houver nenhum plano ativo no momento.
 */
export async function buscarPlanoAtivo() {
  const res = await apiFetch('/redistritamento/ativo')
  if (!res.ok) {
    throw new Error(`Não foi possível buscar o plano ativo (HTTP ${res.status})`)
  }
  const dados = await res.json()
  return dados || null
}

/** Inicia um novo plano de redução, indicando a quantidade final de distritos. */
export async function criarPlano(quantidadeAlvo) {
  return apiFetchJson('/redistritamento/planos', {
    method: 'POST',
    body: JSON.stringify({ quantidade_alvo: quantidadeAlvo }),
  })
}

/** Define o distrito de destino de uma rua órfã dentro do plano. */
export async function reatribuirRua(planoId, planoRuaId, distritoDestino) {
  return apiFetchJson(`/redistritamento/planos/${planoId}/ruas/${planoRuaId}`, {
    method: 'PUT',
    body: JSON.stringify({ distrito_destino: distritoDestino }),
  })
}

/** Marca o plano como concluído — ainda editável, mas pronto pra revisão final. */
export async function concluirPlano(planoId) {
  return apiFetchJson(`/redistritamento/planos/${planoId}/concluir`, {
    method: 'POST',
  })
}

/** Aplica o plano de forma DEFINITIVA — muda o banco de dados de verdade. */
export async function aplicarPlano(planoId) {
  return apiFetchJson(`/redistritamento/planos/${planoId}/aplicar`, {
    method: 'POST',
  })
}

/** Descarta um plano em rascunho/concluído e volta pra tela inicial. */
export async function cancelarPlano(planoId) {
  return apiFetchJson(`/redistritamento/planos/${planoId}`, {
    method: 'DELETE',
  })
}
