import { useState } from 'react'
import AniversarioBadge from './AniversarioBadge.jsx'
import LoginModal from './LoginModal.jsx'
import TrocarSenhaModal from './TrocarSenhaModal.jsx'
import { useAuth } from '../context/AuthContext.jsx'
import styles from './Header.module.css'

export default function Header() {
  const { sessao, autenticado, sair } = useAuth()
  const [busca, setBusca] = useState('')
  const [loginAberto, setLoginAberto] = useState(false)
  const [trocarSenhaAberto, setTrocarSenhaAberto] = useState(false)

  return (
    <header className={styles.header}>
      <div className={styles.marca}>
        <picture>
          <source srcSet="/images/logocorreios.webp" type="image/webp" />
          <img
            className={styles.logo}
            src="/images/logocorreios.png"
            alt="Correios"
            width="120"
            height="20"
          />
        </picture>
      </div>

      <h1 className={styles.titulo}>Guia de Logística: CDD Campos dos Goytacazes</h1>

      <div className={styles.acoes}>
        <label className={styles.busca}>
          <svg viewBox="0 0 24 24" width="15" height="15" aria-hidden="true">
            <circle cx="10" cy="10" r="6.5" fill="none" stroke="currentColor" strokeWidth="2" />
            <line x1="15" y1="15" x2="21" y2="21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
          <input
            type="search"
            placeholder="Pesquisar rua, CEP ou distrito"
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            aria-label="Pesquisar rua, CEP ou distrito"
          />
        </label>

        <AniversarioBadge />

        {autenticado ? (
          <div className={styles.usuarioLogado}>
            <button
              className={styles.usuario}
              type="button"
              onClick={() => setTrocarSenhaAberto(true)}
              title="Trocar senha"
            >
              <span className={styles.avatar} aria-hidden="true">{sessao.matricula.charAt(0)}</span>
              <span className={styles.matriculaLabel}>{sessao.matricula}</span>
              {sessao.papel === 'admin' && <span className={styles.selo}>admin</span>}
            </button>
            <button className={styles.btnSair} type="button" onClick={sair}>
              Sair
            </button>
          </div>
        ) : (
          <button className={styles.usuario} type="button" onClick={() => setLoginAberto(true)}>
            <span className={styles.avatar} aria-hidden="true">?</span>
            Entrar
          </button>
        )}
      </div>

      <LoginModal aberto={loginAberto} onFechar={() => setLoginAberto(false)} onEntrou={() => setLoginAberto(false)} />
      <TrocarSenhaModal
        aberto={trocarSenhaAberto}
        obrigatorio={false}
        onFechar={() => setTrocarSenhaAberto(false)}
        onTrocada={() => setTrocarSenhaAberto(false)}
      />
    </header>
  )
}
