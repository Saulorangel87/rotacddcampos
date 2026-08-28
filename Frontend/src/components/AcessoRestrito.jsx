import { useState } from "react";
import LoginModal from "./LoginModal.jsx";
import styles from "./AcessoRestrito.module.css";

// Substitui todo o conteúdo do site (mapa, ruas, colaboradores, Zé Rota...)
// pra quem ainda não fez login. O Header continua visível por cima (logo +
// botão Entrar) — só o corpo fica bloqueado.
export default function AcessoRestrito() {
  const [loginAberto, setLoginAberto] = useState(false);

  return (
    <div className={styles.bloqueado}>
      <div className={styles.cartao}>
        <span className={styles.icone} aria-hidden="true">
          <svg viewBox="0 0 24 24" width="32" height="32">
            <path
              fill="currentColor"
              d="M12 1a5 5 0 0 0-5 5v3H6a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-9a2 2 0 0 0-2-2h-1V6a5 5 0 0 0-5-5Zm-3 8V6a3 3 0 0 1 6 0v3Z"
            />
          </svg>
        </span>
        <h2>Acesso restrito</h2>
        <p>
          Este conteúdo é de uso interno do CDD Campos dos Goytacazes. Faça
          login com sua matrícula pra consultar o mapa, as ruas e as demais
          informações do site.
        </p>
        <button
          type="button"
          className={styles.btnEntrar}
          onClick={() => setLoginAberto(true)}
        >
          Fazer login
        </button>
      </div>

      <LoginModal
        aberto={loginAberto}
        onFechar={() => setLoginAberto(false)}
        onEntrou={() => setLoginAberto(false)}
      />
    </div>
  );
}
