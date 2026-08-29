import { useEffect, useState } from 'react'
import { DISTRITOS } from '../data/distritos.js'
import { apiFetchJson } from '../api/client.js'

/**
 * Filtra a lista estática DISTRITOS (que tem cores/nomes fixos de todos os
 * 24 códigos já usados algum dia) pra só os que estão ativos de verdade no
 * banco agora — importante depois de um redistritamento aplicado, senão
 * distrito extinto continua aparecendo em chips, legenda, mini-mapa e nos
 * formulários de criar/editar/mover rua.
 *
 * Se a API falhar por qualquer motivo, cai pra lista estática inteira (mesmo
 * comportamento de antes dessa correção) em vez de quebrar a tela.
 */
export function useDistritosAtivos() {
  const [distritosAtivos, setDistritosAtivos] = useState(DISTRITOS)
  const [carregando, setCarregando] = useState(true)

  useEffect(() => {
    let cancelado = false

    apiFetchJson('/distritos')
      .then((lista) => {
        if (cancelado) return
        const codigosAtivos = new Set(lista.map((d) => d.codigo))
        setDistritosAtivos(DISTRITOS.filter((d) => codigosAtivos.has(d.numero)))
      })
      .catch(() => {
        // API falhou — mantém a lista estática completa como fallback
      })
      .finally(() => {
        if (!cancelado) setCarregando(false)
      })

    return () => {
      cancelado = true
    }
  }, [])

  return { distritosAtivos, carregando }
}
