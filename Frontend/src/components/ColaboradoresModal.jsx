import { useEffect, useState } from 'react'
import { listarColaboradores } from '../api/colaboradores.js'
import styles from './ColaboradoresModal.module.css'

export default function ColaboradoresModal({ aberto, onFechar }) {
  const [colaboradores, setColaboradores] = useState([])
  const [carregando, setCarregando] = useState(true)
  const [busca, setBusca] = useState('')

  useEffect(() => {
    if (!aberto) return
    setCarregando(true)
    listarColaboradores().then((dados) => {
      setColaboradores(dados)
      setCarregando(false)
    })
  }, [aberto])

  if (!aberto) return null

  const filtrados = colaboradores.filter((c) => c.nome.toLowerCase().includes(busca.toLowerCase()))

  return (
    <div className={styles.fundo} onClick={onFechar}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Colaboradores da unidade">
        <header className={styles.cabecalho}>
          <h2>Colaboradores</h2>
          <button type="button" className={styles.fechar} onClick={onFechar} aria-label="Fechar">✕</button>
        </header>

        <input
          type="search"
          className={styles.campoBusca}
          placeholder="Pesquisar colaborador..."
          value={busca}
          onChange={(e) => setBusca(e.target.value)}
        />

        {carregando ? (
          <p className={styles.mensagem}>Carregando…</p>
        ) : filtrados.length === 0 ? (
          <p className={styles.mensagem}>Nenhum colaborador encontrado.</p>
        ) : (
          <ul className={styles.lista}>
            {filtrados.map((c) => (
              <li key={c.id}>
                <span className={styles.nome}>{c.nome}</span>
                <span className={styles.matricula}>{c.matricula}</span>
              </li>
            ))}
          </ul>
        )}

        <footer className={styles.rodape}>{filtrados.length} colaborador(es)</footer>
      </div>
    </div>
  )
}
