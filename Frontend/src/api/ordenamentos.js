import { apiFetchJson } from './client.js'

export function buscarOrdenamentoAtivo() {
  return apiFetchJson('/ordenamentos/ativo')
}

export function criarOrdenamento() {
  return apiFetchJson('/ordenamentos', { method: 'POST' })
}
