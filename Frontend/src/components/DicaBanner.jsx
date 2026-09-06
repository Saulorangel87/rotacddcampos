import styles from './DicaBanner.module.css'
import { IconeInformacao } from './icons/Icons.jsx'

export default function DicaBanner({ texto }) {
  return (
    <aside className={styles.banner} role="note">
      <span aria-hidden="true"><IconeInformacao size={19} /></span>
      <div>
        <strong>Dica</strong>
        <p>{texto}</p>
      </div>
    </aside>
  )
}
