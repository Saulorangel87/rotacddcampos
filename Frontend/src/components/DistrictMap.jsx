import { LAYOUT_DISTRITOS, LARGURA_TOTAL, ALTURA_TOTAL } from '../data/distritos.js'
import { useDistritosAtivos } from '../hooks/useDistritosAtivos.js'
import styles from './DistrictMap.module.css'

/**
 * Mini-mapa esquemático dos distritos, desenhado em SVG.
 * Não é geograficamente exato — é um diagrama de blocos na mesma disposição
 * relativa do mapa do Google My Maps atual, o suficiente para orientar
 * qual distrito está selecionado/em destaque sem depender de uma chave de API de mapas.
 * Distritos extintos por um redistritamento aplicado somem sozinhos (usa o
 * mesmo hook useDistritosAtivos que filtra pela API) — as posições de layout
 * continuam fixas, só deixa de desenhar quem não está mais ativo.
 *
 * props:
 *  - emFoco: array de números de distrito para destacar (outros ficam esmaecidos)
 *  - seta: { de, para } desenha uma seta entre dois distritos (usado no antes/depois)
 *  - tamanho: 'grande' | 'pequeno'
 */
export default function DistrictMap({ emFoco = [], seta = null, tamanho = 'grande' }) {
  const { distritosAtivos } = useDistritosAtivos()
  const destacar = (numero) => emFoco.length === 0 || emFoco.includes(numero)

  const centro = (numero) => {
    const b = LAYOUT_DISTRITOS[numero]
    return { x: b.x + b.w / 2, y: b.y + b.h / 2 }
  }

  return (
    <svg
      className={styles.svg}
      data-tamanho={tamanho}
      viewBox={`0 0 ${LARGURA_TOTAL} ${ALTURA_TOTAL}`}
      role="img"
      aria-label={`Mapa esquemático dos ${distritosAtivos.length} distritos`}
    >
      {distritosAtivos.map((d) => {
        const box = LAYOUT_DISTRITOS[d.numero]
        const emDestaque = destacar(d.numero)
        return (
          <g key={d.numero} opacity={emDestaque ? 1 : 0.25}>
            <rect
              x={box.x}
              y={box.y}
              width={box.w}
              height={box.h}
              rx="10"
              fill={d.cor}
              fillOpacity={emDestaque ? 0.35 : 0.15}
              stroke={d.cor}
              strokeWidth={emFoco.includes(d.numero) ? 2.5 : 1.25}
            />
            <text
              x={box.x + box.w / 2}
              y={box.y + box.h / 2}
              textAnchor="middle"
              dominantBaseline="middle"
              fontWeight="800"
              fontSize="15"
              fill="var(--cinza-950)"
            >
              {d.numero}
            </text>
          </g>
        )
      })}

      {seta && (
        <>
          <defs>
            <marker id="ponta-seta" markerWidth="10" markerHeight="10" refX="6" refY="3" orient="auto">
              <path d="M0,0 L6,3 L0,6 Z" fill="var(--sucesso)" />
            </marker>
          </defs>
          <line
            x1={centro(seta.de).x}
            y1={centro(seta.de).y}
            x2={centro(seta.para).x}
            y2={centro(seta.para).y}
            stroke="var(--sucesso)"
            strokeWidth="3.5"
            strokeDasharray="2 6"
            strokeLinecap="round"
            markerEnd="url(#ponta-seta)"
          />
        </>
      )}
    </svg>
  )
}
