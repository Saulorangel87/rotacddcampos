import { useState } from 'react'
import { useAuth } from '../context/AuthContext.jsx'
import styles from './LoginModal.module.css'

export default function LoginModal({ aberto, onFechar, onEntrou }) {
  const { entrar } = useAuth()
  const [matricula, setMatricula] = useState('')
  const [senha, setSenha] = useState('')
  const [entrando, setEntrando] = useState(false)
  const [erro, setErro] = useState('')

  if (!aberto) return null

  async function aoSubmeter(e) {
    e.preventDefault()
    setErro('')

    if (!matricula.trim() || !senha) {
      setErro('Informe matrícula e senha.')
      return
    }

    setEntrando(true)
    try {
      const novaSessao = await entrar(matricula.trim(), senha)
      setMatricula('')
      setSenha('')
      onEntrou?.(novaSessao)
    } catch (err) {
      setErro(err.message)
    } finally {
      setEntrando(false)
    }
  }

  function fechar() {
    setErro('')
    onFechar()
  }

  return (
    <div className={styles.fundo} onClick={fechar}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Entrar">
        <header className={styles.cabecalho}>
          <h2>Entrar</h2>
          <button type="button" className={styles.fechar} onClick={fechar} aria-label="Fechar">✕</button>
        </header>

        <p className={styles.aviso}>
          Consulta de ruas, mapa e impressão continuam livres pra qualquer pessoa.
          O login é só pra ver colaboradores ou editar dados.
        </p>

        <form className={styles.form} onSubmit={aoSubmeter}>
          <label className={styles.campo}>
            Matrícula
            <input
              type="text"
              value={matricula}
              onChange={(e) => setMatricula(e.target.value)}
              autoFocus
              autoComplete="username"
            />
          </label>

          <label className={styles.campo}>
            Senha
            <input
              type="password"
              value={senha}
              onChange={(e) => setSenha(e.target.value)}
              autoComplete="current-password"
            />
          </label>

          {erro && <p className={styles.erro}>{erro}</p>}

          <footer className={styles.rodapeForm}>
            <button type="button" className={styles.botaoSecundario} onClick={fechar}>
              Cancelar
            </button>
            <button type="submit" className={styles.botaoPrimario} disabled={entrando}>
              {entrando ? 'Entrando…' : 'Entrar'}
            </button>
          </footer>
        </form>
      </div>
    </div>
  )
}
