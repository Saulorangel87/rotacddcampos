import { apiFetch, apiFetchJson } from './client.js'

/**
 * Consulta o saldo de folgas por matrícula (GET /folgas/saldo — rota pública,
 * mas exige a matrícula exata; não existe endpoint pra listar todo mundo).
 * Retorna { matricula, nome, saldo, lancamentos[] } ou lança erro com a
 * mensagem da API (ex: "matrícula não encontrada").
 */
export async function consultarSaldoFolgas(matricula) {
  const params = new URLSearchParams({ matricula })
  const res = await apiFetch(`/folgas/saldo?${params.toString()}`)
  const dados = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(dados.error || `Não foi possível consultar (HTTP ${res.status})`)
  }
  return dados
}

/**
 * Lança um crédito ou débito de folga (POST /folgas — exige admin).
 * tipo: 'credito' | 'debito'. dataReferencia no formato DD/MM/AAAA, opcional.
 */
export async function lancarFolga({ matricula, tipo, quantidade, motivo, dataReferencia = '' }) {
  return apiFetchJson('/folgas', {
    method: 'POST',
    body: JSON.stringify({
      matricula,
      tipo,
      quantidade,
      motivo,
      data_referencia: dataReferencia,
    }),
  })
}

/**
 * Exclui um lançamento (DELETE /folgas/:id — exige admin). Usado só pra
 * corrigir um lançamento feito por engano; o extrato correto é sempre lançar
 * um novo registro, nunca editar o que já existe.
 */
export async function excluirLancamentoFolga(id) {
  const res = await apiFetch(`/folgas/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const dados = await res.json().catch(() => ({}))
    throw new Error(dados.error || `Não foi possível excluir (HTTP ${res.status})`)
  }
}
