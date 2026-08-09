import styles from './Sidebar.module.css'
import {
  IconeMapa,
  IconeAjustes,
  IconeRuas,
  IconeCep,
  IconeColaboradores,
  IconeRelatorios,
  IconeUsuarios,
  IconeFolgas,
} from './icons/Icons.jsx'

const ITENS = [
  { id: 'mapa', label: 'Mapa Geral', Icone: IconeMapa },
  { id: 'ajustes', label: 'Ajustes de Rotas', Icone: IconeAjustes },
  { id: 'ruas', label: 'Ruas', Icone: IconeRuas },
  { id: 'cep', label: 'CEP', Icone: IconeCep },
  { id: 'folgas', label: 'Folgas', Icone: IconeFolgas },
  { id: 'colaboradores', label: 'Colaboradores', Icone: IconeColaboradores },
  { id: 'relatorios', label: 'Relatórios', Icone: IconeRelatorios },
  { id: 'usuarios', label: 'Usuários', Icone: IconeUsuarios, soAdmin: true },
]

export default function Sidebar({ ativo, onSelecionar, admin }) {
  const itensVisiveis = ITENS.filter((item) => !item.soAdmin || admin)

  return (
    <aside className={styles.sidebar} aria-label="Navegação principal">
      <nav className={styles.nav}>
        {itensVisiveis.map(({ id, label, Icone }) => (
          <button
            key={id}
            type="button"
            className={styles.item}
            data-ativo={id === ativo}
            onClick={() => onSelecionar(id)}
          >
            <span className={styles.icone} aria-hidden="true">
              <Icone size={19} />
            </span>
            <span className={styles.label}>{label}</span>
          </button>
        ))}
      </nav>

      <div className={styles.marca}>
        <img
          className={styles.marcaLogo}
          src="/images/logocorreios.png?v=2"
          alt=""
          aria-hidden="true"
        />
        <div>
          {/* <strong>Correios</strong> */}
          <span>Entrega que conecta o Brasil!</span>
        </div>
      </div>
    </aside>
  )
}
