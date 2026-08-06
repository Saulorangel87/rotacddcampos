import { useState } from "react";
import AniversarioBadge from "./AniversarioBadge.jsx";
import LoginModal from "./LoginModal.jsx";
import TrocarSenhaModal from "./TrocarSenhaModal.jsx";
import { useAuth } from "../context/AuthContext.jsx";
import styles from "./Header.module.css";

export default function Header({ onBuscar }) {
  const { sessao, autenticado, sair } = useAuth();
  const [busca, setBusca] = useState("");
  const [loginAberto, setLoginAberto] = useState(false);
  const [trocarSenhaAberto, setTrocarSenhaAberto] = useState(false);

  function aoSubmeterBusca(e) {
    e.preventDefault();
    const termo = busca.trim();
    if (!termo) return;
    onBuscar?.(termo);
  }

  return (
    <header className={styles.header}>
      <div className={styles.marca}>
        <img
          className={styles.logo}
          src="/images/logocorreios.png"
          alt="Correios"
          height="32"
        />
      </div>

      <h1 className={styles.titulo}>
        Guia de Logística: CDD Campos dos Goytacazes
      </h1>

      <div className={styles.acoes}>
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
        </form>

        <div className={styles.aniversarioGroup}>
          <AniversarioBadge />

          {autenticado ? (
            <div className={styles.usuarioLogado}>
              <button
                className={styles.usuario}
                type="button"
                onClick={() => setTrocarSenhaAberto(true)}
                title="Trocar senha"
              >
                <span className={styles.avatar} aria-hidden="true">
                  <svg
                    viewBox="0 0 24 24"
                    width="18"
                    height="18"
                    aria-hidden="true"
                  >
                    <path
                      fill="currentColor"
                      d="M12 12a4 4 0 1 0-4-4 4 4 0 0 0 4 4Zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4Z"
                    />
                  </svg>
                </span>
                <span className={styles.matriculaLabel}>
                  {sessao.matricula}
                </span>
                {sessao.papel === "admin" && (
                  <span className={styles.selo}>admin</span>
                )}
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
            >
              <span className={styles.avatar} aria-hidden="true">
                ?
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
