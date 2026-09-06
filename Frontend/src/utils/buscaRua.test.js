import test from 'node:test'
import assert from 'node:assert/strict'
import {
  normalizarNomeRuaBusca,
  normalizarCepBusca,
  ordenarRuasPorCorrespondencia,
  ruaCorrespondeBusca,
} from './buscaRua.js'

test('normaliza acentos, tipos e artigos sem perder nomes curtos', () => {
  assert.equal(normalizarNomeRuaBusca('  Av. Sete de Setembro  '), 'SETE SETEMBRO')
  assert.equal(normalizarNomeRuaBusca('Cecília, Rua Santa'), 'SANTA CECILIA')
  assert.equal(normalizarNomeRuaBusca('Alto da Pitanga, Est.'), 'ALTO PITANGA')
  assert.equal(normalizarNomeRuaBusca('Rua A'), 'A')
})

test('filtra a rua por todos os termos e aceita prefixo de palavra', () => {
  const rua = { id: 1, nome_rua: 'AVENIDA SETE DE SETEMBRO' }
  assert.equal(ruaCorrespondeBusca(rua, 'Av Sete Setembro'), true)
  assert.equal(ruaCorrespondeBusca(rua, 'Sete Outubro'), false)
  assert.equal(ruaCorrespondeBusca(rua, 'Sete Setem'), true)
})

test('prioriza correspondência exata equivalente a resultado parcial', () => {
  const ruas = [
    { id: 1, nome_rua: 'RUA PEDREVAL DA SILVA TAVARES' },
    { id: 2, nome_rua: 'RUA SILVA TAVARES' },
  ]
  const ordenadas = ordenarRuasPorCorrespondencia(ruas, 'Silva Tavares')
  assert.deepEqual(ordenadas.map((rua) => rua.id), [2, 1])
})

test('normaliza CEP com e sem hífen', () => {
  assert.equal(normalizarCepBusca('28010-562'), '28010562')
  assert.equal(normalizarCepBusca('CEP: 28010-562'), '28010562')
})

test('prioriza o CEP exato quando a busca do cabeçalho é numérica', () => {
  const ruas = [
    { id: 1, nome_rua: 'RUA ALFA', cep: '28010-561' },
    { id: 2, nome_rua: 'RUA ZETA', cep: '28010562' },
  ]
  assert.deepEqual(ordenarRuasPorCorrespondencia(ruas, '28010-562').map((rua) => rua.id), [2, 1])
})
