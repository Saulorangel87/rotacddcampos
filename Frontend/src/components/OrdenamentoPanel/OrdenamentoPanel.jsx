import { useEffect, useRef, useState } from 'react'
import {
  adicionarObjeto,
  buscarOrdenamentoAtivo,
  criarOrdenamento,
  excluirObjeto,
  gerarOrdem,
  limparOrdenamento,
  selecionarRua,
} from '../../api/ordenamentos.js'
import { useReconhecimentoDeVoz } from '../../hooks/useReconhecimentoDeVoz.js'
import styles from './OrdenamentoPanel.module.css'

export default function OrdenamentoPanel() {
  const [ordenamento, setOrdenamento] = useState(null)
  const [carregando, setCarregando] = useState(true)
  const [criando, setCriando] = useState(false)
  const [digitando, setDigitando] = useState(false)
  const [entrada, setEntrada] = useState('')
  const [origemEntrada, setOrigemEntrada] = useState('manual')
  const [escaneando, setEscaneando] = useState(false)
  const [avisoScanner, setAvisoScanner] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [excluindoId, setExcluindoId] = useState(null)
  const [limpando, setLimpando] = useState(false)
  const [gerando, setGerando] = useState(false)
  const [resolvendoId, setResolvendoId] = useState(null)
  const [erro, setErro] = useState('')
  const campoEntradaRef = useRef(null)
  const videoScannerRef = useRef(null)
  const leitorScannerRef = useRef(null)
  const controlesScannerRef = useRef(null)
  const scannerAtivoRef = useRef(false)
  const { ouvindo, ouvirVoz } = useReconhecimentoDeVoz((textoTranscrito) => {
    setEntrada(textoTranscrito)
    setOrigemEntrada('voz')
    setDigitando(true)
    setErro('')
  })

  function iniciarEntradaPorVoz() {
    encerrarScanner()
    setOrigemEntrada('voz')
    setDigitando(true)
    setAvisoScanner('')
    ouvirVoz()
  }

  useEffect(() => () => encerrarScanner(), [])

  async function iniciarScanner() {
    setErro('')
    setAvisoScanner('')
    setDigitando(false)
    if (!navigator.mediaDevices?.getUserMedia) {
      setErro('A câmera não está disponível neste navegador. Use Digitar ou Falar.')
      return
    }

    try {
      const { BrowserMultiFormatReader } = await import('@zxing/browser')
      leitorScannerRef.current = new BrowserMultiFormatReader()
      scannerAtivoRef.current = true
      setEscaneando(true)
      requestAnimationFrame(iniciarLeituraScanner)
    } catch (e) {
      encerrarScanner()
      setErro(e.name === 'NotAllowedError'
        ? 'Permita o acesso à câmera para escanear a etiqueta.'
        : 'Não foi possível abrir a câmera. Use Digitar ou Falar.')
    }
  }

  async function iniciarLeituraScanner() {
    if (!scannerAtivoRef.current || !videoScannerRef.current || !leitorScannerRef.current) return
    try {
      const controles = await leitorScannerRef.current.decodeFromVideoDevice(
        undefined,
        videoScannerRef.current,
        (resultado) => {
          if (!scannerAtivoRef.current) return
          const codigo = resultado?.getText()?.trim()
          if (!codigo) return
          encerrarScanner()
          setEntrada(codigo)
          setOrigemEntrada('scanner')
          setDigitando(true)
          setAvisoScanner('Código lido. Confira ou complemente com o endereço antes de adicionar.')
          setErro('')
        },
      )
      if (!scannerAtivoRef.current) {
        controles.stop()
        return
      }
      controlesScannerRef.current = controles
    } catch (e) {
      if (!scannerAtivoRef.current) return
      encerrarScanner()
      setErro(e.name === 'NotAllowedError'
        ? 'Permita o acesso à câmera para escanear a etiqueta.'
        : 'Não foi possível abrir a câmera. Use Digitar ou Falar.')
    }
  }

  function encerrarScanner() {
    scannerAtivoRef.current = false
    controlesScannerRef.current?.stop()
    controlesScannerRef.current = null
    leitorScannerRef.current?.reset()
    leitorScannerRef.current = null
    if (videoScannerRef.current) {
      videoScannerRef.current.pause()
      videoScannerRef.current.srcObject = null
    }
    setEscaneando(false)
  }

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
      setOrdenamento(await adicionarObjeto(ordenamento.id, texto, origemEntrada))
      setEntrada('')
      setOrigemEntrada('manual')
      setAvisoScanner('')
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

  async function confirmarRua(objetoId, ruaId) {
    if (!ordenamento || resolvendoId !== null) return
    setResolvendoId(objetoId)
    setErro('')
    try {
      setOrdenamento(await selecionarRua(ordenamento.id, objetoId, ruaId))
    } catch (e) {
      setErro(e.message)
    } finally {
      setResolvendoId(null)
    }
  }

  async function gerarOrdenamento() {
    if (!ordenamento || gerando) return
    setGerando(true)
    setErro('')
    try {
      setOrdenamento(await gerarOrdem(ordenamento.id))
    } catch (e) {
      setErro(e.message)
    } finally {
      setGerando(false)
    }
  }

  async function limparLista() {
    if (!ordenamento || limpando || (ordenamento.total_objetos ?? 0) === 0) return
    const confirmado = window.confirm(
      'Limpar esta lista? As encomendas adicionadas e a ordem sugerida serão removidas. O cadastro geral de ruas não será alterado.',
    )
    if (!confirmado) return

    setLimpando(true)
    setErro('')
    try {
      setOrdenamento(await limparOrdenamento(ordenamento.id))
      setEntrada('')
      setOrigemEntrada('manual')
      setAvisoScanner('')
      setDigitando(false)
    } catch (e) {
      setErro(e.message)
    } finally {
      setLimpando(false)
    }
  }

  const motivoBloqueio = mensagemBloqueio(ordenamento)

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
            <div className={styles.andamentoAcoes}>
              <span className={styles.data}>Iniciado {formatarData(ordenamento.created_at)}</span>
              {(ordenamento.total_objetos ?? 0) > 0 && (
                <button type="button" className={styles.limparLista} onClick={limparLista} disabled={limpando}>
                  {limpando ? 'Limpando…' : 'Limpar lista'}
                </button>
              )}
            </div>
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
            <button
              type="button"
              className={escaneando ? styles.modoAtivo : ''}
              onClick={escaneando ? encerrarScanner : iniciarScanner}
              disabled={salvando}
              aria-pressed={escaneando}
            >
              <span aria-hidden="true">▣</span> {escaneando ? 'Parar câmera' : 'Escanear'}
            </button>
            <button
              type="button"
              className={ouvindo ? styles.modoAtivo : ''}
              onClick={iniciarEntradaPorVoz}
              disabled={salvando}
              aria-pressed={ouvindo}
            >
              <span aria-hidden="true">●</span> {ouvindo ? 'Ouvindo…' : 'Falar'}
            </button>
            <button
              type="button"
              className={digitando ? styles.modoAtivo : ''}
              onClick={() => {
                setOrigemEntrada('manual')
                setAvisoScanner('')
                setDigitando(true)
              }}
            >
              <span aria-hidden="true">⌨</span> Digitar
            </button>
          </div>

          {digitando && (
            <form className={styles.formulario} onSubmit={cadastrarObjeto}>
              <label htmlFor="endereco-encomenda">{avisoScanner ? 'Código lido — confira o endereço' : 'Rua e número'}</label>
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
              <span>{avisoScanner || 'Após adicionar, o campo fica pronto para a próxima encomenda.'}</span>
            </form>
          )}

          {escaneando && (
            <div className={styles.scanner} aria-label="Leitura da etiqueta">
              <video ref={videoScannerRef} autoPlay muted playsInline />
              <p>Aponte a câmera para um código da etiqueta.</p>
              <button type="button" onClick={encerrarScanner}>Cancelar</button>
            </div>
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
                          {(objeto.opcoes_resolucao?.length ?? 0) > 0 && (
                            <div className={styles.opcoesResolucao}>
                              <span>Escolha o cadastro correto:</span>
                              {objeto.opcoes_resolucao.map((opcao) => (
                                <button
                                  key={opcao.rua_id}
                                  type="button"
                                  onClick={() => confirmarRua(objeto.id, opcao.rua_id)}
                                  disabled={resolvendoId !== null}
                                >
                                  <strong>{opcao.nome_rua}</strong>
                                  <small>
                                    {opcao.distrito ? `Distrito ${opcao.distrito}` : ''}
                                    {opcao.distrito && opcao.cep ? ' · ' : ''}
                                    {opcao.cep ? `CEP ${opcao.cep}` : ''}
                                  </small>
                                </button>
                              ))}
                            </div>
                          )}
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

          <section className={styles.geracao} aria-labelledby="titulo-gerar-ordem">
            <div>
              <h3 id="titulo-gerar-ordem">Ordenamento por proximidade</h3>
              <p>Gera uma sequência sugerida a partir do CDD Campos dos Goytacazes.</p>
              {motivoBloqueio && <span>{motivoBloqueio}</span>}
            </div>
            <button
              type="button"
              onClick={gerarOrdenamento}
              disabled={Boolean(motivoBloqueio) || gerando}
            >
              {gerando ? 'Gerando…' : ordenamento.ordem_sugerida?.length ? 'Gerar novamente' : 'Gerar ordenamento'}
            </button>
          </section>

          {(ordenamento.ordem_sugerida?.length ?? 0) > 0 && (
            <section className={styles.ordem} aria-labelledby="titulo-ordem-sugerida">
              <div className={styles.listaTopo}>
                <h3 id="titulo-ordem-sugerida">Ordem sugerida</h3>
                <span>{ordenamento.ordem_sugerida.length} {ordenamento.ordem_sugerida.length === 1 ? 'parada' : 'paradas'}</span>
              </div>
              <p>Sequência calculada pela proximidade entre as ruas, começando no CDD.</p>
              <ol>
                {ordenamento.ordem_sugerida.map((parada) => (
                  <li key={parada.id ?? parada.chave_agrupamento}>
                    <span className={styles.numeroOrdem}>{parada.ordem_sugerida}</span>
                    <span>{parada.nome_rua}</span>
                    <strong>{parada.quantidade_objetos} {parada.quantidade_objetos === 1 ? 'objeto' : 'objetos'}</strong>
                  </li>
                ))}
              </ol>
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

function mensagemBloqueio(ordenamento) {
  if (!ordenamento) return ''
  if ((ordenamento.total_pendentes ?? 0) > 0) return 'Revise as encomendas pendentes antes de gerar a ordem.'
  if ((ordenamento.total_sem_coordenadas ?? 0) > 0) return 'Aguarde a coordenada de todas as ruas identificadas.'
  if ((ordenamento.total_ruas ?? 0) === 0) return 'Adicione ao menos uma encomenda identificada para gerar a ordem.'
  return ''
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
