import { useState } from "react";
import Header from "./components/Header.jsx";
import DistrictNav from "./components/DistrictNav.jsx";
import Sidebar from "./components/Sidebar.jsx";
import MapPanel from "./components/MapPanel.jsx";
import AjustesRotasPanel from "./components/AjustesRotasPanel/AjustesRotasPanel.jsx";
import RecentChanges from "./components/RecentChanges.jsx";
import DicaBanner from "./components/DicaBanner.jsx";
import RuasTable from "./components/RuasTable.jsx";
import OperacaoResumo from "./components/OperacaoResumo.jsx";
import CepLookup from "./components/CepLookup.jsx";
import RelatorioMovimentacoes from "./components/RelatorioMovimentacoes.jsx";
import ColaboradoresModal from "./components/ColaboradoresModal.jsx";
import FolgasModal from "./components/FolgasModal.jsx";
import UsuariosModal from "./components/UsuariosModal.jsx";
import LoginModal from "./components/LoginModal.jsx";
import TrocarSenhaModal from "./components/TrocarSenhaModal.jsx";
import Footer from "./components/Footer.jsx";
import ZeRotaChat from "./components/ZeRotaChat.jsx";
import AcessoRestrito from "./components/AcessoRestrito.jsx";
import { listarRuas } from "./api/ruas.js";
import { useAuth } from "./context/AuthContext.jsx";
import styles from "./App.module.css";

export default function App() {
  const { sessao, autenticado, admin } = useAuth();
  const [distritoAtivo, setDistritoAtivo] = useState("");
  const [secaoAtiva, setSecaoAtiva] = useState("mapa");
  const [painelAjustesAberto, setPainelAjustesAberto] = useState(false);
  const [colaboradoresAberto, setColaboradoresAberto] = useState(false);
  const [folgasAberto, setFolgasAberto] = useState(false);
  const [usuariosAberto, setUsuariosAberto] = useState(false);
  const [loginAberto, setLoginAberto] = useState(false);
  const [statsVersao, setStatsVersao] = useState(0);
  const [historicoVersao, setHistoricoVersao] = useState(0);
  const [alteracoes, setAlteracoes] = useState([]);
  const [resultadoBusca, setResultadoBusca] = useState(null);

  async function executarBusca(termo) {
    // A busca sempre leva pro Mapa Geral, senão o resultado não teria onde
    // aparecer (se a pessoa estiver em CEP ou Relatórios, por exemplo).
    setSecaoAtiva("mapa");
    setPainelAjustesAberto(false);
    try {
      const ruas = await listarRuas({ nome: termo });
      // Prioriza um match exato de nome (ex: buscou "Alberto Torres" e existe
      // uma rua com esse nome exato); senão usa o primeiro resultado parcial.
      const exata = ruas.find(
        (r) => r.nome_rua?.toLowerCase() === termo.toLowerCase(),
      );
      const encontrada = exata || ruas[0] || null;
      setResultadoBusca({ termo, rua: encontrada });
    } catch {
      setResultadoBusca({ termo, rua: null });
    }
  }

  function registrarAlteracao(resultado) {
    setAlteracoes((prev) => [
      { ...resultado, quando: "Agora", por: sessao?.matricula || "—" },
      ...prev,
    ]);
    setHistoricoVersao((v) => v + 1);
    setPainelAjustesAberto(false);
  }

  function selecionarSidebar(id) {
    if (id === "ajustes") {
      // Mover rua de distrito é escrita — exige admin. Sem isso, o painel
      // abre normalmente mas toda tentativa de salvar toma 401/403 no backend.
      if (!admin) {
        setLoginAberto(true);
        return;
      }
      setPainelAjustesAberto(true);
      return;
    }
    if (id === "folgas") {
      // A Sidebar só aparece pra quem já está logado (ver AcessoRestrito),
      // então aqui já é sempre um usuário autenticado.
      setFolgasAberto(true);
      return;
    }
    if (id === "colaboradores") {
      // Ver colaboradores exige qualquer usuário autenticado.
      if (!autenticado) {
        setLoginAberto(true);
        return;
      }
      setColaboradoresAberto(true);
      return;
    }
    if (id === "usuarios") {
      // Gerenciar usuários é admin-only — o item nem aparece na sidebar pra
      // quem não é admin, mas a checagem fica aqui também por segurança.
      if (!admin) {
        setLoginAberto(true);
        return;
      }
      setUsuariosAberto(true);
      return;
    }
    if (id === "ruas") {
      // Já existe a lista completa de ruas na tela principal — em vez de duplicar,
      // "Ruas" leva direto pra ela e foca a busca.
      setSecaoAtiva("mapa");
      setPainelAjustesAberto(false);
      requestAnimationFrame(() => {
        const alvo = document.getElementById("tabela-ruas");
        alvo?.scrollIntoView({ behavior: "smooth", block: "start" });
        alvo?.querySelector('input[type="search"]')?.focus();
      });
      return;
    }

    setSecaoAtiva(id);
    setPainelAjustesAberto(false);
  }

  return (
    <div className={styles.app}>
      <Header onBuscar={executarBusca} />

      {!autenticado ? (
        <AcessoRestrito />
      ) : (
        <>
          <DistrictNav
            distritoAtivo={distritoAtivo}
            onSelecionar={setDistritoAtivo}
          />

          <div className={styles.corpo}>
            <Sidebar
              ativo={painelAjustesAberto ? "ajustes" : secaoAtiva}
              onSelecionar={selecionarSidebar}
              admin={admin}
            />

            <main className={styles.principal}>
              {secaoAtiva === "mapa" && (
                <>
                  <div className={styles.grade}>
                    <MapPanel
                      distritoAtivo={distritoAtivo}
                      onSelecionarDistrito={setDistritoAtivo}
                      onAbrirAjustes={() => selecionarSidebar("ajustes")}
                      versao={historicoVersao}
                      resultadoBusca={resultadoBusca}
                      onLimparBusca={() => setResultadoBusca(null)}
                    />
                  </div>

                  {painelAjustesAberto && admin && (
                    <AjustesRotasPanel
                      distritoOrigem={distritoAtivo}
                      onFechar={() => setPainelAjustesAberto(false)}
                      onConcluido={registrarAlteracao}
                    />
                  )}

                  <OperacaoResumo versao={statsVersao} />

                  <div className={styles.grade2}>
                    <RecentChanges alteracoes={alteracoes} />
                    <DicaBanner texto="As alterações são registradas no histórico e podem ser acompanhadas por relatórios." />
                  </div>

                  <div id="tabela-ruas">
                    <RuasTable versao={historicoVersao} />
                  </div>
                </>
              )}

              {secaoAtiva === "cep" && <CepLookup />}

              {secaoAtiva === "relatorios" && (
                <RelatorioMovimentacoes versao={historicoVersao} />
              )}
            </main>
          </div>

          <ZeRotaChat />
        </>
      )}

      <Footer />

      <FolgasModal
        aberto={folgasAberto}
        onFechar={() => setFolgasAberto(false)}
        onAlterado={() => setHistoricoVersao((v) => v + 1)}
      />

      {autenticado && (
        <ColaboradoresModal
          aberto={colaboradoresAberto}
          onFechar={() => setColaboradoresAberto(false)}
          onAlterado={() => setStatsVersao((v) => v + 1)}
        />
      )}

      {admin && (
        <UsuariosModal
          aberto={usuariosAberto}
          onFechar={() => setUsuariosAberto(false)}
        />
      )}

      <LoginModal
        aberto={loginAberto}
        onFechar={() => setLoginAberto(false)}
        onEntrou={() => setLoginAberto(false)}
      />

      {/* Senha provisória (primeiro login ou reset pelo admin) — bloqueia o resto até trocar */}
      {autenticado && sessao.senhaProvisoria && (
        <TrocarSenhaModal
          aberto
          obrigatorio
          onFechar={() => {}}
          onTrocada={() => {}}
        />
      )}
    </div>
  );
}
