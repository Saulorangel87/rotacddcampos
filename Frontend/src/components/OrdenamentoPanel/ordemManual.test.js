import test from 'node:test'
import assert from 'node:assert/strict'
import { ordenarParadas, restaurarSequencia } from './ordemManual.js'

const paradas = [
  { id: 101, chave_agrupamento: 'rua:1', nome_rua: 'Rua igual', ordem_sugerida: 1 },
  { id: 202, chave_agrupamento: 'rua:2', nome_rua: 'Rua igual', ordem_sugerida: 2 },
]
const historico = { paradas: [
  { chave_agrupamento: 'rua:2', ordem_final: 1 },
  { chave_agrupamento: 'rua:1', ordem_final: 2 },
] }

test('restaura pelos cadastros, mantendo IDs atuais e ruas homônimas distintas', () => {
  assert.deepEqual(restaurarSequencia(paradas, historico).map((p) => p.id), [202, 101])
  assert.deepEqual(paradas.map((p) => p.id), [101, 202])
})

test('recusa listas parciais, novas ruas, chaves repetidas e posições inválidas', () => {
  assert.equal(restaurarSequencia(paradas.slice(1), historico), null)
  assert.equal(restaurarSequencia([...paradas, { chave_agrupamento: 'rua:3' }], historico), null)
  assert.equal(restaurarSequencia([{ ...paradas[0], chave_agrupamento: 'rua:3' }, paradas[1]], historico), null)
  assert.equal(restaurarSequencia(paradas, { paradas: [historico.paradas[0], historico.paradas[0]] }), null)
  assert.equal(restaurarSequencia(paradas, { paradas: historico.paradas.map((p) => ({ ...p, ordem_final: 1 })) }), null)
})

test('ordem final tem prioridade sobre sugestão sem alterar a resposta original', () => {
  const salvas = paradas.map((p, indice) => ({ ...p, ordem_final: 2 - indice }))
  assert.deepEqual(ordenarParadas(salvas).map((p) => p.id), [202, 101])
  assert.deepEqual(salvas.map((p) => p.id), [101, 202])
})
