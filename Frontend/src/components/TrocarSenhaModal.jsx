import { useState } from 'react'
import { trocarSenha } from '../api/auth.js'
import { useAuth } from '../context/AuthContext.jsx'
import styles from './LoginModal.module.css'

// Reaproveita o CSS do LoginModal — mesmo tipo de tela, mesmo visual.
export default function TrocarSenhaModal({ aberto, obrigatorio, onFechar, onTrocada }) {
  const { marcarSenhaTrocada, sessao } = useAuth()
  const [senhaAtual, setSenhaAtual] = useState('')
  const [senhaNova, setSenhaNova] = useState('')
  const [confirmacao, setConfirmacao] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [erro, setErro] = useState('')

  if (!aberto) return null

  async function aoSubmeter(e) {
    e.preventDefault()
    setErro('')

    if (senhaNova.length < 8) {
      setErro('A nova senha precisa ter pelo menos 8 caracteres.')
      return
    }
    if (senhaNova !== confirmacao) {
      setErro('A confirmação não bate com a nova senha.')
      return
    }

    setSalvando(true)
    try {
      await trocarSenha(senhaAtual, senhaNova)
      marcarSenhaTrocada()
      setSenhaAtual('')
      setSenhaNova('')
      setConfirmacao('')
      onTrocada?.()
    } catch (err) {
      setErro(err.message)
    } finally {
      setSalvando(false)
    }
  }

  return (
    <div className={styles.fundo} onClick={obrigatorio ? undefined : onFechar}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Trocar senha">
        <header className={styles.cabecalho}>
          <h2>Trocar senha</h2>
          {!obrigatorio && (
            <button type="button" className={styles.fechar} onClick={onFechar} aria-label="Fechar">✕</button>
          )}
        </header>

        <p className={styles.aviso}>
          {obrigatorio
            ? `Sua senha é provisória, ${sessao?.matricula}. Defina uma senha definitiva pra continuar.`
            : 'Defina uma nova senha de acesso.'}
        </p>

        <form className={styles.form} onSubmit={aoSubmeter}>
          <label className={styles.campo}>
            Senha atual
            <input
              type="password"
              value={senhaAtual}
              onChange={(e) => setSenhaAtual(e.target.value)}
              autoFocus
              autoComplete="current-password"
            />
          </label>

          <label className={styles.campo}>
            Nova senha
            <input
              type="password"
              value={senhaNova}
              onChange={(e) => setSenhaNova(e.target.value)}
              autoComplete="new-password"
            />
          </label>

          <label className={styles.campo}>
            Confirmar nova senha
            <input
              type="password"
              value={confirmacao}
              onChange={(e) => setConfirmacao(e.target.value)}
              autoComplete="new-password"
            />
          </label>

          {erro && <p className={styles.erro}>{erro}</p>}

          <footer className={styles.rodapeForm}>
            {!obrigatorio && (
              <button type="button" className={styles.botaoSecundario} onClick={onFechar}>
                Cancelar
              </button>
            )}
            <button type="submit" className={styles.botaoPrimario} disabled={salvando}>
              {salvando ? 'Salvando…' : 'Salvar nova senha'}
            </button>
          </footer>
        </form>
      </div>
    </div>
  )
}
