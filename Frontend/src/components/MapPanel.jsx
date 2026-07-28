import DistrictMap from './DistrictMap.jsx'
import styles from './MapPanel.module.css'

export default function MapPanel({ distritoAtivo, onAbrirAjustes }) {
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
        <DistrictMap emFoco={distritoAtivo ? [distritoAtivo] : []} />

        <ul className={styles.legenda} aria-label="Legenda de distritos">
          {['601', '602', '603', '604', '605', '606', '607', '608', '609'].map((n) => (
            <li key={n}>
              <span className={styles.corLegenda} style={{ background: `var(--d${n})` }} aria-hidden="true" />
              {n}
            </li>
          ))}
        </ul>
      </div>

      <p className={styles.rodape}>
        Dados cartográficos de referência interna — não substitui o Google Maps para navegação em rota.
      </p>
    </section>
  )
}
