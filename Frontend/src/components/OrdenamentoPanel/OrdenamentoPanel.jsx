import { useEffect, useRef, useState } from 'react'
import {
  adicionarObjeto,
  buscarOrdenamentoAtivo,
  criarOrdenamento,
  excluirObjeto,
} from '../../api/ordenamentos.js'
import styles from './OrdenamentoPanel.module.css'

export default function OrdenamentoPanel() {
  const [ordenamento, setOrdenamento] = useState(null)
  const [carregando, setCarregando] = useState(true)
  const [criando, setCriando] = useState(false)
  const [digitando, setDigitando] = useState(false)
  const [entrada, setEntrada] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [excluindoId, setExcluindoId] = useState(null)
  const [erro, setErro] = useState('')
  const campoEntradaRef = useRef(null)

  useEffect(() => {
    buscarOrdenamentoAtivo()
      .then(setOrdenamento)
      .catch((e) => setErro(e.message))
      .finally(() => setCarregando(false))
  }, [])

  useEffect(() => {
    if (digitando) campoEntradaRef.current?.focus()
  }, [digitando])

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

  async function cadastrarObjeto(evento) {
    evento.preventDefault()
    const texto = entrada.trim()
    if (!texto || !ordenamento) return

    setSalvando(true)
    setErro('')
    try {
      setOrdenamento(await adicionarObjeto(ordenamento.id, texto))
      setEntrada('')
      requestAnimationFrame(() => campoEntradaRef.current?.focus())
    } catch (e) {
      setErro(e.message)
    } finally {
      setSalvando(false)
    }
  }

  async function removerObjeto(objetoId) {
    if (!ordenamento || excluindoId) return
    setExcluindoId(objetoId)
    setErro('')
    try {
      setOrdenamento(await excluirObjeto(ordenamento.id, objetoId))
    } catch (e) {
      setErro(e.message)
    } finally {
      setExcluindoId(null)
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
          <strong>CDD Campos dos Goytacazes</strong>
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
            <div><strong>{ordenamento.total_objetos ?? 0}</strong><span>objetos</span></div>
            <div><strong>{ordenamento.total_ruas ?? 0}</strong><span>ruas</span></div>
            {(ordenamento.total_pendentes ?? 0) > 0 && (
              <div className={styles.contadorPendente}>
                <strong>{ordenamento.total_pendentes}</strong><span>para revisar</span>
              </div>
            )}
            {(ordenamento.total_sem_coordenadas ?? 0) > 0 && (
              <div className={styles.contadorPendente}>
                <strong>{ordenamento.total_sem_coordenadas}</strong><span>sem coordenada</span>
              </div>
            )}
          </div>

          <div className={styles.adicionarCabecalho}>
            <div>
              <span className={styles.rotulo}>Adicionar encomenda</span>
              <p>Informe a rua e, se souber, o número.</p>
            </div>
          </div>

          <div className={styles.modosEntrada} aria-label="Formas de adicionar encomenda">
            <button type="button" disabled title="Disponível em breve">
              <span aria-hidden="true">▣</span> Escanear <small>Em breve</small>
            </button>
            <button type="button" disabled title="Disponível em breve">
              <span aria-hidden="true">●</span> Falar <small>Em breve</small>
            </button>
            <button
              type="button"
              className={digitando ? styles.modoAtivo : ''}
              onClick={() => setDigitando(true)}
            >
              <span aria-hidden="true">⌨</span> Digitar
            </button>
          </div>

          {digitando && (
            <form className={styles.formulario} onSubmit={cadastrarObjeto}>
              <label htmlFor="endereco-encomenda">Rua e número</label>
              <div>
                <input
                  ref={campoEntradaRef}
                  id="endereco-encomenda"
                  value={entrada}
                  onChange={(evento) => setEntrada(evento.target.value)}
                  placeholder="Ex.: Av. Sete de Setembro 342"
                  autoComplete="off"
                  disabled={salvando}
                />
                <button type="submit" disabled={salvando || !entrada.trim()}>
                  {salvando ? 'Adicionando…' : 'Adicionar'}
                </button>
              </div>
              <span>Após adicionar, o campo fica pronto para a próxima encomenda.</span>
            </form>
          )}

          {(ordenamento.objetos?.length ?? 0) > 0 ? (
            <section className={styles.lista} aria-labelledby="titulo-encomendas">
              <div className={styles.listaTopo}>
                <h3 id="titulo-encomendas">Encomendas adicionadas</h3>
                <span>{ordenamento.total_objetos} no total</span>
              </div>
              <ul>
                {ordenamento.objetos.map((objeto) => (
                  <li key={objeto.id}>
                    <div className={styles.objetoTexto}>
                      {objeto.status_resolucao === 'identificado' ? (
                        <>
                          <strong>
                            {objeto.nome_rua}{objeto.numero ? `, ${objeto.numero}` : ''}
                          </strong>
                          <span>
                            Rua identificada{objeto.cep ? ` · CEP ${objeto.cep}` : ''}
                          </span>
                        </>
                      ) : (
                        <>
                          <strong>{objeto.texto_entrada}</strong>
                          <span className={styles.pendente}>
                            {mensagemPendencia(objeto.motivo_pendencia)}
                          </span>
                        </>
                      )}
                    </div>
                    <button
                      type="button"
                      className={styles.remover}
                      onClick={() => removerObjeto(objeto.id)}
                      disabled={excluindoId !== null}
                      aria-label={`Remover encomenda ${objeto.texto_entrada}`}
                    >
                      {excluindoId === objeto.id ? 'Removendo…' : 'Remover'}
                    </button>
                  </li>
                ))}
              </ul>
            </section>
          ) : (
            <p className={styles.aviso}>O ordenamento está pronto para receber as encomendas.</p>
          )}

          {(ordenamento.ruas?.length ?? 0) > 0 && (
            <section className={styles.ruas} aria-labelledby="titulo-ruas">
              <h3 id="titulo-ruas">Ruas identificadas</h3>
              <ul>
                {ordenamento.ruas.map((rua) => (
                  <li key={rua.chave}>
                    <div>
                      <span>{rua.nome_rua}</span>
                      <small className={rua.latitude == null ? styles.coordenadaPendente : ''}>
                        {rotuloCoordenada(rua)}
                      </small>
                    </div>
                    <strong>{rua.quantidade} {rua.quantidade === 1 ? 'objeto' : 'objetos'}</strong>
                  </li>
                ))}
              </ul>
              {ordenamento.ruas.some((rua) => rua.fonte_coordenada === 'nominatim') && (
                <p className={styles.atribuicao}>
                  Coordenadas externas © <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noreferrer">colaboradores do OpenStreetMap</a>
                </p>
              )}
            </section>
          )}
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

function mensagemPendencia(motivo) {
  if (motivo === 'rua_ambigua') return 'Precisa de revisão · Há mais de uma rua possível'
  return 'Precisa de revisão · Rua não encontrada no cadastro'
}

function rotuloCoordenada(rua) {
  if (rua.latitude == null || rua.longitude == null) return 'Coordenada ainda indisponível'
  if (rua.fonte_coordenada === 'nominatim') return 'Coordenada aproximada · OpenStreetMap'
  return 'Coordenada obtida do mapa interno'
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
