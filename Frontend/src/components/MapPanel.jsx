import LeafletMap from './LeafletMap.jsx'
import { DISTRITOS } from '../data/distritos.js'
import styles from './MapPanel.module.css'

export default function MapPanel({ distritoAtivo, onSelecionarDistrito, onAbrirAjustes }) {
  return (
    <section className={styles.painel} aria-label="Mapa de distritos">
      <header className={styles.cabecalho}>
        <div className={styles.tituloBox}>
          <span className={styles.icone} aria-hidden="true">🗺️</span>
          <h2>Distritamento CDD Campos</h2>
        </div>
        <button type="button" className={styles.btnAjustes} onClick={onAbrirAjustes}>
          <span aria-hidden="true">🛠️</span> Ajustes de Rotas
        </button>
      </header>

      <div className={styles.corpo}>
        <LeafletMap distritoAtivo={distritoAtivo} onSelecionarDistrito={onSelecionarDistrito} />

        <ul className={styles.legenda} aria-label="Legenda de distritos">
          {DISTRITOS.map((d) => (
            <li key={d.numero}>
              <span className={styles.corLegenda} style={{ background: d.cor }} aria-hidden="true" />
              {d.numero}
            </li>
          ))}
        </ul>
      </div>

      <p className={styles.rodape}>
        Mapa real (OpenStreetMap) com os contornos desenhados pelo Saulo — ainda em ajuste fino, alguns limites podem não estar 100% precisos.
      </p>
    </section>
  )
}
