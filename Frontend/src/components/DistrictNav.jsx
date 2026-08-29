import { useDistritosAtivos } from '../hooks/useDistritosAtivos.js'
import styles from './DistrictNav.module.css'

export default function DistrictNav({ distritoAtivo, onSelecionar }) {
  const { distritosAtivos } = useDistritosAtivos()

  return (
    <nav className={styles.nav} aria-label="Selecionar distrito">
      <span className={styles.rotulo}>Selecione o distrito</span>
      <ul className={styles.lista}>
        {distritosAtivos.map((d) => (
          <li key={d.numero}>
            <button
              type="button"
              className={styles.item}
              data-ativo={d.numero === distritoAtivo}
              style={{ '--cor-distrito': d.cor }}
              onClick={() => onSelecionar(d.numero)}
              aria-pressed={d.numero === distritoAtivo}
            >
              {d.numero}
            </button>
          </li>
        ))}
      </ul>
    </nav>
  )
}
