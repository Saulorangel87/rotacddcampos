import styles from './DicaBanner.module.css'

export default function DicaBanner({ texto }) {
  return (
    <aside className={styles.banner} role="note">
      <span aria-hidden="true">ℹ️</span>
      <div>
        <strong>Dica</strong>
        <p>{texto}</p>
      </div>
    </aside>
  )
}
