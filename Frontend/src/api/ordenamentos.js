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

export function selecionarRua(ordenamentoId, objetoId, ruaId) {
  return apiFetchJson(`/ordenamentos/${ordenamentoId}/objetos/${objetoId}/rua`, {
    method: 'PATCH',
    body: JSON.stringify({ rua_id: ruaId }),
  })
}

export function excluirObjeto(ordenamentoId, objetoId) {
  return apiFetchJson(`/ordenamentos/${ordenamentoId}/objetos/${objetoId}`, {
    method: 'DELETE',
  })
}

export function limparOrdenamento(ordenamentoId) {
  return apiFetchJson(`/ordenamentos/${ordenamentoId}/objetos`, {
    method: 'DELETE',
  })
}

export function gerarOrdem(ordenamentoId, somenteAlgoritmo = false) {
  return apiFetchJson(`/ordenamentos/${ordenamentoId}/gerar-ordem`, {
    method: 'POST',
    body: JSON.stringify({ somente_algoritmo: somenteAlgoritmo }),
  })
}

export function salvarOrdemFinal(ordenamentoId, paradaIds, reutilizar = false) {
  return apiFetchJson(`/ordenamentos/${ordenamentoId}/ordem-final`, {
    method: 'PATCH',
    body: JSON.stringify({ parada_ids: paradaIds, reutilizar }),
  })
}

export function esquecerReferencia(referenciaId) {
  return apiFetchJson(`/ordenamentos/referencias/${referenciaId}`, { method: 'DELETE' })
}
