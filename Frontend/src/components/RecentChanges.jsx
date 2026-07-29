import { corDoDistrito } from '../data/distritos.js'
import styles from './RecentChanges.module.css'

export default function RecentChanges({ alteracoes }) {
  return (
    <section className={styles.caixa} aria-label="Alterações recentes">
      <h3>Alterações recentes</h3>
      {alteracoes.length === 0 ? (
        <p className={styles.vazio}>Nenhuma alteração registrada ainda nesta sessão.</p>
      ) : (
        <ul className={styles.lista}>
          {alteracoes.map((a, i) => (
            <li key={i}>
              <span className={styles.ponto} style={{ background: corDoDistrito(a.para) }} aria-hidden="true" />
              <div>
                <p>{a.quantidade} rua(s) movidas de {a.de || '—'} para {a.para}</p>
                <span className={styles.meta}>{a.quando} · Por: {a.por}</span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
