import { useState } from 'react'
import LeafletMap from './LeafletMap.jsx'
import { DISTRITOS } from '../data/distritos.js'
import styles from './MapPanel.module.css'

export default function MapPanel({ distritoAtivo, onSelecionarDistrito, onAbrirAjustes }) {
  const [mostrarRuasReais, setMostrarRuasReais] = useState(false)

  return (
    <section className={styles.painel} aria-label="Mapa de distritos">
      <header className={styles.cabecalho}>
        <div className={styles.tituloBox}>
          <span className={styles.icone} aria-hidden="true">🗺️</span>
          <h2>Distritamento CDD Campos</h2>
        </div>
        <div className={styles.acoesCabecalho}>
          <button
            type="button"
            className={styles.btnCamada}
            data-ativo={mostrarRuasReais}
            onClick={() => setMostrarRuasReais((v) => !v)}
            title="Mostra o traçado real das ruas (OpenStreetMap) por cima do mapa de distritos"
          >
            <span aria-hidden="true">🛣️</span> Ruas reais (OSM)
          </button>
          <button type="button" className={styles.btnAjustes} onClick={onAbrirAjustes}>
            <span aria-hidden="true">🛠️</span> Ajustes de Rotas
          </button>
        </div>
      </header>

      <div className={styles.corpo}>
        <LeafletMap
          distritoAtivo={distritoAtivo}
          onSelecionarDistrito={onSelecionarDistrito}
          mostrarRuasReais={mostrarRuasReais}
        />

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
        Mapa real (OpenStreetMap) com os contornos desenhados por Saulo Rangel — ainda em ajuste fino, alguns limites podem não estar 100% precisos.
        {mostrarRuasReais && (distritoAtivo
          ? ` Mostrando as ruas reais do distrito ${distritoAtivo}.`
          : ' Selecione um distrito acima pra ver as ruas reais dele.')}
      </p>
    </section>
  )
}
