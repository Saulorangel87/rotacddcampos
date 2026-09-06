export function ordenarParadas(paradas) {
  return [...paradas].sort((a, b) => (a.ordem_final ?? a.ordem_sugerida) - (b.ordem_final ?? b.ordem_sugerida))
}

// IDs das paradas mudam a cada geração; a comparação usa o cadastro agrupado
// e exige todas as ruas, sem adivinhar posições de ruas ausentes ou novas.
export function restaurarSequencia(paradas, historico) {
  const itens = historico.paradas ?? []
  if (!paradas.length || paradas.length !== itens.length) return null
  const porChave = new Map(paradas.map((parada) => [parada.chave_agrupamento, parada]))
  const chaves = new Set(itens.map((item) => item.chave_agrupamento))
  const ordenados = [...itens].sort((a, b) => a.ordem_final - b.ordem_final)
  if (porChave.size !== itens.length || chaves.size !== itens.length
    || ordenados.some((item, indice) => item.ordem_final !== indice + 1 || !porChave.has(item.chave_agrupamento))) return null
  return ordenados.map((item) => porChave.get(item.chave_agrupamento))
}
