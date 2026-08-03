import { apiFetch, apiFetchJson } from './client.js'

/**
 * Lista colaboradores (GET /colaboradores — exige login, qualquer papel).
 * Se não estiver autenticado, a API responde 401 e a lista fica vazia — é
 * esperado; quem chama a tela cuida de mostrar a mensagem certa / abrir login.
 */
export async function listarColaboradores({ nome = '', matricula = '', carteiro = '' } = {}) {
  const params = new URLSearchParams()
  if (nome) params.set('nome', nome)
  if (matricula) params.set('matricula', matricula)
  if (carteiro) params.set('carteiro', carteiro)

  const res = await apiFetch(`/colaboradores?${params.toString()}`)
  if (!res.ok) {
    throw new Error(res.status === 401 ? 'Faça login pra ver os colaboradores.' : `API respondeu ${res.status}`)
  }
  return await res.json()
}

/**
 * Aniversariantes do dia (GET /colaboradores/aniversariantes-hoje) — rota pública.
 * Aceita simular outra data no formato DD/MM (o mesmo que a API espera).
 */
export async function aniversariantesDeHoje(dataSimulada = '') {
  const params = new URLSearchParams()
  if (dataSimulada) params.set('data', dataSimulada)

  try {
    const res = await apiFetch(`/colaboradores/aniversariantes-hoje?${params.toString()}`)
    if (!res.ok) throw new Error(`API respondeu ${res.status}`)
    return await res.json()
  } catch (err) {
    console.warn('[api/colaboradores] falha ao buscar aniversariantes:', err.message)
    return []
  }
}

/**
 * Exclui um colaborador (DELETE /colaboradores/:id — exige admin).
 */
export async function excluirColaborador(id) {
  const res = await apiFetch(`/colaboradores/${id}`, { method: 'DELETE' })
  if (!res.ok) {
    const dados = await res.json().catch(() => ({}))
    throw new Error(dados.error || `Não foi possível excluir o colaborador (HTTP ${res.status})`)
  }
}

/**
 * Cadastra um colaborador novo (POST /colaboradores — exige admin). Nome e
 * matrícula são obrigatórios. Datas no formato DD/MM/AAAA (o mesmo que o
 * backend espera). Lança erro com a mensagem da API em caso de falha.
 */
export async function criarColaborador({ nome, matricula, funcao = '', cargo = '', dataAdmissao = '', dataNascimento = '' }) {
  return apiFetchJson('/colaboradores', {
    method: 'POST',
    body: JSON.stringify({
      nome,
      matricula,
      funcao,
      cargo,
      data_admissao: dataAdmissao,
      data_nascimento: dataNascimento,
    }),
  })
}
