import { useEffect, useState } from 'react'
import styles from './OperacaoResumo.module.css'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export default function OperacaoResumo({ versao = 0 }) {
  const [dados, setDados] = useState(null)
  const [erro, setErro] = useState(false)

  useEffect(() => {
    fetch(`${API_URL}/estatisticas/operacao`)
      .then((res) => {
        if (!res.ok) throw new Error(`API respondeu ${res.status}`)
        return res.json()
      })
      .then(setDados)
      .catch(() => setErro(true))
  }, [versao])

  if (erro || !dados) return null

  const itens = [
    { icone: '🗺️', valor: dados.total_distritos, label: 'Distritos' },
    { icone: '👥', valor: dados.total_colaboradores, label: 'Colaboradores' },
    { icone: '🏍️', valor: dados.motorizados_moto, label: 'Motos' },
    { icone: '🚐', valor: dados.motorizados_carro, label: 'Carros' },
    { icone: '🚲', valor: dados.ciclistas, label: 'Ciclistas' },
    { icone: '🏢', valor: dados.internos, label: 'Interno' },
    { icone: '🗂️', valor: dados.administrativos, label: 'Administrativo' },
    { icone: '📦', valor: dados.ott, label: 'OTT' },
    { icone: '✉️', valor: dados.ot, label: 'OT' },
    { icone: '🧭', valor: dados.supervisores, label: 'Supervisor' },
    { icone: '⭐', valor: dados.gerentes, label: 'Gerente' },
  ]

  return (
    <section className={styles.faixa} aria-label="Operação da unidade em números">
      {itens.map((item) => (
        <div key={item.label} className={styles.item}>
          <span className={styles.icone} aria-hidden="true">{item.icone}</span>
          <div>
            <strong className={styles.valor}>{item.valor}</strong>
            <span className={styles.label}>{item.label}</span>
          </div>
        </div>
      ))}
    </section>
  )
}
