import { useEffect, useState } from 'react'
import { aniversariantesDeHoje } from '../api/colaboradores.js'
import styles from './AniversarioBadge.module.css'

export default function AniversarioBadge() {
  const [aniversariantes, setAniversariantes] = useState([])
  const [carregado, setCarregado] = useState(false)
  const [aberto, setAberto] = useState(false)

  useEffect(() => {
    aniversariantesDeHoje().then((dados) => {
      setAniversariantes(dados)
      setCarregado(true)
    })
  }, [])

  const temAniversariante = aniversariantes.length > 0

  return (
    <div
      className={styles.wrapper}
      onMouseEnter={() => setAberto(true)}
      onMouseLeave={() => setAberto(false)}
    >
      <button
        type="button"
        className={styles.botao}
        onClick={() => setAberto((v) => !v)}
        aria-label="Aniversariantes do dia"
        aria-expanded={aberto}
      >
        <span aria-hidden="true">🎂</span>
        {temAniversariante && <span className={styles.badge}>{aniversariantes.length}</span>}
      </button>

      {aberto && carregado && (
        <div className={styles.popover} role="dialog" aria-label="Aniversariantes do dia">
          {!temAniversariante ? (
            <p className={styles.vazio}>Nenhum aniversariante hoje.</p>
          ) : (
            <ul className={styles.lista}>
              {aniversariantes.map((c) => (
                <li key={c.id}>
                  <span className={styles.icone} aria-hidden="true">🎉</span>
                  <div>
                    <strong>Parabéns, {c.nome}! A equipe CDD Campos te deseja um feliz aniversário!</strong>
                    {/* {c.funcao && <span className={styles.funcao}>{c.funcao}</span>} */}
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
