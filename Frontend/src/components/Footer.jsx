import { useEffect, useState } from 'react'
import { APP_VERSION } from '../release.js'
import styles from './Footer.module.css'

const ANO = new Date().getFullYear()
const CHAVE_PWA_INSTALADO = 'rotas_pwa_instalado'

function ehCelular() {
  const mobileData = navigator.userAgentData?.mobile
  if (typeof mobileData === 'boolean') return mobileData

  return /Android.*Mobile|iPhone|iPod|Windows Phone|IEMobile|Opera Mini/i.test(
    navigator.userAgent,
  )
}

// Ícones desenhados como SVG inline (mesma abordagem do resto do projeto) em vez
// de importar uma biblioteca de ícones inteira só para 3 símbolos — mantém o
// bundle final menor, o que ajuda a performance no PageSpeed Insights.
function IconeLinkedin(props) {
  return (
    <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor" {...props}>
      <path d="M19 3H5C3.9 3 3 3.9 3 5v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zM8.5 18.5H6V9h2.5v9.5zM7.25 7.65c-.8 0-1.45-.65-1.45-1.45S6.45 4.75 7.25 4.75s1.45.65 1.45 1.45-.65 1.45-1.45 1.45zM18.5 18.5h-2.5v-4.75c0-1.13-.02-2.58-1.57-2.58-1.57 0-1.81 1.23-1.81 2.5v4.83H10.5V9h2.4v1.28h.03c.33-.62 1.14-1.27 2.35-1.27 2.52 0 2.99 1.66 2.99 3.82v5.67z" />
    </svg>
  )
}

function IconeGithub(props) {
  return (
    <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor" {...props}>
      <path d="M12 .5C5.73.5.98 5.24.98 11.52c0 4.84 3.14 8.94 7.49 10.39.55.1.75-.24.75-.53 0-.26-.01-1.13-.02-2.05-3.05.66-3.69-1.3-3.69-1.3-.5-1.26-1.22-1.6-1.22-1.6-.99-.68.08-.67.08-.67 1.1.08 1.68 1.13 1.68 1.13.98 1.68 2.57 1.2 3.2.92.1-.71.38-1.2.7-1.48-2.43-.28-4.99-1.22-4.99-5.42 0-1.2.43-2.18 1.13-2.95-.11-.28-.49-1.41.11-2.94 0 0 .92-.3 3.02 1.13a10.5 10.5 0 0 1 5.5 0c2.1-1.43 3.02-1.13 3.02-1.13.6 1.53.22 2.66.11 2.94.7.77 1.13 1.75 1.13 2.95 0 4.21-2.57 5.14-5.01 5.41.39.34.74 1 .74 2.03 0 1.47-.01 2.65-.01 3.01 0 .29.2.64.76.53A11.03 11.03 0 0 0 23.02 11.5C23.02 5.24 18.27.5 12 .5Z" />
    </svg>
  )
}

function IconeEmail(props) {
  return (
    <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor" {...props}>
      <path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4-8 5-8-5V6l8 5 8-5v2z" />
    </svg>
  )
}

function IconeDownload(props) {
  return (
    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <path d="M12 3v12" />
      <path d="m7 10 5 5 5-5" />
      <path d="M5 21h14" />
    </svg>
  )
}

export default function Footer() {
  const [installPrompt, setInstallPrompt] = useState(null)
  const [instalado, setInstalado] = useState(false)
  const [celular, setCelular] = useState(false)

  useEffect(() => {
    const dispositivoMovel = ehCelular()
    setCelular(dispositivoMovel)
    if (!dispositivoMovel) return undefined

    const standalone =
      window.matchMedia('(display-mode: standalone)').matches ||
      Boolean(navigator.standalone)
    let instaladoSalvo = false
    try {
      instaladoSalvo = localStorage.getItem(CHAVE_PWA_INSTALADO) === 'true'
    } catch {
      // O botão continua funcionando mesmo se o navegador bloquear o storage.
    }
    queueMicrotask(() => setInstalado(standalone || instaladoSalvo))

    function capturarPrompt(evento) {
      // O evento só é retido no celular; em desktop o navegador mantém seu
      // comportamento nativo de instalação, sem criar um botão no rodapé.
      evento.preventDefault()
      try {
        localStorage.removeItem(CHAVE_PWA_INSTALADO)
      } catch {
        // O storage é apenas uma lembrança visual, não é requisito da instalação.
      }
      setInstalado(false)
      setInstallPrompt(evento)
    }

    function marcarInstalado() {
      try {
        localStorage.setItem(CHAVE_PWA_INSTALADO, 'true')
      } catch {
        // O estado em memória ainda informa a instalação nesta sessão.
      }
      setInstalado(true)
      setInstallPrompt(null)
    }

    window.addEventListener('beforeinstallprompt', capturarPrompt)
    window.addEventListener('appinstalled', marcarInstalado)
    return () => {
      window.removeEventListener('beforeinstallprompt', capturarPrompt)
      window.removeEventListener('appinstalled', marcarInstalado)
    }
  }, [])

  async function instalarApp() {
    if (!installPrompt) return

    try {
      await installPrompt.prompt()
      const escolha = await installPrompt.userChoice
      if (escolha.outcome === 'accepted') {
        try {
          localStorage.setItem(CHAVE_PWA_INSTALADO, 'true')
        } catch {
          // O evento appinstalled atualiza o estado quando o storage falhar.
        }
        setInstalado(true)
      }
    } catch {
      // O navegador pode encerrar o prompt sem concluir a instalação.
    } finally {
      // Um prompt nativo só pode ser usado uma vez; aguardamos um novo evento
      // do navegador para oferecer o botão novamente.
      setInstallPrompt(null)
    }
  }

  return (
    <footer className={styles.rodape}>
      <span>
        &copy; {ANO} Desenvolvido por Saulo Rangel - v{APP_VERSION}
      </span>
      <div className={styles.icones}>
        {celular && installPrompt && !instalado && (
          <button type="button" className={styles.instalar} onClick={instalarApp}>
            <IconeDownload />
            Instalar app
          </button>
        )}
        {celular && instalado && (
          <span className={styles.instalado}>
            <span className={styles.pontoInstalado} aria-hidden="true" />
            App instalado
          </span>
        )}
        <a
          href="https://www.linkedin.com/in/saulorangel87"
          target="_blank"
          rel="noopener noreferrer"
          aria-label="LinkedIn"
        >
          <IconeLinkedin />
        </a>
        <a
          href="https://github.com/Saulorangel87"
          target="_blank"
          rel="noopener noreferrer"
          aria-label="GitHub"
        >
          <IconeGithub />
        </a>
        <a href="mailto:sauloleonardo1987@gmail.com" aria-label="Email">
          <IconeEmail />
        </a>
      </div>
    </footer>
  )
}
