import { MOCK_RUAS } from '../data/mockRuas.js'
import { apiFetch, apiFetchJson } from './client.js'

/**
 * Busca ruas na API. Filtros aceitos hoje pelo backend (Backend/handlers/rua_handler.go):
 * nome, cep, distrito. Se a API não responder (ainda não subiu, CORS, etc.),
 * cai para os dados de exemplo para não travar o desenvolvimento da tela.
 * Essa rota é pública — não precisa de token.
 */
export async function listarRuas({ nome = '', cep = '', distrito = '' } = {}) {
  const params = new URLSearchParams()
  if (nome) params.set('nome', nome)
  if (cep) params.set('cep', cep)
  if (distrito) params.set('distrito', distrito)

  try {
    const res = await apiFetch(`/ruas?${params.toString()}`)
    if (!res.ok) throw new Error(`API respondeu ${res.status}`)
    return await res.json()
  } catch (err) {
    console.warn('[api/ruas] usando dados de exemplo — API indisponível:', err.message)
    return filtrarMock({ nome, cep, distrito })
  }
}

/**
 * Atualiza qualquer combinação de campos de uma rua (PUT /ruas/:id — exige
 * admin). Usada tanto pra corrigir nome/CEP/bairro quanto pra mover distrito.
 */
export async function atualizarRua(id, dados) {
  return apiFetchJson(`/ruas/${id}`, {
    method: 'PUT',
    body: JSON.stringify(dados),
  })
}

/**
 * Atualiza o distrito de uma rua (PUT /ruas/:id — exige admin).
 * IMPORTANTE: antes essa função engolia qualquer erro (inclusive 401/403) e
 * devolvia null, então a tela achava que tinha salvo mesmo quando a API
 * recusou. Agora ela lança o erro de verdade — quem chama precisa tratar.
 */
export async function atualizarDistritoDaRua(id, novoDistrito) {
  return apiFetchJson(`/ruas/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ distrito: novoDistrito }),
  })
}

/**
 * Move várias ruas de uma vez para outro distrito.
 * OBS: o backend ainda não tem um endpoint de lote — isso faz PUT individual
 * por rua. Se qualquer uma falhar (ex: sessão expirou no meio do lote), o
 * Promise.all rejeita e quem chamou precisa tratar — não finge sucesso parcial.
 */
export async function moverRuasEmLote(ids, novoDistrito) {
  return Promise.all(ids.map((id) => atualizarDistritoDaRua(id, novoDistrito)))
}

function filtrarMock({ nome, cep, distrito }) {
  return MOCK_RUAS.filter((r) => {
    const okNome = !nome || r.nome_rua.toLowerCase().includes(nome.toLowerCase())
    const okCep = !cep || r.cep.includes(cep)
    const okDistrito = !distrito || r.distrito === distrito
    return okNome && okCep && okDistrito
  })
}

/**
 * Cadastra uma rua nova (POST /ruas — exige admin). nome_rua, cep e distrito
 * são obrigatórios no backend. Lança erro com a mensagem da API em caso de
 * falha (validação, 401/403 sem login/sem ser admin), pra exibir na tela.
 */
export async function criarRua({ nomeRua, bairro = '', cep, distrito, rota = '', observacao = '' }) {
  return apiFetchJson('/ruas', {
    method: 'POST',
    body: JSON.stringify({
      nome_rua: nomeRua,
      bairro,
      cep,
      distrito,
      rota,
      observacao,
    }),
  })
}

/**
 * Exclui uma rua (DELETE /ruas/:id — exige admin).
 */
export async function excluirRua(id) {
  const res = await apiFetch(`/ruas/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const dados = await res.json().catch(() => ({}))
    throw new Error(dados.error || `Não foi possível excluir a rua (HTTP ${res.status})`)
  }
}
