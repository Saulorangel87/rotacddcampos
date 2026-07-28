import { useEffect, useMemo, useState } from 'react'
import { listarRuas } from '../api/ruas.js'
import styles from './RuasTable.module.css'

const ABAS = [
  { id: 'todas', label: 'Todas as ruas' },
  { id: 'distrito', label: 'Por distrito' },
  { id: 'carteiro', label: 'Por carteiro' },
  { id: 'cep', label: 'Por CEP' },
]

export default function RuasTable() {
  const [aba, setAba] = useState('todas')
  const [busca, setBusca] = useState('')
  const [ruas, setRuas] = useState([])

  useEffect(() => {
    listarRuas().then(setRuas)
  }, [])

  const linhas = useMemo(() => {
    const filtradas = ruas.filter((r) =>
      [r.nome_rua, r.cep, r.distrito, r.rota].join(' ').toLowerCase().includes(busca.toLowerCase())
    )
    if (aba === 'todas') return filtradas
    const chavePorAba = { distrito: 'distrito', carteiro: 'rota', cep: 'cep' }
    const chave = chavePorAba[aba]
    return [...filtradas].sort((a, b) => String(a[chave]).localeCompare(String(b[chave])))
  }, [ruas, busca, aba])

  function exportarCsv() {
    const cabecalho = ['Rua', 'CEP', 'Distrito', 'Carteiro']
    const corpo = linhas.map((r) => [r.nome_rua, r.cep, r.distrito, r.rota])
    const csv = [cabecalho, ...corpo].map((linha) => linha.join(';')).join('\n')
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'ruas-cdd-campos.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <section className={styles.caixa} aria-label="Tabela de ruas cadastradas">
      <div className={styles.barra}>
        <div className={styles.abas} role="tablist">
          {ABAS.map((a) => (
            <button
              key={a.id}
              role="tab"
              type="button"
              aria-selected={aba === a.id}
              className={styles.aba}
              data-ativa={aba === a.id}
              onClick={() => setAba(a.id)}
            >
              {a.label}
            </button>
          ))}
        </div>

        <div className={styles.busca}>
          <input
            type="search"
            placeholder="Pesquisar..."
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            aria-label="Pesquisar ruas cadastradas"
          />
          <button type="button" className={styles.btnExportar} onClick={exportarCsv}>
            ⬇ Exportar
          </button>
        </div>
      </div>

      <div className={styles.tabelaWrapper}>
        <table className={styles.tabela}>
          <thead>
            <tr>
              <th>Rua</th>
              <th>CEP</th>
              <th>Distrito</th>
              <th>Carteiro</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {linhas.map((r) => (
              <tr key={r.id}>
                <td>{r.nome_rua}</td>
                <td>{r.cep}</td>
                <td>
                  <span className={styles.chipDistrito} style={{ background: `var(--d${r.distrito})` }}>
                    {r.distrito}
                  </span>
                </td>
                <td>{r.rota}</td>
                <td className={styles.status}>Atualizado {r.atualizado_em}</td>
              </tr>
            ))}
            {linhas.length === 0 && (
              <tr>
                <td colSpan={5} className={styles.vazio}>Nenhuma rua encontrada para essa busca.</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </section>
  )
}
