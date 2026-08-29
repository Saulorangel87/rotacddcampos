import { useEffect, useMemo, useRef, useState } from 'react'
import { MapContainer, TileLayer, GeoJSON, Marker } from 'react-leaflet'
import 'leaflet/dist/leaflet.css'
import L from 'leaflet'
import { corDoDistrito } from '../data/distritos.js'
import { buscarDistritosGeoJSON } from '../api/distritos.js'
import { listarRuas } from '../api/ruas.js'
import styles from './LeafletMap.module.css'

const CENTRO_CAMPOS = [-21.7649, -41.3211]

function criarIconeAlfinete(codigo, cor) {
  return L.divIcon({
    className: styles.alfineteWrapper,
    html: `<div class="${styles.alfinete}" style="background:${cor}"><span class="${styles.alfineteTexto}">${codigo}</span></div>`,
    iconSize: [0, 0], // o tamanho real vem do CSS do próprio pino, isso só evita offset duplicado
    iconAnchor: [0, 34],
  })
}

export default function LeafletMap({
  distritoAtivo,
  onSelecionarDistrito,
  mostrarRuasReais = false,
  versao = 0,
  resultadoBusca = null,
  onLimparBusca,
}) {
  const [dados, setDados] = useState(null)
  const [centroAtivo, setCentroAtivo] = useState(null)
  const [ruasGeoJSON, setRuasGeoJSON] = useState(null)
  const [ruaDestacada, setRuaDestacada] = useState(null)
  const geoJsonRef = useRef(null)
  const mapRef = useRef(null)

  useEffect(() => {
    buscarDistritosGeoJSON().then(setDados)
  }, [])

  // Depois de mover uma rua de distrito, o cache abaixo fica desatualizado
  // (continua mostrando a rua no distrito antigo). Zerar aqui força o efeito
  // seguinte a buscar de novo.
  useEffect(() => {
    if (versao > 0) setRuasGeoJSON(null)
  }, [versao])

  // Camada opcional com o traçado real das ruas (vindo do OpenStreetMap, via
  // ruas.geometria) — só busca quando a camada é ligada (ou quando o cache
  // foi invalidado por uma movimentação, acima).
  useEffect(() => {
    if (!mostrarRuasReais || ruasGeoJSON) return
    listarRuas().then((ruas) => {
      const features = ruas
        .filter((r) => r.geometria)
        .map((r) => {
          try {
            return { type: 'Feature', properties: { nome: r.nome_rua, distrito: r.distrito }, geometry: JSON.parse(r.geometria) }
          } catch {
            return null
          }
        })
        .filter(Boolean)
      setRuasGeoJSON({ type: 'FeatureCollection', features })
    })
  }, [mostrarRuasReais, ruasGeoJSON])

  // Só mostra as ruas do distrito selecionado — sem distrito ativo, não mostra
  // nenhuma (evita poluir o mapa com 1252 linhas de uma vez só).
  const ruasDoDistrito = useMemo(() => {
    if (!ruasGeoJSON || !distritoAtivo) return null
    const features = ruasGeoJSON.features.filter((f) => f.properties.distrito === distritoAtivo)
    return { type: 'FeatureCollection', features }
  }, [ruasGeoJSON, distritoAtivo])

  function estiloPorDistrito(feature) {
    const codigo = feature.properties.name
    const ativo = codigo === distritoAtivo
    return {
      // Contorno do distrito selecionado fica escuro e bem mais grosso —
      // assim o destaque funciona mesmo em distritos de cor clara.
      color: ativo ? '#14181f' : corDoDistrito(codigo),
      weight: ativo ? 5 : 1.2,
      fillColor: corDoDistrito(codigo),
      fillOpacity: ativo ? 0.55 : 0.22,
    }
  }

  function aoAdicionarFeature(feature, layer) {
    const codigo = feature.properties.name
    layer.bindTooltip(codigo, { sticky: true })
    layer.on({
      click: () => onSelecionarDistrito?.(codigo),
      mouseover: (e) => {
        if (feature.properties.name !== distritoAtivo) {
          e.target.setStyle({ weight: 3, fillOpacity: 0.4 })
        }
      },
      mouseout: () => geoJsonRef.current?.resetStyle(layer),
    })
  }

  // Sempre que o distrito ativo muda: reaplica o destaque, dá zoom até ele
  // e reposiciona o alfinete no centro da área dele.
  useEffect(() => {
    geoJsonRef.current?.setStyle(estiloPorDistrito)

    if (!distritoAtivo || !geoJsonRef.current || !mapRef.current) return

    let bounds = null
    geoJsonRef.current.eachLayer((camada) => {
      if (camada.feature?.properties?.name === distritoAtivo) {
        bounds = bounds ? bounds.extend(camada.getBounds()) : camada.getBounds()
      }
    })

    if (bounds && bounds.isValid()) {
      mapRef.current.flyToBounds(bounds, { padding: [40, 40], duration: 0.6 })
      setCentroAtivo(bounds.getCenter())
    }
  }, [distritoAtivo])

  // Resultado de uma busca por rua (Header): destaca a rua encontrada e dá
  // zoom nela. Se a rua não tiver geometria real cadastrada, cai pro
  // fallback de selecionar o distrito dela. Se não achou nada, só limpa
  // qualquer destaque anterior — o aviso "não encontrada" é mostrado
  // separado, fora do mapa em si.
  useEffect(() => {
    if (!resultadoBusca) {
      setRuaDestacada(null)
      return
    }

    const rua = resultadoBusca.rua
    if (!rua) {
      setRuaDestacada(null)
      return
    }

    if (rua.geometria && mapRef.current) {
      try {
        const geometry = JSON.parse(rua.geometria)
        const feature = { type: 'Feature', properties: { nome: rua.nome_rua }, geometry }
        const camadaTemporaria = L.geoJSON(feature)
        const bounds = camadaTemporaria.getBounds()
        if (bounds.isValid()) {
          mapRef.current.flyToBounds(bounds, { padding: [60, 60], duration: 0.6 })
        }
        setRuaDestacada(feature)
        return
      } catch {
        // geometria inválida — cai pro fallback abaixo
      }
    }

    // Sem geometria real: melhor que nada é levar pelo menos até o distrito
    // onde a rua fica.
    setRuaDestacada(null)
    if (rua.distrito) onSelecionarDistrito?.(rua.distrito)
  }, [resultadoBusca])

  const semResultado = resultadoBusca && !resultadoBusca.rua

  return (
    <div className={styles.mapa}>
      {(ruaDestacada || semResultado) && (
        <div className={styles.avisoBusca} data-erro={semResultado ? 'true' : 'false'}>
          <span>
            {semResultado
              ? `Rua "${resultadoBusca.termo}" não encontrada.`
              : `Mostrando: ${ruaDestacada.properties.nome}`}
          </span>
          <button type="button" onClick={onLimparBusca} aria-label="Fechar aviso">✕</button>
        </div>
      )}

      <MapContainer
        center={CENTRO_CAMPOS}
        zoom={13}
        scrollWheelZoom={true}
        className={styles.container}
        ref={mapRef}
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> colaboradores'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        {dados && (
          <GeoJSON
            ref={geoJsonRef}
            data={dados}
            style={estiloPorDistrito}
            onEachFeature={aoAdicionarFeature}
          />
        )}
        {mostrarRuasReais && ruasDoDistrito && (
          <GeoJSON
            key={distritoAtivo}
            data={ruasDoDistrito}
            style={{ color: '#0b3d91', weight: 3, opacity: 0.9 }}
            // Ruas que só têm um ponto (geocodificadas, sem linha real ainda)
            // viram um círculo, não o ícone padrão do Leaflet — esse ícone
            // depende de uma imagem cujo caminho quebra com o build do Vite.
            pointToLayer={(feature, latlng) =>
              L.circleMarker(latlng, {
                radius: 7,
                color: '#0b3d91',
                weight: 2,
                fillColor: '#0b3d91',
                fillOpacity: 0.6,
              })
            }
            onEachFeature={(feature, layer) => {
              if (feature.properties?.nome) layer.bindTooltip(feature.properties.nome, { sticky: true })
            }}
          />
        )}
        {ruaDestacada && (
          <GeoJSON
            key={`destaque-${ruaDestacada.properties.nome}`}
            data={ruaDestacada}
            style={{ color: '#ff5722', weight: 7, opacity: 0.95 }}
            pointToLayer={(feature, latlng) =>
              L.circleMarker(latlng, {
                radius: 10,
                color: '#ff5722',
                weight: 3,
                fillColor: '#ff5722',
                fillOpacity: 0.5,
              })
            }
          />
        )}
        {centroAtivo && distritoAtivo && (
          <Marker
            position={centroAtivo}
            icon={criarIconeAlfinete(distritoAtivo, corDoDistrito(distritoAtivo))}
            interactive={false}
          />
        )}
      </MapContainer>
    </div>
  )
}
