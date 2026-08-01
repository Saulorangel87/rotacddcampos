import { useState } from 'react'
import Header from './components/Header.jsx'
import DistrictNav from './components/DistrictNav.jsx'
import Sidebar from './components/Sidebar.jsx'
import MapPanel from './components/MapPanel.jsx'
import AjustesRotasPanel from './components/AjustesRotasPanel/AjustesRotasPanel.jsx'
import RecentChanges from './components/RecentChanges.jsx'
import DicaBanner from './components/DicaBanner.jsx'
import RuasTable from './components/RuasTable.jsx'
import OperacaoResumo from './components/OperacaoResumo.jsx'
import CepLookup from './components/CepLookup.jsx'
import RelatorioMovimentacoes from './components/RelatorioMovimentacoes.jsx'
import ColaboradoresModal from './components/ColaboradoresModal.jsx'
import styles from './App.module.css'

export default function App() {
  const [distritoAtivo, setDistritoAtivo] = useState('')
  const [secaoAtiva, setSecaoAtiva] = useState('mapa')
  const [painelAjustesAberto, setPainelAjustesAberto] = useState(false)
  const [colaboradoresAberto, setColaboradoresAberto] = useState(false)
  const [statsVersao, setStatsVersao] = useState(0)
  const [alteracoes, setAlteracoes] = useState([])

  function registrarAlteracao(resultado) {
    setAlteracoes((prev) => [
      { ...resultado, quando: 'Agora', por: 'Saulo' },
      ...prev,
    ])
    setPainelAjustesAberto(false)
  }

  function selecionarSidebar(id) {
    if (id === 'ajustes') {
      setPainelAjustesAberto(true)
      return
    }
    if (id === 'colaboradores') {
      setColaboradoresAberto(true)
      return
    }
    if (id === 'ruas') {
      // Já existe a lista completa de ruas na tela principal — em vez de duplicar,
      // "Ruas" leva direto pra ela e foca a busca.
      setSecaoAtiva('mapa')
      setPainelAjustesAberto(false)
      requestAnimationFrame(() => {
        const alvo = document.getElementById('tabela-ruas')
        alvo?.scrollIntoView({ behavior: 'smooth', block: 'start' })
        alvo?.querySelector('input[type="search"]')?.focus()
      })
      return
    }

    setSecaoAtiva(id)
    setPainelAjustesAberto(false)
  }

  return (
    <div className={styles.app}>
      <Header />
      <DistrictNav distritoAtivo={distritoAtivo} onSelecionar={setDistritoAtivo} />

      <div className={styles.corpo}>
        <Sidebar
          ativo={painelAjustesAberto ? 'ajustes' : secaoAtiva}
          onSelecionar={selecionarSidebar}
        />

        <main className={styles.principal}>
          {secaoAtiva === 'mapa' && (
            <>
              <div className={styles.grade}>
                <MapPanel
                  distritoAtivo={distritoAtivo}
                  onSelecionarDistrito={setDistritoAtivo}
                  onAbrirAjustes={() => setPainelAjustesAberto(true)}
                />

                {painelAjustesAberto && (
                  <AjustesRotasPanel
                    distritoOrigem={distritoAtivo}
                    onFechar={() => setPainelAjustesAberto(false)}
                    onConcluido={registrarAlteracao}
                  />
                )}
              </div>

              <OperacaoResumo versao={statsVersao} />

              <div className={styles.grade2}>
                <RecentChanges alteracoes={alteracoes} />
                <DicaBanner texto="As alterações são registradas no histórico e podem ser acompanhadas por relatórios." />
              </div>

              <div id="tabela-ruas">
                <RuasTable />
              </div>
            </>
          )}

          {secaoAtiva === 'cep' && <CepLookup />}

          {secaoAtiva === 'relatorios' && <RelatorioMovimentacoes alteracoes={alteracoes} />}
        </main>
      </div>

      <ColaboradoresModal
        aberto={colaboradoresAberto}
        onFechar={() => setColaboradoresAberto(false)}
        onAlterado={() => setStatsVersao((v) => v + 1)}
      />
    </div>
  )
}
