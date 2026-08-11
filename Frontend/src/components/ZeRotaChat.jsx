import { useEffect, useRef, useState } from 'react'
import { conversarComZeRota } from '../api/zeRota.js'
import { useReconhecimentoDeVoz } from '../hooks/useReconhecimentoDeVoz.js'
import { useArrastar } from '../hooks/useArrastar.js'
import styles from './ZeRotaChat.module.css'

const MENSAGEM_BOAS_VINDAS = 'Oi, eu sou o Zé Rota! Pergunta onde fica uma rua que eu confiro pra você — pode falar ou escrever.'

export default function ZeRotaChat() {
  const [aberto, setAberto] = useState(false)
  const [balaoVisivel, setBalaoVisivel] = useState(true)
  const [mensagens, setMensagens] = useState([])
  const [texto, setTexto] = useState('')
  const [enviando, setEnviando] = useState(false)
  const fimDaListaRef = useRef(null)
  const campoRef = useRef(null)

  const { ouvindo, ouvirVoz } = useReconhecimentoDeVoz((transcrito) => {
    setTexto(transcrito)
    enviarPergunta(transcrito) // ao terminar de falar, envia direto, sem precisar apertar Enter
  })

  const containerRef = useRef(null)
  const { posicao, onPointerDown, onPointerMove, onPointerUp, foiArrasteReal } = useArrastar(
    'ze_rota_posicao',
    containerRef
  )

  useEffect(() => {
    fimDaListaRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [mensagens, enviando])

  // Garante que o cursor volte pro campo assim que ele deixar de estar
  // ocupado (troca de estado já aplicada no DOM, diferente de chamar
  // focus() no meio da função async de envio).
  useEffect(() => {
    if (!enviando) campoRef.current?.focus()
  }, [enviando])

  function aoClicarPersonagem() {
    if (foiArrasteReal()) return // evita abrir o chat sem querer logo depois de soltar um arraste
    setAberto(true)
  }

  async function enviarPergunta(perguntaBruta) {
    const pergunta = perguntaBruta.trim()
    if (!pergunta || enviando) return

    const historico = [...mensagens, { role: 'user', content: pergunta }]
    setMensagens(historico)
    setTexto('')
    setEnviando(true)

    try {
      const resposta = await conversarComZeRota(historico)
      setMensagens([...historico, { role: 'assistant', content: resposta }])
    } catch (err) {
      setMensagens([...historico, { role: 'assistant', content: '⚠️ ' + err.message }])
    } finally {
      setEnviando(false)
    }
  }

  function enviar(e) {
    e.preventDefault()
    enviarPergunta(texto)
  }

  const estiloPosicao = posicao ? { left: posicao.x, top: posicao.y, right: 'auto', bottom: 'auto' } : undefined

  return (
    <>
      {!aberto && (
        <div
          ref={containerRef}
          className={styles.personagemContainer}
          style={estiloPosicao}
        >
          {balaoVisivel && (
            <div className={styles.balaoBoasVindas}>
              <button
                type="button"
                className={styles.fecharBalao}
                onClick={(e) => {
                  e.stopPropagation()
                  setBalaoVisivel(false)
                }}
                aria-label="Fechar mensagem"
              >
                ✕
              </button>
              {MENSAGEM_BOAS_VINDAS}
            </div>
          )}

          <button
            type="button"
            className={styles.personagemBotao}
            onPointerDown={onPointerDown}
            onPointerMove={onPointerMove}
            onPointerUp={onPointerUp}
            onPointerCancel={onPointerUp}
            onClick={aoClicarPersonagem}
            aria-label="Falar com o Zé Rota (segure e arraste pra mover)"
          >
            <img draggable="false" className={styles.imgCorpo} src="/images/ze-rota-corpo.png" alt="Zé Rota, assistente de rotas" />
          </button>
        </div>
      )}

      {aberto && (
        <div className={styles.painel} role="dialog" aria-label="Chat com o Zé Rota">
          <header className={styles.cabecalho}>
            <img src="/images/ze-rota-avatar.png" alt="" aria-hidden="true" />
            <div>
              <strong>Zé Rota</strong>
              <span>Seu assistente de rotas</span>
            </div>
            <button type="button" className={styles.fechar} onClick={() => setAberto(false)} aria-label="Fechar">
              ✕
            </button>
          </header>

          <div className={styles.corpo}>
            <div className={`${styles.balao} ${styles.balaoZe}`}>{MENSAGEM_BOAS_VINDAS}</div>

            {mensagens.map((m, i) => (
              <div
                key={i}
                className={`${styles.balao} ${m.role === 'user' ? styles.balaoUsuario : styles.balaoZe}`}
              >
                {m.content}
              </div>
            ))}

            {enviando && (
              <div className={`${styles.balao} ${styles.balaoZe} ${styles.digitando}`}>
                <span />
                <span />
                <span />
              </div>
            )}

            <div ref={fimDaListaRef} />
          </div>

          <form className={styles.rodape} onSubmit={enviar}>
            <input
              ref={campoRef}
              type="text"
              className={styles.campo}
              placeholder="Pergunta pro Zé Rota…"
              value={texto}
              onChange={(e) => setTexto(e.target.value)}
              autoFocus
            />
            <button
              type="button"
              className={`${styles.btnMic} ${ouvindo ? styles.ouvindo : ''}`}
              onClick={ouvirVoz}
              title="Falar com o Zé Rota"
              aria-label="Falar com o Zé Rota"
            >
              🎤
            </button>
            <button type="submit" className={styles.btnEnviar} disabled={enviando || !texto.trim()}>
              ➤
            </button>
          </form>
        </div>
      )}
    </>
  )
}
