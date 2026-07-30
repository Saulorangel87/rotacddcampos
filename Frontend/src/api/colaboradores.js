const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

/**
 * Lista colaboradores (GET /colaboradores). Filtros opcionais: nome, matricula, carteiro.
 */
export async function listarColaboradores({ nome = '', matricula = '', carteiro = '' } = {}) {
  const params = new URLSearchParams()
  if (nome) params.set('nome', nome)
  if (matricula) params.set('matricula', matricula)
  if (carteiro) params.set('carteiro', carteiro)

  try {
    const res = await fetch(`${API_URL}/colaboradores?${params.toString()}`)
    if (!res.ok) throw new Error(`API respondeu ${res.status}`)
    return await res.json()
  } catch (err) {
    console.warn('[api/colaboradores] falha ao buscar colaboradores:', err.message)
    return []
  }
}

/**
 * Aniversariantes do dia (GET /colaboradores/aniversariantes-hoje).
 * Aceita simular outra data no formato DD/MM (o mesmo que a API espera).
 */
export async function aniversariantesDeHoje(dataSimulada = '') {
  const params = new URLSearchParams()
  if (dataSimulada) params.set('data', dataSimulada)

  try {
    const res = await fetch(`${API_URL}/colaboradores/aniversariantes-hoje?${params.toString()}`)
    if (!res.ok) throw new Error(`API respondeu ${res.status}`)
    return await res.json()
  } catch (err) {
    console.warn('[api/colaboradores] falha ao buscar aniversariantes:', err.message)
    return []
  }
}
