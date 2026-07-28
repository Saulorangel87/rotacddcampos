import styles from './Sidebar.module.css'

const ITENS = [
  { id: 'mapa', label: 'Mapa Geral', icone: '🗺️' },
  { id: 'ajustes', label: 'Ajustes de Rotas', icone: '🛠️' },
  { id: 'ruas', label: 'Ruas', icone: '🛣️' },
  { id: 'cep', label: 'CEP', icone: '📮' },
  { id: 'carteiros', label: 'Carteiros', icone: '👤' },
  { id: 'relatorios', label: 'Relatórios', icone: '📊' },
]

export default function Sidebar({ ativo, onSelecionar }) {
  return (
    <aside className={styles.sidebar} aria-label="Navegação principal">
      {ITENS.map((item) => (
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
