import { useEffect } from 'react'
import { APP_VERSION, UPDATE_NOTES } from '../release.js'
import styles from './NovidadesModal.module.css'

export default function NovidadesModal({ aberto, onFechar }) {
  useEffect(() => {
    if (!aberto) return undefined

    function aoPressionarTecla(evento) {
      if (evento.key === 'Escape') onFechar()
    }

    document.addEventListener('keydown', aoPressionarTecla)
    return () => document.removeEventListener('keydown', aoPressionarTecla)
  }, [aberto, onFechar])

  if (!aberto) return null

  return (
    <div className={styles.fundo} onClick={onFechar} role="presentation">
      <section
        className={styles.modal}
        onClick={(evento) => evento.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="titulo-novidades"
      >
        <header className={styles.cabecalho}>
          <div>
            <span className={styles.icone} aria-hidden="true">✦</span>
            <span className={styles.rotulo}>NOVIDADES · V{APP_VERSION}</span>
            <h2 id="titulo-novidades">O site ganhou melhorias</h2>
          </div>
          <button
            type="button"
            className={styles.fechar}
            onClick={onFechar}
            aria-label="Fechar novidades"
          >
            ✕
          </button>
        </header>

        <p className={styles.introducao}>
          Confira as novidades para facilitar o lançamento e o ordenamento das encomendas.
        </p>

        <ul className={styles.lista}>
          {UPDATE_NOTES.map((novidade) => (
            <li key={novidade.title} className={styles.item}>
              <span className={styles.marcador} aria-hidden="true">✓</span>
              <div>
                <strong>{novidade.title}</strong>
                <p>{novidade.description}</p>
              </div>
            </li>
          ))}
        </ul>

        <footer className={styles.rodape}>
          <button type="button" className={styles.botao} onClick={onFechar}>
            Entendi, continuar
          </button>
        </footer>
      </section>
    </div>
  )
}
