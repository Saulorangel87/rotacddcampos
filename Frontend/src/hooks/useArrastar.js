import { useEffect, useRef, useState } from 'react'

/**
 * Deixa um elemento posicionado (fixed) arrastável com mouse ou toque.
 * Guarda a posição no localStorage pra lembrar onde a pessoa deixou da
 * última vez.
 *
 * Usa Pointer Capture no próprio elemento (em vez de listeners na
 * window): assim que o pointerdown acontece, os eventos de move/up
 * seguintes são "presos" nesse elemento, não importa pra onde o
 * ponteiro vá. Isso evita o efeito de clique "grudado" que acontecia
 * antes (o navegador tentava iniciar um arraste nativo de imagem por
 * baixo, competindo com a nossa lógica).
 *
 * Também limita a posição pra sempre caber dentro da tela atual — uma
 * posição salva numa tela grande (PC) não empurra mais o personagem
 * pra fora da área visível quando a página abre numa tela menor
 * (celular) ou quando a janela é redimensionada.
 *
 * Parâmetros:
 *  - chaveStorage: chave usada no localStorage
 *  - elementoRef: ref do elemento que tem a posição (left/top) aplicada,
 *    usado só pra medir o tamanho dele e recalcular o limite da tela
 *
 * Retorna:
 *  - posicao: null (usa o CSS padrão, ex: bottom-right) ou {x, y} em pixels
 *  - onPointerDown / onPointerMove / onPointerUp / onPointerCancel:
 *    espalha tudo isso no elemento que inicia o arraste
 *  - foiArrasteReal(): true se o gesto que acabou de terminar moveu de
 *    verdade (mais que um tremor de dedo) — usa isso pra decidir se o
 *    onClick que vem junto deve "contar" ou não. Usa ref por baixo (não
 *    state), porque o onClick dispara logo depois do pointerup e um state
 *    React pode não ter atualizado a tempo ainda.
 */
export function useArrastar(chaveStorage, elementoRef) {
  const [posicao, setPosicao] = useState(() => {
    try {
      const salvo = localStorage.getItem(chaveStorage)
      return salvo ? JSON.parse(salvo) : null
    } catch {
      return null
    }
  })
  const arrastandoRef = useRef(false)
  const moveuRef = useRef(false)
  const inicioRef = useRef(null)

  function limitarNaTela(pos, largura, altura) {
    const margem = 4
    const maxX = Math.max(margem, window.innerWidth - largura - margem)
    const maxY = Math.max(margem, window.innerHeight - altura - margem)
    return {
      x: Math.min(Math.max(pos.x, margem), maxX),
      y: Math.min(Math.max(pos.y, margem), maxY),
    }
  }

  // Reajusta a posição salva sempre que a tela mudar de tamanho (inclui
  // a primeira vez que o componente aparece), pra nunca deixar o
  // personagem fora da área visível.
  useEffect(() => {
    function ajustar() {
      const el = elementoRef?.current
      if (!el) return
      setPosicao((atual) => {
        if (!atual) return atual
        const rect = el.getBoundingClientRect()
        const largura = rect.width || 90
        const altura = rect.height || 90
        return limitarNaTela(atual, largura, altura)
      })
    }
    ajustar()
    window.addEventListener('resize', ajustar)
    return () => window.removeEventListener('resize', ajustar)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [elementoRef])

  function onPointerDown(e) {
    // Evita o navegador iniciar o arraste nativo de imagem (é isso que
    // deixava o clique "grudado" como se o mouse continuasse pressionado).
    e.preventDefault()
    const el = e.currentTarget
    try {
      el.setPointerCapture(e.pointerId)
    } catch {
      // navegadores antigos sem suporte: segue sem capture, ainda funciona
    }
    const alvo = el.getBoundingClientRect()
    inicioRef.current = {
      clientX: e.clientX,
      clientY: e.clientY,
      left: alvo.left,
      top: alvo.top,
      largura: alvo.width,
      altura: alvo.height,
      // Toque tem tremor natural bem maior que mouse — um limite pequeno
      // (4px) fazia até um toque parado (tap) ser lido como arraste real
      // no celular, e aí o clique de abrir o chat era bloqueado.
      limite: e.pointerType === 'touch' ? 10 : 4,
    }
    arrastandoRef.current = true
    moveuRef.current = false
  }

  function onPointerMove(e) {
    if (!arrastandoRef.current || !inicioRef.current) return
    const dx = e.clientX - inicioRef.current.clientX
    const dy = e.clientY - inicioRef.current.clientY
    const passouDoLimite = Math.abs(dx) > inicioRef.current.limite || Math.abs(dy) > inicioRef.current.limite

    // Enquanto o movimento estiver dentro da margem de tremor, não mexe em
    // nada ainda — evita o personagem "escorregando" visualmente num toque
    // que na real é só um tap parado.
    if (!passouDoLimite && !moveuRef.current) return

    moveuRef.current = true
    const bruto = { x: inicioRef.current.left + dx, y: inicioRef.current.top + dy }
    setPosicao(limitarNaTela(bruto, inicioRef.current.largura, inicioRef.current.altura))
  }

  function onPointerUp(e) {
    if (!arrastandoRef.current) return
    arrastandoRef.current = false
    try {
      e.currentTarget.releasePointerCapture(e.pointerId)
    } catch {
      // sem problema
    }
    setPosicao((atual) => {
      if (atual && moveuRef.current) {
        try {
          localStorage.setItem(chaveStorage, JSON.stringify(atual))
        } catch {
          // sem problema, só não persiste
        }
      }
      return atual
    })
    // O clique sintético do botão dispara logo em seguida, no mesmo
    // ciclo — deixa moveuRef.current=true até lá pra esse clique não
    // abrir o chat. Depois disso reseta, senão ele fica "preso" e só o
    // PRÓXIMO clique (um clique normal, sem arrastar) é que acaba
    // abrindo o chat sem querer.
    setTimeout(() => {
      moveuRef.current = false
    }, 0)
  }

  function foiArrasteReal() {
    return moveuRef.current
  }

  function resetarPosicao() {
    setPosicao(null)
    try {
      localStorage.removeItem(chaveStorage)
    } catch {
      // sem problema
    }
  }

  return {
    posicao,
    onPointerDown,
    onPointerMove,
    onPointerUp,
    onPointerCancel: onPointerUp,
    foiArrasteReal,
    resetarPosicao,
  }
}
