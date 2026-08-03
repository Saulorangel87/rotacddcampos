import styles from './Sidebar.module.css'

const ITENS = [
  { id: 'mapa', label: 'Mapa Geral', icone: '🗺️' },
  { id: 'ajustes', label: 'Ajustes de Rotas', icone: '🛠️' },
  { id: 'ruas', label: 'Ruas', icone: '🛣️' },
  { id: 'cep', label: 'CEP', icone: '📮' },
  { id: 'colaboradores', label: 'Colaboradores', icone: '👤' },
  { id: 'relatorios', label: 'Relatórios', icone: '📊' },
  { id: 'usuarios', label: 'Usuários', icone: '🔑', soAdmin: true },
]

export default function Sidebar({ ativo, onSelecionar, admin }) {
  const itensVisiveis = ITENS.filter((item) => !item.soAdmin || admin)

  return (
    <aside className={styles.sidebar} aria-label="Navegação principal">
      {itensVisiveis.map((item) => (
        <button
          key={item.id}
          type="button"
          className={styles.item}
          data-ativo={item.id === ativo}
          onClick={() => onSelecionar(item.id)}
        >
          <span className={styles.icone} aria-hidden="true">{item.icone}</span>
          <span className={styles.label}>{item.label}</span>
        </button>
      ))}
    </aside>
  )
}
