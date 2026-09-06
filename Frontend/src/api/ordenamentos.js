import { apiFetchJson } from './client.js'

export function buscarOrdenamentoAtivo() {
  return apiFetchJson('/ordenamentos/ativo')
}

export function criarOrdenamento() {
  return apiFetchJson('/ordenamentos', { method: 'POST' })
}

export function adicionarObjeto(ordenamentoId, entrada, origem = 'manual') {
  return apiFetchJson(`/ordenamentos/${ordenamentoId}/objetos`, {
    method: 'POST',
    body: JSON.stringify({ entrada, origem }),
  })
}

export function excluirObjeto(ordenamentoId, objetoId) {
  return apiFetchJson(`/ordenamentos/${ordenamentoId}/objetos/${objetoId}`, {
    method: 'DELETE',
  })
}
