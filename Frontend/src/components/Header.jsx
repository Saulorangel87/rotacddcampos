import { useState } from 'react'
import AniversarioBadge from './AniversarioBadge.jsx'
import styles from './Header.module.css'

export default function Header({ usuario = 'Saulo' }) {
  const [busca, setBusca] = useState('')

  return (
    <header className={styles.header}>
      <div className={styles.marca}>
        <picture>
          <source srcSet="/images/logocorreios.webp" type="image/webp" />
          <img className={styles.logo} src="/images/logocorreios.png" alt="Correios" />
        </picture>
      </div>

      <h1 className={styles.titulo}>Guia de Logística: CDD Campos dos Goytacazes</h1>

      <div className={styles.acoes}>
        <label className={styles.busca}>
          <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
            <circle cx="10" cy="10" r="6.5" fill="none" stroke="currentColor" strokeWidth="2" />
            <line x1="15" y1="15" x2="21" y2="21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
          <input
            type="search"
            placeholder="Pesquisar rua, CEP ou distrito"
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            aria-label="Pesquisar rua, CEP ou distrito"
          />
        </label>

        <AniversarioBadge />

        <button className={styles.usuario} type="button">
          <span className={styles.avatar} aria-hidden="true">{usuario.charAt(0)}</span>
          {usuario}
        </button>
      </div>
    </header>
  )
}
