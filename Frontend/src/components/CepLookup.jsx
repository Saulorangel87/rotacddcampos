import { useEffect, useState } from 'react'
import { listarRuas } from '../api/ruas.js'
import { corDoDistrito } from '../data/distritos.js'
import styles from './CepLookup.module.css'

export default function CepLookup() {
  const [termo, setTermo] = useState('')
  const [resultados, setResultados] = useState([])
  const [buscando, setBuscando] = useState(false)
  const [jaBuscou, setJaBuscou] = useState(false)
  const [ouvindo, setOuvindo] = useState(false)

  useEffect(() => {
    if (termo.trim().length < 3) {
      setResultados([])
      setJaBuscou(false)
      return
    }
    setBuscando(true)
    const timer = setTimeout(() => {
      listarRuas({ nome: termo }).then((dados) => {
        setResultados(dados)
        setBuscando(false)
        setJaBuscou(true)
      })
    }, 400)
    return () => clearTimeout(timer)
  }, [termo])

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
      // Sem isso, qualquer falha (permissão bloqueada por política do
      // servidor, sem microfone, sem internet, etc.) acontecia calada —
      // não aparecia nem no console. Alertar aqui é temporário, só pra
      // diagnosticar em produção; depois trocamos por algo mais discreto.
      console.error('Erro no reconhecimento de voz:', event.error)
      alert('Erro no microfone: ' + event.error)
      setOuvindo(false)
    }

    recognition.onresult = (event) => {
      const textoTranscrito = event.results[0][0].transcript
      setTermo(textoTranscrito)
    }

    recognition.start()
  }

  return (
    <section className={styles.caixa} aria-label="Consulta de CEP por rua">
      <h2 className={styles.titulo}>📮 Consultar CEP por rua</h2>
      <p className={styles.dica}>Digite pelo menos 3 letras do nome da rua.</p>

      {/* Wrapper do Input */}
      <div className={styles.inputWrapper}>
        <input
          type="search"
          className={styles.campo}
          placeholder="Ex: Alberto Torres"
          value={termo}
          onChange={(e) => setTermo(e.target.value)}
          autoFocus
        />
        
        {/* Botão com o seu SVG inline */}
        <button
          type="button"
          onClick={ouvirVoz}
          className={`${styles.micBtn} ${ouvindo ? styles.ouvindo : ''}`}
          title="Falar nome da rua"
        >
          <svg
            className={styles.micIcon}
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 384 512"
          >
            <path d="M96 96c0-53 43-96 96-96 50.3 0 91.6 38.7 95.7 88L232 88c-13.3 0-24 10.7-24 24s10.7 24 24 24l56 0 0 48-56 0c-13.3 0-24 10.7-24 24s10.7 24 24 24l55.7 0c-4.1 49.3-45.3 88-95.7 88-53 0-96-43-96-96L96 96zM24 160c13.3 0 24 10.7 24 24l0 40c0 79.5 64.5 144 144 144s144-64.5 144-144l0-40c0-13.3 10.7-24 24-24s24 10.7 24 24l0 40c0 97.9-73.3 178.7-168 190.5l0 49.5 48 0c13.3 0 24 10.7 24 24s-10.7 24-24 24l-144 0c-13.3 0-24-10.7-24-24s10.7-24 24-24l48 0 0-49.5C73.3 402.7 0 321.9 0 224l0-40c0-13.3 10.7-24 24-24z" />
          </svg>
        </button>
      </div>

      {buscando && <p className={styles.mensagem}>Buscando…</p>}

      {!buscando && jaBuscou && resultados.length === 0 && (
        <p className={styles.mensagem}>Nenhuma rua encontrada com esse nome.</p>
      )}

      {!buscando && resultados.length > 0 && (
        <ul className={styles.lista}>
          {resultados.map((r) => (
            <li key={r.id}>
              <div>
                <strong>{r.nome_rua}</strong>
                <span className={styles.bairro}>{r.bairro}</span>
              </div>
              <div className={styles.direita}>
                <span className={styles.cep}>{r.cep}</span>
                <span className={styles.chipDistrito} style={{ background: corDoDistrito(r.distrito) }}>
                  {r.distrito}
                </span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}