import { useEffect, useRef, useState } from 'react'
import { conversarComZeRota } from '../api/zeRota.js'
import { useReconhecimentoDeVoz } from '../hooks/useReconhecimentoDeVoz.js'
import styles from './ZeRotaChat.module.css'

const MENSAGEM_BOAS_VINDAS = 'Oi, eu sou o Zé Rota! Pergunta onde fica uma rua que eu confiro pra você — pode falar ou escrever.'

export default function ZeRotaChat() {
  const [aberto, setAberto] = useState(false)
  const [balaoVisivel, setBalaoVisivel] = useState(true)
  const [mensagens, setMensagens] = useState([])
  const [texto, setTexto] = useState('')
  const [enviando, setEnviando] = useState(false)
  const fimDaListaRef = useRef(null)

  const { ouvindo, ouvirVoz } = useReconhecimentoDeVoz((transcrito) => {
    setTexto(transcrito)
  })

  useEffect(() => {
    fimDaListaRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [mensagens, enviando])

  async function enviar(e) {
    e.preventDefault()
    const pergunta = texto.trim()
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

  return (
    <>
      {!aberto && (
        <div className={styles.personagemContainer}>
          {balaoVisivel && (
            <div className={styles.balaoBoasVindas}>
              <button
                type="button"
                className={styles.fecharBalao}
                onClick={(e) => {
                  e.stopPropagation()
                  setBalaoVisivel(false)
                }}
                aria-label="Fechar"
              >
                ✕
              </button>
              {MENSAGEM_BOAS_VINDAS}
            </div>
          )}
          <button
            type="button"
            className={styles.personagemBotao}
            onClick={() => setAberto(true)}
            aria-label="Falar com o Zé Rota"
          >
            <img src="/images/ze-rota-corpo.png" alt="Zé Rota, assistente de rotas" />
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
              type="text"
              className={styles.campo}
              placeholder="Pergunta pro Zé Rota…"
              value={texto}
              onChange={(e) => setTexto(e.target.value)}
              disabled={enviando}
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
