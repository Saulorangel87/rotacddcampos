import { useEffect, useState } from 'react'
import { corDoDistrito } from '../data/distritos.js'
import styles from './RelatorioMovimentacoes.module.css'

const ITENS_POR_PAGINA = 10

export default function RelatorioMovimentacoes({ alteracoes }) {
  const [pagina, setPagina] = useState(1)

  useEffect(() => {
    setPagina(1)
  }, [alteracoes.length])

  const totalPaginas = Math.max(1, Math.ceil(alteracoes.length / ITENS_POR_PAGINA))
  const paginaAtual = Math.min(pagina, totalPaginas)
  const itensDaPagina = alteracoes.slice(
    (paginaAtual - 1) * ITENS_POR_PAGINA,
    paginaAtual * ITENS_POR_PAGINA
  )

  return (
    <section className={styles.caixa} aria-label="Relatório de movimentações">
      <h2 className={styles.titulo}>📊 Relatório de movimentações</h2>

      <p className={styles.aviso}>
        Por enquanto isso mostra só as mudanças feitas <strong>nesta sessão do navegador</strong> —
        ainda não é salvo no banco de dados. Movimentação de carteiro entre distritos também não
        existe como funcionalidade ainda. Os dois ficam registrados como próximo passo de backend.
      </p>

      {alteracoes.length === 0 ? (
        <p className={styles.vazio}>Nenhuma movimentação registrada nesta sessão ainda.</p>
      ) : (
        <>
          <ul className={styles.lista}>
            {itensDaPagina.map((a, i) => (
              <li key={i}>
                <span className={styles.ponto} style={{ background: corDoDistrito(a.para) }} aria-hidden="true" />
                <div className={styles.conteudo}>
                  <p>{a.quantidade} rua(s) movidas de {a.de || '—'} para {a.para}</p>
                  <span className={styles.meta}>{a.quando} · Por: {a.por}</span>
                </div>
              </li>
            ))}
          </ul>

          <footer className={styles.paginacao}>
            <button type="button" disabled={paginaAtual <= 1} onClick={() => setPagina((p) => p - 1)}>
              ← Anterior
            </button>
            <span>Página {paginaAtual} de {totalPaginas}</span>
            <button
              type="button"
              disabled={paginaAtual >= totalPaginas}
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
