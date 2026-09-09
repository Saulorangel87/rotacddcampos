import { useState } from "react";
import AniversarioBadge from "./AniversarioBadge.jsx";
import LoginModal from "./LoginModal.jsx";
import TrocarSenhaModal from "./TrocarSenhaModal.jsx";
import { useAuth } from "../context/AuthContext.jsx";
import { useReconhecimentoDeVoz } from "../hooks/useReconhecimentoDeVoz.js";
import {
  IconeAdministrador,
  IconeColaborador,
  IconeEntrar,
  IconeMicrofone,
} from "./icons/Icons.jsx";
import styles from "./Header.module.css";

export default function Header({ onBuscar }) {
  const { sessao, autenticado, sair } = useAuth();
  const [busca, setBusca] = useState("");
  const [loginAberto, setLoginAberto] = useState(false);
  const [trocarSenhaAberto, setTrocarSenhaAberto] = useState(false);
  const { ouvindo, ouvirVoz } = useReconhecimentoDeVoz((textoTranscrito) => {
    const termo = textoTranscrito.trim();
    if (!termo) return;
    setBusca(termo);
    onBuscar?.(termo);
  });

  function aoSubmeterBusca(e) {
    e.preventDefault();
    const termo = busca.trim();
    if (!termo) return;
    onBuscar?.(termo);
  }

  function iniciarBuscaPorVoz() {
    document.activeElement?.blur?.();
    ouvirVoz();
  }

  return (
    <header className={styles.header}>
      <div className={styles.marca}>
        <img
          className={styles.logo}
          src="/images/logocorreios.png?v=2"
          alt="Correios"
          height="32"
        />
      </div>

      <h1 className={styles.titulo}>
        Guia de Logística: CDD Campos dos Goytacazes
      </h1>

      <div className={styles.acoes}>
        {/* Busca depende de /ruas, que agora exige login — sem sentido mostrar
            pra quem ainda não entrou */}
        {autenticado && (
          <form className={styles.busca} onSubmit={aoSubmeterBusca}>
            <input
              type="search"
              placeholder="Buscar rua no mapa"
              value={busca}
              onChange={(e) => setBusca(e.target.value)}
              aria-label="Buscar rua no mapa"
            />
            <button
              type="submit"
              className={styles.btnBuscar}
              aria-label="Buscar"
            >
              <svg viewBox="0 0 24 24" width="15" height="15" aria-hidden="true">
                <circle
                  cx="10"
                  cy="10"
                  r="6.5"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                />
                <line
                  x1="15"
                  y1="15"
                  x2="21"
                  y2="21"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                />
              </svg>
            </button>
            <button
              type="button"
              className={`${styles.btnMicrofone} ${ouvindo ? styles.microfoneAtivo : ''}`}
              onClick={iniciarBuscaPorVoz}
              aria-label={ouvindo ? 'Ouvindo busca por voz' : 'Pesquisar por voz'}
              aria-pressed={ouvindo}
              title="Pesquisar por voz"
            >
              <IconeMicrofone size={15} aria-hidden="true" />
            </button>
          </form>
        )}

        <div className={styles.aniversarioGroup}>
          {/* Aniversariante do dia também exige login agora */}
          {autenticado && <AniversarioBadge />}

          {autenticado ? (
            <div className={styles.usuarioLogado}>
              <button
                className={styles.usuario}
                type="button"
                onClick={() => setTrocarSenhaAberto(true)}
                title={`Trocar senha — ${sessao.papel === "admin" ? "administrador" : "colaborador"}`}
              >
                <span
                  className={`${styles.avatar} ${
                    sessao.papel === "admin" ? styles.avatarAdmin : styles.avatarColaborador
                  }`}
                  aria-hidden="true"
                >
                  {sessao.papel === "admin" ? <IconeAdministrador size={18} /> : <IconeColaborador size={18} />}
                </span>
                <span className={styles.matriculaLabel}>
                  {sessao.matricula}
                </span>
                <span
                  className={`${styles.selo} ${
                    sessao.papel === "admin" ? styles.seloAdmin : styles.seloColaborador
                  }`}
                  title={sessao.papel === "admin" ? "Administrador" : "Colaborador"}
                >
                  {sessao.papel === "admin" ? "ADMIN" : "COLAB."}
                </span>
              </button>
              <button className={styles.btnSair} type="button" onClick={sair}>
                Sair
              </button>
            </div>
          ) : (
            <button
              className={styles.usuario}
              type="button"
              onClick={() => setLoginAberto(true)}
              title="Entrar no sistema"
            >
              <span className={styles.avatar} aria-hidden="true">
                <IconeEntrar size={18} />
              </span>
              Entrar
            </button>
          )}
        </div>
      </div>

      <LoginModal
        aberto={loginAberto}
        onFechar={() => setLoginAberto(false)}
        onEntrou={() => setLoginAberto(false)}
      />
      <TrocarSenhaModal
        aberto={trocarSenhaAberto}
        obrigatorio={false}
        onFechar={() => setTrocarSenhaAberto(false)}
        onTrocada={() => setTrocarSenhaAberto(false)}
      />
    </header>
  );
}
