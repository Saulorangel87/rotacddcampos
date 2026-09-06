const TIPOS_LOGRADOURO = new Set([
  'R', 'RUA', 'AV', 'AVENIDA', 'TV', 'TRAVESSA', 'PCA', 'PRACA',
  'EST', 'ESTRADA', 'RODOVIA', 'ALAMEDA', 'LARGO', 'BOULEVARD',
])

const PALAVRAS_IGNORADAS = new Set(['A', 'AS', 'O', 'OS', 'DA', 'DAS', 'DE', 'DO', 'DOS', 'E'])

export function normalizarTextoBusca(valor) {
  return String(valor ?? '')
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toUpperCase()
    .replace(/[^A-Z0-9]+/g, ' ')
    .trim()
    .replace(/\s+/g, ' ')
}

/**
 * Remove o tipo do logradouro e uniformiza cadastros importados como
 * "NOME, RUA TIPO". Artigos são ignorados apenas quando existe outra palavra,
 * permitindo que nomes curtos como "Rua A" continuem pesquisáveis.
 */
export function normalizarNomeRuaBusca(valor) {
  const partesOriginais = normalizarTextoBusca(valor).split(' ').filter(Boolean)
  if (partesOriginais.length === 0) return ''

  const indiceTipo = partesOriginais.findIndex((parte) => TIPOS_LOGRADOURO.has(parte))
  let partes = partesOriginais
  if (indiceTipo === 0) {
    partes = partesOriginais.slice(1)
  } else if (indiceTipo === partesOriginais.length - 1) {
    partes = partesOriginais.slice(0, -1)
  } else if (indiceTipo > 0) {
    partes = [...partesOriginais.slice(indiceTipo + 1), ...partesOriginais.slice(0, indiceTipo)]
  }

  const relevantes = partes.filter((parte) => !PALAVRAS_IGNORADAS.has(parte))
  return (relevantes.length > 0 ? relevantes : partes).join(' ')
}

export function normalizarCepBusca(valor) {
  return String(valor ?? '').replace(/\D/g, '')
}

function todosOsTermosEstaoPresentes(cadastro, consulta) {
  const palavrasCadastro = cadastro.split(' ').filter(Boolean)
  return consulta.split(' ').filter(Boolean).every((termo) =>
    palavrasCadastro.some((palavra) => palavra === termo || palavra.startsWith(termo)),
  )
}

/** Retorna 0 para o melhor resultado e valores maiores para resultados mais fracos. */
export function pontuacaoRua(rua, termo) {
  const consultaBruta = normalizarTextoBusca(termo)
  const cadastroBruto = normalizarTextoBusca(rua?.nome_rua)
  const consultaCep = normalizarCepBusca(termo)
  if (/^\d{8}$/.test(consultaCep)) {
    const cadastroCep = normalizarCepBusca(rua?.cep)
    if (cadastroCep === consultaCep) return 0
    if (cadastroCep.startsWith(consultaCep)) return 1
    return 99
  }
  const consulta = normalizarNomeRuaBusca(termo)
  const cadastro = normalizarNomeRuaBusca(rua?.nome_rua)
  if (!consulta || !cadastro) return 99
  if (cadastroBruto === consultaBruta || cadastro === consulta) return 0
  if (todosOsTermosEstaoPresentes(cadastro, consulta)) return 1
  if (cadastro.includes(consulta)) return 2
  return 3
}

export function ordenarRuasPorCorrespondencia(ruas, termo) {
  return [...(ruas ?? [])].sort((a, b) => {
    const diferenca = pontuacaoRua(a, termo) - pontuacaoRua(b, termo)
    if (diferenca !== 0) return diferenca
    const nome = String(a?.nome_rua ?? '').localeCompare(String(b?.nome_rua ?? ''), 'pt-BR')
    if (nome !== 0) return nome
    return Number(a?.id ?? 0) - Number(b?.id ?? 0)
  })
}

export function ruaCorrespondeBusca(rua, termo) {
  const consulta = normalizarNomeRuaBusca(termo)
  if (!consulta) return true
  const cadastro = normalizarNomeRuaBusca(rua?.nome_rua)
  return todosOsTermosEstaoPresentes(cadastro, consulta)
    || normalizarTextoBusca(rua?.rota).includes(normalizarTextoBusca(termo))
}
