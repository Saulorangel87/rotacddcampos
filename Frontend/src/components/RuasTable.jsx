import { useEffect, useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import { listarRuas, excluirRua } from '../api/ruas.js'
import { corDoDistrito } from '../data/distritos.js'
import { useAuth } from '../context/AuthContext.jsx'
import NovaRuaModal from './NovaRuaModal.jsx'
import ConfirmModal from './ConfirmModal.jsx'
import styles from './RuasTable.module.css'

const ABAS = [
  { id: 'todas', label: 'Todas as ruas' },
  { id: 'distrito', label: 'Por distrito' },
  { id: 'carteiro', label: 'Por carteiro' },
  { id: 'cep', label: 'Por CEP' },
]

const ITENS_POR_PAGINA = 30

export default function RuasTable({ versao = 0 }) {
  const { admin } = useAuth()
  const [aba, setAba] = useState('todas')
  const [busca, setBusca] = useState('')
  const [ruas, setRuas] = useState([])
  const [pagina, setPagina] = useState(1)
  const [novaRuaAberta, setNovaRuaAberta] = useState(false)
  const [excluindoId, setExcluindoId] = useState(null)
  const [ruaParaExcluir, setRuaParaExcluir] = useState(null)

  useEffect(() => {
    listarRuas().then(setRuas)
  }, [versao])

  useEffect(() => {
    setPagina(1)
  }, [busca, aba])

  function aoCriarRua(novaRua) {
    setRuas((prev) => [novaRua, ...prev])
  }

  function aoExcluirRua(rua) {
    setRuaParaExcluir(rua)
  }

  async function confirmarExclusaoRua() {
    if (!ruaParaExcluir) return
    setExcluindoId(ruaParaExcluir.id)
    try {
      await excluirRua(ruaParaExcluir.id)
      setRuas((prev) => prev.filter((r) => r.id !== ruaParaExcluir.id))
      setRuaParaExcluir(null)
    } catch (err) {
      alert(err.message)
    } finally {
      setExcluindoId(null)
    }
  }

  const linhas = useMemo(() => {
    const alvo = busca.trim().toLowerCase()
    const filtradas = ruas.filter((r) => {
      if (!alvo) return true
      const bateDistrito = String(r.distrito).toLowerCase() === alvo
      const bateCep = String(r.cep).toLowerCase().startsWith(alvo)
      const bateTexto = [r.nome_rua, r.rota].join(' ').toLowerCase().includes(alvo)
      return bateDistrito || bateCep || bateTexto
    })
    if (aba === 'todas') return filtradas
    const chavePorAba = { distrito: 'distrito', carteiro: 'rota', cep: 'cep' }
    const chave = chavePorAba[aba]
    return [...filtradas].sort((a, b) => String(a[chave]).localeCompare(String(b[chave])))
  }, [ruas, busca, aba])

  const totalPaginas = Math.max(1, Math.ceil(linhas.length / ITENS_POR_PAGINA))
  const paginaAtual = Math.min(pagina, totalPaginas)
  const linhasDaPagina = linhas.slice(
    (paginaAtual - 1) * ITENS_POR_PAGINA,
    paginaAtual * ITENS_POR_PAGINA
  )

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

  function imprimirTabela() {
    window.print()
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

        {admin && (
          <button type="button" className={styles.btnNovaRua} onClick={() => setNovaRuaAberta(true)}>
            + Nova rua
          </button>
        )}

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
          <button type="button" className={styles.btnExportar} onClick={imprimirTabela}>
            🖨️ Imprimir
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
              <th aria-hidden="true"></th>
            </tr>
          </thead>
          <tbody>
            {linhasDaPagina.map((r) => (
              <tr key={r.id}>
                <td>{r.nome_rua}</td>
                <td>{r.cep}</td>
                <td>
                  <span className={styles.chipDistrito} style={{ background: corDoDistrito(r.distrito) }}>
                    {r.distrito}
                  </span>
                </td>
                <td>{r.rota}</td>
                <td className={styles.status}>Atualizado {r.atualizado_em}</td>
                <td>
                  {admin && (
                    <button
                      type="button"
                      className={styles.btnExcluir}
                      onClick={() => aoExcluirRua(r)}
                      disabled={excluindoId === r.id}
                      aria-label={`Excluir ${r.nome_rua}`}
                      title="Excluir rua"
                    >
                      {excluindoId === r.id ? '…' : '🗑️'}
                    </button>
                  )}
                </td>
              </tr>
            ))}
            {linhas.length === 0 && (
              <tr>
                <td colSpan={6} className={styles.vazio}>Nenhuma rua encontrada para essa busca.</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <footer className={styles.paginacao}>
        <span>
          {linhas.length === 0
            ? '0 ruas'
            : `${(paginaAtual - 1) * ITENS_POR_PAGINA + 1}–${Math.min(paginaAtual * ITENS_POR_PAGINA, linhas.length)} de ${linhas.length} ruas`}
        </span>
        <div className={styles.paginacaoBotoes}>
          <button
            type="button"
            disabled={paginaAtual <= 1}
            onClick={() => setPagina((p) => p - 1)}
          >
            ← Anterior
          </button>
          <span className={styles.paginaAtual}>Página {paginaAtual} de {totalPaginas}</span>
          <button
            type="button"
            disabled={paginaAtual >= totalPaginas}
            onClick={() => setPagina((p) => p + 1)}
          >
            Próxima →
          </button>
        </div>
      </footer>

      {createPortal(
        <div id="area-impressao-ruas">
          <h1>Lista de Ruas — CDD Campos dos Goytacazes</h1>
          <p>Gerado em {new Date().toLocaleDateString('pt-BR')} — {linhas.length} rua(s)</p>
          <table>
            <thead>
              <tr>
                <th>Rua</th>
                <th>CEP</th>
                <th>Distrito</th>
                <th>Carteiro</th>
              </tr>
            </thead>
            <tbody>
              {linhas.map((r) => (
                <tr key={r.id}>
                  <td>{r.nome_rua}</td>
                  <td>{r.cep}</td>
                  <td>{r.distrito}</td>
                  <td>{r.rota}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>,
        document.body
      )}

      <NovaRuaModal
        aberto={novaRuaAberta}
        onFechar={() => setNovaRuaAberta(false)}
        onCriada={aoCriarRua}
      />

      <ConfirmModal
        aberto={!!ruaParaExcluir}
        titulo="Excluir rua"
        mensagem={ruaParaExcluir ? `Tem certeza que quer excluir "${ruaParaExcluir.nome_rua}" (${ruaParaExcluir.cep})? Essa ação não pode ser desfeita.` : ''}
        onConfirmar={confirmarExclusaoRua}
        onCancelar={() => setRuaParaExcluir(null)}
        confirmando={excluindoId === ruaParaExcluir?.id}
      />
    </section>
  )
}
