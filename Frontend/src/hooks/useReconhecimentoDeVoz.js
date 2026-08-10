import { useRef, useState } from 'react'

/**
 * Reconhecimento de voz do navegador (Web Speech API), extraído do
 * CepLookup pra reaproveitar no chat do Zé Rota. `onResultado` recebe o
 * texto transcrito quando o usuário termina de falar.
 */
export function useReconhecimentoDeVoz(onResultado) {
  const [ouvindo, setOuvindo] = useState(false)
  const onResultadoRef = useRef(onResultado)
  onResultadoRef.current = onResultado

  const ouvirVoz = () => {
    const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition

    if (!SpeechRecognition) {
      alert('Seu navegador não suporta reconhecimento de voz.')
      return
    }

    const recognition = new SpeechRecognition()
    recognition.lang = 'pt-BR'
    recognition.continuous = false

    recognition.onstart = () => setOuvindo(true)
    recognition.onend = () => setOuvindo(false)

    recognition.onerror = (event) => {
      console.error('Erro no reconhecimento de voz:', event.error)
      if (event.error !== 'aborted' && event.error !== 'no-speech') {
        alert('Erro no microfone: ' + event.error)
      }
      setOuvindo(false)
    }

    recognition.onresult = (event) => {
      const textoTranscrito = event.results[0][0].transcript
      onResultadoRef.current(textoTranscrito)
    }

    recognition.start()
  }

  return { ouvindo, ouvirVoz }
}
