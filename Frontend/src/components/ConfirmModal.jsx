import styles from './ConfirmModal.module.css'
import { IconeAlerta } from './icons/Icons.jsx'

export default function ConfirmModal({
  aberto,
  titulo,
  mensagem,
  onConfirmar,
  onCancelar,
  confirmando,
  textoConfirmar = 'Excluir',
  textoConfirmando = 'Excluindo…',
}) {
  if (!aberto) return null

  return (
    <div
      className={styles.fundo}
      onClick={(e) => {
        e.stopPropagation()
        onCancelar()
      }}
    >
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="alertdialog" aria-label={titulo}>
        <div className={styles.icone} aria-hidden="true"><IconeAlerta size={30} /></div>
        <h2 className={styles.titulo}>{titulo}</h2>
        <p className={styles.mensagem}>{mensagem}</p>

        <footer className={styles.rodape}>
          <button type="button" className={styles.botaoSecundario} onClick={onCancelar} disabled={confirmando}>
            Cancelar
          </button>
          <button type="button" className={styles.botaoPerigo} onClick={onConfirmar} disabled={confirmando}>
            {confirmando ? textoConfirmando : textoConfirmar}
          </button>
        </footer>
      </div>
    </div>
  )
}
