const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

/**
 * Busca os distritos no banco (GET /distritos) e monta uma FeatureCollection
 * pronta pro Leaflet, normalizando a propriedade de identificação pra "name"
 * (cada linha do banco guarda um Feature com properties.codigo, aqui a gente
 * usa o código da própria linha — não depende do conteúdo interno do GeoJSON).
 *
 * Se a API falhar ou a tabela ainda estiver vazia, cai pro arquivo estático
 * em public/geojson/distritos.json (o que já tínhamos antes de existir banco).
 */
export async function buscarDistritosGeoJSON() {
  try {
    const res = await fetch(`${API_URL}/distritos`)
    if (!res.ok) throw new Error(`API respondeu ${res.status}`)
    const linhas = await res.json()

    if (!Array.isArray(linhas) || linhas.length === 0) {
      throw new Error('tabela distritos vazia')
    }

    const features = linhas
      .filter((linha) => linha.geojson)
      .map((linha) => {
        const feature = JSON.parse(linha.geojson)
        return {
          ...feature,
          properties: { ...feature.properties, name: linha.codigo, cor: linha.cor },
        }
      })

    return { type: 'FeatureCollection', features }
  } catch (err) {
    console.warn('[api/distritos] usando arquivo estático — API/banco indisponível:', err.message)
    const res = await fetch('/geojson/distritos.json')
    return res.json()
  }
}
