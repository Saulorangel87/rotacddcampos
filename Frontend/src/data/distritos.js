// Metadados dos distritos do CDD Campos dos Goytacazes.
// 601-609 mantêm as cores originais da legenda do site antigo (identidade já
// conhecida pelos carteiros). 610-624 foram adicionados depois, então ganharam
// cores novas, escolhidas pra não repetir nenhuma das 9 primeiras nem entre si.
const CORES_POR_DISTRITO = {
  601: '#cc0000', // vermelho (original)
  602: '#ccb300', // amarelo-ouro (original)
  603: '#0000cc', // azul (original)
  604: '#007700', // verde (original)
  605: '#ff8000', // laranja (original)
  606: '#222222', // preto (original)
  607: '#8B4513', // marrom (original)
  608: '#cc00cc', // magenta (original)
  609: '#009999', // ciano (original)
  610: '#4B0082', // índigo
  611: '#FF1493', // rosa-choque
  612: '#2E8B57', // verde-mar
  613: '#DAA520', // dourado
  614: '#4169E1', // azul-royal
  615: '#A0522D', // marrom-terracota
  616: '#DC143C', // carmesim
  617: '#20B2AA', // verde-água
  618: '#9932CC', // roxo
  619: '#FFA07A', // salmão
  620: '#556B2F', // verde-oliva escuro
  621: '#C71585', // vinho-rosado
  622: '#708090', // cinza-ardósia
  623: '#B8860B', // âmbar escuro
  624: '#1E90FF', // azul-dodger
}

export const DISTRITOS = Object.entries(CORES_POR_DISTRITO).map(([numero, cor]) => ({
  numero,
  cor,
  nome: `Distrito ${numero}`,
}))

export function corDoDistrito(numero) {
  return CORES_POR_DISTRITO[numero] || 'var(--cinza-400)'
}

// Layout em grade automática pro mini-mapa esquemático (SVG). Não é geográfico
// — é só uma disposição em blocos, gerada pra caber qualquer quantidade de
// distritos sem precisar redesenhar posições à mão a cada novo distrito criado.
const COLUNAS = 6
const LARGURA_BLOCO = 88
const ALTURA_BLOCO = 78
const ESPACO = 8

export const LAYOUT_DISTRITOS = Object.fromEntries(
  DISTRITOS.map(({ numero }, i) => {
    const col = i % COLUNAS
    const lin = Math.floor(i / COLUNAS)
    return [
      numero,
      {
        x: 10 + col * (LARGURA_BLOCO + ESPACO),
        y: 10 + lin * (ALTURA_BLOCO + ESPACO),
        w: LARGURA_BLOCO,
        h: ALTURA_BLOCO,
      },
    ]
  })
)

export const LARGURA_TOTAL = COLUNAS * (LARGURA_BLOCO + ESPACO) + 10
export const ALTURA_TOTAL = Math.ceil(DISTRITOS.length / COLUNAS) * (ALTURA_BLOCO + ESPACO) + 10
