import { apiFetch } from './client.js'

/**
 * Busca os distritos no banco (GET /distritos) e monta uma FeatureCollection
 * pronta pro Leaflet, normalizando a propriedade de identificação pra "name"
 * (cada linha do banco guarda um Feature com properties.codigo, aqui a gente
 * usa o código da própria linha — não depende do conteúdo interno do GeoJSON).
 *
 * Se a API falhar ou a tabela ainda estiver vazia, cai pro arquivo estático
 * em public/geojson/distritos.json (o que já tínhamos antes de existir banco).
 *
 * Usa apiFetch (não fetch direto) porque /distritos exige login desde a
 * restrição de acesso público — sem isso, essa chamada sempre voltava 401 e
 * caía pro arquivo estático em silêncio, nunca refletindo o banco de verdade.
 */
export async function buscarDistritosGeoJSON() {
  try {
    const res = await apiFetch('/distritos')
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
