import { apiFetch } from './client.js'

const ROTULOS_CATEGORIA = {
  acesso: 'Acesso',
  seguranca: 'Segurança',
  numero_irregular: 'Numeração irregular',
  varios_nomes: 'Vários nomes',
  outros: 'Outros',
}

export function rotuloCategoria(categoria) {
  return ROTULOS_CATEGORIA[categoria] || categoria
}

export const CATEGORIAS = Object.keys(ROTULOS_CATEGORIA)

/**
 * Lista as observações de campo de uma rua (GET /ruas/:id/observacoes —
 * pública, mesmo espírito da consulta de rua).
 */
export async function listarObservacoes(ruaId) {
  const res = await apiFetch(`/ruas/${ruaId}/observacoes`)
  if (!res.ok) throw new Error(`Não consegui buscar as observações (HTTP ${res.status})`)
  return await res.json()
}

/**
 * Adiciona uma observação (POST — exige admin).
 */
export async function adicionarObservacao(ruaId, { categoria, texto }) {
  const res = await apiFetch(`/ruas/${ruaId}/observacoes`, {
    method: 'POST',
    body: JSON.stringify({ categoria, texto }),
  })
  const dados = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(dados.error || `Não consegui salvar (HTTP ${res.status})`)
  return dados
}

/**
 * Exclui uma observação (DELETE — exige admin).
 */
export async function excluirObservacao(id) {
  const res = await apiFetch(`/observacoes/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const dados = await res.json().catch(() => ({}))
    throw new Error(dados.error || `Não consegui excluir (HTTP ${res.status})`)
  }
}
