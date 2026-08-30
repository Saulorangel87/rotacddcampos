import { useEffect, useState } from 'react'
import { apiFetch } from '../api/client.js'
import styles from './OperacaoResumo.module.css'
import {
  IconeMapa,
  IconeColaboradores,
  IconeMoto,
  IconeCarro,
  IconeBike,
  IconePredio,
  IconePasta,
  IconeCaixa,
  IconeEnvelope,
  IconeBussola,
  IconeEstrela,
} from './icons/Icons.jsx'

export default function OperacaoResumo({ versao = 0 }) {
  const [dados, setDados] = useState(null)
  const [erro, setErro] = useState(false)

  useEffect(() => {
    // /estatisticas/operacao exige login desde a restrição de acesso
    // público — precisa do apiFetch (que já manda o token), não um fetch cru.
    apiFetch('/estatisticas/operacao')
      .then((res) => {
        if (!res.ok) throw new Error(`API respondeu ${res.status}`)
        return res.json()
      })
      .then(setDados)
      .catch(() => setErro(true))
  }, [versao])

  if (erro || !dados) return null

  const itens = [
    { Icone: IconeMapa, valor: dados.total_distritos, label: 'Distritos' },
    { Icone: IconeColaboradores, valor: dados.total_colaboradores, label: 'Colaboradores' },
    { Icone: IconeMoto, valor: dados.motorizados_moto, label: 'Motos' },
    { Icone: IconeCarro, valor: dados.motorizados_carro, label: 'Carros' },
    { Icone: IconeBike, valor: dados.ciclistas, label: 'Ciclistas' },
    { Icone: IconePredio, valor: dados.internos, label: 'Interno' },
    { Icone: IconePasta, valor: dados.administrativos, label: 'Administrativo' },
    { Icone: IconeCaixa, valor: dados.ott, label: 'OTT' },
    { Icone: IconeEnvelope, valor: dados.ot, label: 'OT' },
    { Icone: IconeBussola, valor: dados.supervisores, label: 'Supervisor' },
    { Icone: IconeEstrela, valor: dados.gerentes, label: 'Gerente' },
  ]

  return (
    <section className={styles.faixa} aria-label="Operação da unidade em números">
      {itens.map(({ Icone, valor, label }) => (
        <div key={label} className={styles.item}>
          <span className={styles.icone} aria-hidden="true">
            <Icone size={18} />
          </span>
          <div className={styles.texto}>
            <strong className={styles.valor}>{valor}</strong>
            <span className={styles.label}>{label}</span>
          </div>
        </div>
      ))}
    </section>
  )
}
