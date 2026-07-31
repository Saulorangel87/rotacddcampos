import { MOCK_RUAS } from '../data/mockRuas.js'

// Aponte para a API Go (Backend/) via .env: VITE_API_URL=http://localhost:8080
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

/**
 * Busca ruas na API. Filtros aceitos hoje pelo backend (Backend/handlers/rua_handler.go):
 * nome, cep, distrito. Se a API não responder (ainda não subiu, CORS, etc.),
 * cai para os dados de exemplo para não travar o desenvolvimento da tela.
 */
export async function listarRuas({ nome = '', cep = '', distrito = '' } = {}) {
  const params = new URLSearchParams()
  if (nome) params.set('nome', nome)
  if (cep) params.set('cep', cep)
  if (distrito) params.set('distrito', distrito)

  try {
    const res = await fetch(`${API_URL}/ruas?${params.toString()}`)
    if (!res.ok) throw new Error(`API respondeu ${res.status}`)
    return await res.json()
  } catch (err) {
    console.warn('[api/ruas] usando dados de exemplo — API indisponível:', err.message)
    return filtrarMock({ nome, cep, distrito })
  }
}

/**
 * Atualiza o distrito de uma rua (PUT /ruas/:id — já existe no backend).
 */
export async function atualizarDistritoDaRua(id, novoDistrito) {
  try {
    const res = await fetch(`${API_URL}/ruas/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ distrito: novoDistrito }),
    })
    if (!res.ok) throw new Error(`API respondeu ${res.status}`)
    return await res.json()
  } catch (err) {
    console.warn('[api/ruas] PUT falhou, aplicando só localmente:', err.message)
    return null
  }
}

/**
 * Move várias ruas de uma vez para outro distrito.
 * OBS: o backend ainda não tem um endpoint de lote nem tabela de histórico —
 * isso faz PUT individual por rua (services/rua_service.go + repositories/rua_repository.go
 * já dão suporte a update por id). Quando o endpoint POST /ruas/mover-lote existir,
 * troque esta função para chamá-lo direto e devolver o registro de histórico do servidor.
 */
export async function moverRuasEmLote(ids, novoDistrito) {
  const resultados = await Promise.all(ids.map((id) => atualizarDistritoDaRua(id, novoDistrito)))
  return resultados
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
 * Cadastra uma rua nova (POST /ruas). nome_rua, cep e distrito são obrigatórios no backend.
 * Lança erro com a mensagem da API em caso de falha de validação, pra exibir na tela.
 */
export async function criarRua({ nomeRua, bairro = '', cep, distrito, rota = '', observacao = '' }) {
  const res = await fetch(`${API_URL}/ruas`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      nome_rua: nomeRua,
      bairro,
      cep,
      distrito,
      rota,
      observacao,
    }),
  })
  const dados = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(dados.error || `Não foi possível criar a rua (HTTP ${res.status})`)
  }
  return dados
}

/**
 * Exclui uma rua (DELETE /ruas/:id) — usado em caso de cadastro duplicado.
 */
export async function excluirRua(id) {
  const res = await fetch(`${API_URL}/ruas/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const dados = await res.json().catch(() => ({}))
    throw new Error(dados.error || `Não foi possível excluir a rua (HTTP ${res.status})`)
  }
}
