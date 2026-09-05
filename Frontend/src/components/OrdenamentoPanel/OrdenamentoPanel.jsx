import { useEffect, useState } from 'react'
import { buscarOrdenamentoAtivo, criarOrdenamento } from '../../api/ordenamentos.js'
import styles from './OrdenamentoPanel.module.css'

export default function OrdenamentoPanel() {
  const [ordenamento, setOrdenamento] = useState(null)
  const [carregando, setCarregando] = useState(true)
  const [criando, setCriando] = useState(false)
  const [erro, setErro] = useState('')

  useEffect(() => {
    buscarOrdenamentoAtivo()
      .then(setOrdenamento)
      .catch((e) => setErro(e.message))
      .finally(() => setCarregando(false))
  }, [])

  async function iniciarOrdenamento() {
    setCriando(true)
    setErro('')
    try {
      setOrdenamento(await criarOrdenamento())
    } catch (e) {
      if (e.message.includes('já existe')) {
        try {
          setOrdenamento(await buscarOrdenamentoAtivo())
          return
        } catch (erroBusca) {
          setErro(erroBusca.message)
          return
        }
      }
      setErro(e.message)
    } finally {
      setCriando(false)
    }
  }

  return (
    <section className={styles.pagina} aria-labelledby="titulo-ordenamento">
      <header className={styles.cabecalho}>
        <div>
          <span className={styles.sobretitulo}>Organização da carga</span>
          <h1 id="titulo-ordenamento">Ordenamento de Entregas</h1>
          <p>Organize as ruas por proximidade antes de sair para entrega.</p>
        </div>
      </header>

      <article className={styles.partida} aria-label="Ponto de partida">
        <div className={styles.marcador} aria-hidden="true">CDD</div>
        <div>
          <span className={styles.rotulo}>Ponto de partida fixo</span>
          <strong>CDD Campos</strong>
          <span>Av. Sete de Setembro, 342</span>
          <span>Campos dos Goytacazes — RJ</span>
        </div>
      </article>

      {carregando ? (
        <div className={styles.estado} role="status">Consultando ordenamento…</div>
      ) : ordenamento ? (
        <article className={styles.andamento}>
          <div className={styles.andamentoTopo}>
            <div>
              <span className={styles.rotulo}>Em andamento</span>
              <h2>Novo ordenamento</h2>
            </div>
            <span className={styles.data}>Iniciado {formatarData(ordenamento.created_at)}</span>
          </div>

          <div className={styles.contadores} aria-label="Resumo do ordenamento">
            <div><strong>0</strong><span>objetos</span></div>
            <div><strong>0</strong><span>ruas</span></div>
          </div>

          <p className={styles.aviso}>O ordenamento está pronto para receber as encomendas.</p>
        </article>
      ) : (
        <article className={styles.estadoInicial}>
          <div>
            <h2>Nenhum ordenamento em andamento</h2>
            <p>Inicie uma nova organização para adicionar as encomendas da sua carga.</p>
          </div>
          <button type="button" onClick={iniciarOrdenamento} disabled={criando}>
            {criando ? 'Iniciando…' : '+ Novo ordenamento'}
          </button>
        </article>
      )}

      {erro && <p className={styles.erro} role="alert">{erro}</p>}
    </section>
  )
}

function formatarData(valor) {
  if (!valor) return 'agora'
  const data = new Date(valor)
  if (Number.isNaN(data.getTime())) return 'agora'
  return data.toLocaleString('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
