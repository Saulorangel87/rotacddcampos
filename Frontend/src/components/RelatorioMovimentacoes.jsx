import { useEffect, useState } from 'react'
import { listarHistorico } from '../api/historico.js'
import { corDoDistrito } from '../data/distritos.js'
import { useAuth } from '../context/AuthContext.jsx'
import styles from './RelatorioMovimentacoes.module.css'

const ITENS_POR_PAGINA = 10

export default function RelatorioMovimentacoes({ versao }) {
  const { autenticado } = useAuth()
  const [pagina, setPagina] = useState(1)
  const [dados, setDados] = useState({ itens: [], total: 0, total_paginas: 1 })
  const [carregando, setCarregando] = useState(true)
  const [erro, setErro] = useState('')

  useEffect(() => {
    if (!autenticado) {
      setCarregando(false)
      return
    }
    setCarregando(true)
    setErro('')
    listarHistorico({ pagina, limite: ITENS_POR_PAGINA })
      .then(setDados)
      .catch((err) => setErro(err.message))
      .finally(() => setCarregando(false))
  }, [autenticado, pagina, versao])

  if (!autenticado) {
    return (
      <section className={styles.caixa} aria-label="Relatório de movimentações">
        <h2 className={styles.titulo}>📊 Relatório de movimentações</h2>
        <p className={styles.vazio}>Faça login pra ver o histórico de movimentações entre distritos.</p>
      </section>
    )
  }

  return (
    <section className={styles.caixa} aria-label="Relatório de movimentações">
      <h2 className={styles.titulo}>📊 Relatório de movimentações</h2>

      <p className={styles.aviso}>
        Histórico completo, salvo no banco — não some ao fechar o navegador.
        Movimentação de carteiro entre distritos ainda não existe como funcionalidade.
      </p>

      {carregando ? (
        <p className={styles.vazio}>Carregando…</p>
      ) : erro ? (
        <p className={styles.vazio}>{erro}</p>
      ) : dados.itens.length === 0 ? (
        <p className={styles.vazio}>Nenhuma movimentação registrada ainda.</p>
      ) : (
        <>
          <ul className={styles.lista}>
            {dados.itens.map((a) =>
              a.tipo === 'folga' ? (
                <li key={a.id}>
                  <span className={styles.ponto} style={{ background: 'var(--cor-azul)' }} aria-hidden="true" />
                  <div className={styles.conteudo}>
                    <p>{a.descricao}</p>
                    <span className={styles.meta}>{formatarData(a.data)} · Por: {a.usuario || '—'}</span>
                  </div>
                </li>
              ) : (
                <li key={a.id}>
                  <span className={styles.ponto} style={{ background: corDoDistrito(a.distrito_destino) }} aria-hidden="true" />
                  <div className={styles.conteudo}>
                    <p>{a.nome_rua}: {a.distrito_origem || '—'} → {a.distrito_destino}</p>
                    <span className={styles.meta}>{formatarData(a.data)} · Por: {a.usuario || '—'}</span>
                  </div>
                </li>
              ),
            )}
          </ul>

          <footer className={styles.paginacao}>
            <button type="button" disabled={pagina <= 1} onClick={() => setPagina((p) => p - 1)}>
              ← Anterior
            </button>
            <span>Página {pagina} de {dados.total_paginas}</span>
            <button
              type="button"
              disabled={pagina >= dados.total_paginas}
              onClick={() => setPagina((p) => p + 1)}
            >
              Próxima →
            </button>
          </footer>
        </>
      )}
    </section>
  )
}

function formatarData(iso) {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' })
  } catch {
    return iso
  }
}
