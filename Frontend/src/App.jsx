import { useState } from 'react'
import Header from './components/Header.jsx'
import DistrictNav from './components/DistrictNav.jsx'
import Sidebar from './components/Sidebar.jsx'
import MapPanel from './components/MapPanel.jsx'
import AjustesRotasPanel from './components/AjustesRotasPanel/AjustesRotasPanel.jsx'
import RecentChanges from './components/RecentChanges.jsx'
import DicaBanner from './components/DicaBanner.jsx'
import RuasTable from './components/RuasTable.jsx'
import styles from './App.module.css'

export default function App() {
  const [distritoAtivo, setDistritoAtivo] = useState('602')
  const [secaoAtiva, setSecaoAtiva] = useState('mapa')
  const [painelAjustesAberto, setPainelAjustesAberto] = useState(false)
  const [alteracoes, setAlteracoes] = useState([])

  function registrarAlteracao(resultado) {
    setAlteracoes((prev) => [
      { ...resultado, quando: 'Agora', por: 'Saulo' },
      ...prev,
    ])
    setPainelAjustesAberto(false)
  }

  return (
    <div className={styles.app}>
      <Header />
      <DistrictNav distritoAtivo={distritoAtivo} onSelecionar={setDistritoAtivo} />

      <div className={styles.corpo}>
        <Sidebar
          ativo={painelAjustesAberto ? 'ajustes' : secaoAtiva}
          onSelecionar={(id) => {
            if (id === 'ajustes') {
              setPainelAjustesAberto(true)
            } else {
              setSecaoAtiva(id)
              setPainelAjustesAberto(false)
            }
          }}
        />

        <main className={styles.principal}>
          <div className={styles.grade}>
            <MapPanel distritoAtivo={distritoAtivo} onAbrirAjustes={() => setPainelAjustesAberto(true)} />

            {painelAjustesAberto && (
              <AjustesRotasPanel
                distritoOrigem={distritoAtivo}
                onFechar={() => setPainelAjustesAberto(false)}
                onConcluido={registrarAlteracao}
              />
            )}
          </div>

          <div className={styles.grade2}>
            <RecentChanges alteracoes={alteracoes} />
            <DicaBanner texto="As alterações são registradas no histórico e podem ser acompanhadas por relatórios." />
          </div>

          <RuasTable />
        </main>
      </div>
    </div>
  )
}
