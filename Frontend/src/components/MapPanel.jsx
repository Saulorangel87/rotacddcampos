import { useState } from 'react'
import LeafletMap from './LeafletMap.jsx'
import { useDistritosAtivos } from '../hooks/useDistritosAtivos.js'
import { IconeAjustes, IconeMapa, IconeRuas } from './icons/Icons.jsx'
import styles from './MapPanel.module.css'

export default function MapPanel({ distritoAtivo, onSelecionarDistrito, onAbrirAjustes, admin = false, versao = 0, resultadoBusca = null, onLimparBusca }) {
  const [mostrarRuasReais, setMostrarRuasReais] = useState(false)
  const { distritosAtivos } = useDistritosAtivos()

  return (
    <section className={styles.painel} aria-label="Mapa de distritos">
      <header className={styles.cabecalho}>
        <div className={styles.tituloBox}>
          <span className={styles.icone} aria-hidden="true"><IconeMapa size={20} /></span>
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
            <IconeRuas size={17} aria-hidden="true" /> Ruas reais (OSM)
          </button>
          {admin && (
            <button type="button" className={styles.btnAjustes} onClick={onAbrirAjustes}>
              <IconeAjustes size={17} aria-hidden="true" /> Ajustes de Rotas
            </button>
          )}
        </div>
      </header>

      <div className={styles.corpo}>
        <LeafletMap
          distritoAtivo={distritoAtivo}
          onSelecionarDistrito={onSelecionarDistrito}
          mostrarRuasReais={mostrarRuasReais}
          versao={versao}
          resultadoBusca={resultadoBusca}
          onLimparBusca={onLimparBusca}
        />

        <ul className={styles.legenda} aria-label="Legenda de distritos">
          {distritosAtivos.map((d) => (
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
