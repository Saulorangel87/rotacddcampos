import { useEffect, useRef, useState } from 'react'
import {
  adicionarObjeto,
  buscarOrdenamentoAtivo,
  criarOrdenamento,
  excluirObjeto,
  esquecerReferencia,
  gerarOrdem,
  limparOrdenamento,
  salvarOrdemFinal,
  selecionarRua,
} from '../../api/ordenamentos.js'
import { useReconhecimentoDeVoz } from '../../hooks/useReconhecimentoDeVoz.js'
import HistoricoOrdens from './HistoricoOrdens.jsx'
import { ordenarParadas } from './ordemManual.js'
import { IconeMicrofone, IconeScanner, IconeTeclado, IconeUnidadeCorreios } from '../icons/Icons.jsx'
import styles from './OrdenamentoPanel.module.css'

function instrucaoPermissaoCamera() {
  const agente = navigator.userAgent || ''
  if (/Android/i.test(agente)) {
    return 'No Android, a permissão é gerenciada pelo Chrome: abra o site no Chrome, toque no ícone de controles ao lado do endereço > Permissões > Câmera > Permitir. Se Câmera não aparecer, abra Chrome > Configurações > Configurações do site > Câmera, remova o bloqueio e tente novamente.'
  }
  if (/iPhone|iPad|iPod/i.test(agente)) {
    return 'No iPhone ou iPad, abra Ajustes > Apps > Safari > Câmera e selecione Perguntar ou Permitir. Depois feche e abra o app novamente.'
  }
  return 'Abra as permissões do site no cadeado ao lado do endereço e permita o uso da câmera. Depois tente novamente.'
}

function mensagemErroScanner(erro) {
  if (erro?.name === 'NotAllowedError' || erro?.name === 'SecurityError') {
    return `A câmera está bloqueada. ${instrucaoPermissaoCamera()}`
  }
  if (erro?.name === 'NotReadableError') {
    return 'A câmera está sendo usada por outro aplicativo. Feche a câmera e outros leitores, depois tente novamente.'
  }
  return 'Não foi possível abrir a câmera. Use Digitar ou Falar.'
}

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
  const [feedbackAdicao, setFeedbackAdicao] = useState(null)
  const [ordemEditada, setOrdemEditada] = useState(null)
  const [salvandoOrdemFinal, setSalvandoOrdemFinal] = useState(false)
  const [paradaArrastadaId, setParadaArrastadaId] = useState(null)
  const [reutilizarOrdem, setReutilizarOrdem] = useState(false)
  const [esquecendo, setEsquecendo] = useState(false)
  const campoEntradaRef = useRef(null)
  const videoScannerRef = useRef(null)
  const leitorScannerRef = useRef(null)
  const controlesScannerRef = useRef(null)
  const scannerAtivoRef = useRef(false)
  const ultimoCodigoScannerRef = useRef('')
  const { ouvindo, ouvirVoz } = useReconhecimentoDeVoz((textoTranscrito) => {
    setEntrada(textoTranscrito)
    setOrigemEntrada('voz')
    setDigitando(true)
    setErro('')
  })

  function iniciarEntradaPorVoz() {
    encerrarScanner()
    ocultarTecladoEntrada()
    setOrigemEntrada('voz')
    setDigitando(true)
    setAvisoScanner('')
    setFeedbackAdicao(null)
    ouvirVoz()
  }

  useEffect(() => () => encerrarScanner(), [])

  async function iniciarScanner() {
    setErro('')
    setAvisoScanner('')
    setFeedbackAdicao(null)
    ultimoCodigoScannerRef.current = ''
    ocultarTecladoEntrada()
    setDigitando(false)
    if (!navigator.mediaDevices?.getUserMedia) {
      setErro('A câmera não está disponível neste navegador. Use Digitar ou Falar.')
      return
    }

    try {
      try {
        const permissao = await navigator.permissions?.query?.({ name: 'camera' })
        if (permissao?.state === 'denied') {
          setErro(mensagemErroScanner({ name: 'NotAllowedError' }))
          return
        }
      } catch {
        // Alguns navegadores não expõem camera em Permissions API; o leitor
        // fará a solicitação normalmente ao abrir o dispositivo.
      }
      const { BrowserMultiFormatReader } = await import('@zxing/browser')
      leitorScannerRef.current = new BrowserMultiFormatReader()
      scannerAtivoRef.current = true
      setEscaneando(true)
      requestAnimationFrame(iniciarLeituraScanner)
    } catch (e) {
      encerrarScanner()
      setErro(mensagemErroScanner(e))
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

          const cep = extrairCepScanner(codigo)
          if (!cep) {
            // O leitor pode encontrar o código de rastreio antes do código
            // postal. Mantemos a câmera aberta e ignoramos qualquer leitura
            // que não seja exclusivamente um CEP de oito dígitos.
            if (ultimoCodigoScannerRef.current !== codigo) {
              ultimoCodigoScannerRef.current = codigo
              setAvisoScanner('Código ignorado. Aponte para o código de barras do CEP com 8 dígitos.')
            }
            return
          }

          encerrarScanner()
          setEntrada(cep)
          setOrigemEntrada('scanner')
          setDigitando(true)
          setAvisoScanner(mensagemCodigoLido(cep))
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
      setErro(mensagemErroScanner(e))
    }
  }

  function encerrarScanner() {
    scannerAtivoRef.current = false
    controlesScannerRef.current?.stop()
    controlesScannerRef.current = null
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
    setOrdemEditada(null)
    setReutilizarOrdem(false)
    setParadaArrastadaId(null)
  }, [ordenamento])

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
    if (!texto || !ordenamento || operacaoEmAndamento) return

    setSalvando(true)
    setErro('')
    setFeedbackAdicao(null)
    try {
      const origemAtual = origemEntrada
      const atualizado = await adicionarObjeto(ordenamento.id, texto, origemAtual)
      setOrdenamento(atualizado)
      const objetoAdicionado = atualizado.objetos?.[atualizado.objetos.length - 1]
      setFeedbackAdicao(feedbackDaAdicao(objetoAdicionado, origemAtual, texto))
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
    if (!ordenamento || operacaoEmAndamento) return
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
    if (!ordenamento || operacaoEmAndamento) return
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

  async function gerarOrdenamento(somenteAlgoritmo = false) {
    if (!ordenamento || operacaoEmAndamento || ordemFoiAlterada) return
    setGerando(true)
    setErro('')
    try {
      setOrdenamento(await gerarOrdem(ordenamento.id, somenteAlgoritmo))
    } catch (e) {
      setErro(e.message)
    } finally {
      setGerando(false)
    }
  }

  async function limparLista() {
    if (!ordenamento || operacaoEmAndamento || (ordenamento.total_objetos ?? 0) === 0) return
    const confirmado = window.confirm(
      'Limpar esta carga? As encomendas e a ordem atual serão removidas. O histórico salvo e suas sequências habituais serão mantidos.',
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

  function moverParada(indice, deslocamento) {
    if (operacaoEmAndamento) return
    setOrdemEditada((atual) => {
      const base = atual ?? ordenarParadas(ordenamento?.ordem_sugerida ?? [])
      const destino = indice + deslocamento
      if (destino < 0 || destino >= base.length) return base
      const nova = [...base]
      const [movida] = nova.splice(indice, 1)
      nova.splice(destino, 0, movida)
      return nova
    })
  }

  function iniciarArraste(evento, paradaId) {
    if (operacaoEmAndamento) { evento.preventDefault(); return }
    setParadaArrastadaId(paradaId)
    evento.dataTransfer.effectAllowed = 'move'
    evento.dataTransfer.setData('text/plain', String(paradaId))
  }

  function soltarParada(evento, destinoId) {
    evento.preventDefault()
    if (operacaoEmAndamento) { setParadaArrastadaId(null); return }
    const origemId = paradaArrastadaId || Number(evento.dataTransfer.getData('text/plain'))
    if (!origemId || origemId === destinoId) {
      setParadaArrastadaId(null)
      return
    }

    setOrdemEditada((atual) => {
      const base = atual ?? ordenarParadas(ordenamento?.ordem_sugerida ?? [])
      const origem = base.findIndex((parada) => parada.id === origemId)
      const destino = base.findIndex((parada) => parada.id === destinoId)
      if (origem < 0 || destino < 0) return base
      const nova = [...base]
      const [movida] = nova.splice(origem, 1)
      nova.splice(destino, 0, movida)
      return nova
    })
    setParadaArrastadaId(null)
  }

  async function salvarOrdemManual() {
    if (!ordenamento || operacaoEmAndamento || (!ordemFoiAlterada && !reutilizarOrdem) || ordemAtual.length === 0) return
    setSalvandoOrdemFinal(true)
    setErro('')
    try {
      setOrdenamento(await salvarOrdemFinal(ordenamento.id, ordemAtual.map((parada) => parada.id), reutilizarOrdem))
    } catch (e) {
      setErro(e.message)
    } finally {
      setSalvandoOrdemFinal(false)
    }
  }

  const motivoBloqueio = mensagemBloqueio(ordenamento)
  const operacaoEmAndamento = salvando || excluindoId !== null || resolvendoId !== null || limpando || gerando || salvandoOrdemFinal || esquecendo
  const ordemBase = ordenarParadas(ordenamento?.ordem_sugerida ?? [])
  const ordemAtual = ordemEditada ?? ordemBase
  const ordemFoiAlterada = ordemAtual.length === ordemBase.length
    && ordemAtual.some((parada, indice) => parada.id !== ordemBase[indice]?.id)
  const temOrdemFinal = ordemAtual.some((parada) => parada.ordem_final != null)

  async function desativarReferencia(referenciaId) {
    if (operacaoEmAndamento || ordemFoiAlterada) return
    if (!window.confirm('Deixar de reutilizar esta sequência? O histórico e a ordem desta carga serão mantidos.')) return
    setEsquecendo(true)
    setErro('')
    try {
      await esquecerReferencia(referenciaId)
      setOrdenamento(await buscarOrdenamentoAtivo())
    } catch (e) { setErro(e.message) }
    finally { setEsquecendo(false) }
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
        <div className={styles.marcador} role="img" aria-label="Unidade CDD Campos dos Goytacazes">
          <IconeUnidadeCorreios size={22} aria-hidden="true" />
          <span className={styles.marcadorSigla}>CDD</span>
        </div>
        <div>
          <span className={styles.rotulo}>Ponto de partida fixo</span>
          <strong>Campos dos Goytacazes</strong>
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
                <button type="button" className={styles.limparLista} onClick={limparLista} disabled={operacaoEmAndamento}>
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
              <IconeScanner size={18} aria-hidden="true" /> {escaneando ? 'Parar câmera' : 'Escanear'}
            </button>
            <button
              type="button"
              className={ouvindo ? styles.modoAtivo : ''}
              onClick={iniciarEntradaPorVoz}
              disabled={salvando}
              aria-pressed={ouvindo}
            >
              <IconeMicrofone size={18} aria-hidden="true" /> {ouvindo ? 'Ouvindo…' : 'Falar'}
            </button>
            <button
              type="button"
              className={digitando ? styles.modoAtivo : ''}
              onClick={() => {
                setOrigemEntrada('manual')
                setAvisoScanner('')
                setFeedbackAdicao(null)
                setDigitando(true)
              }}
            >
              <IconeTeclado size={18} aria-hidden="true" /> Digitar
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
                <button type="submit" disabled={operacaoEmAndamento || !entrada.trim()}>
                  {salvando ? 'Adicionando…' : 'Adicionar'}
                </button>
              </div>
              <span>{avisoScanner || 'Após adicionar, o campo fica pronto para a próxima encomenda.'}</span>
            </form>
          )}

          {escaneando && (
            <div className={styles.scanner} aria-label="Leitura da etiqueta">
              <video ref={videoScannerRef} autoPlay muted playsInline />
              <p>{avisoScanner || 'Aponte a câmera para o código de barras do CEP (8 dígitos).'}</p>
              <button type="button" onClick={encerrarScanner}>Cancelar</button>
            </div>
          )}

          {feedbackAdicao && (
            <p className={`${styles.feedback} ${styles[`feedback${feedbackAdicao.tipo}`]}`} role="status">
              {feedbackAdicao.texto}
            </p>
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
                                  disabled={operacaoEmAndamento}
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
                      disabled={operacaoEmAndamento}
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
              {ordemFoiAlterada && <span>Salve ou descarte o ajuste antes de gerar novamente.</span>}
            </div>
            <div className={styles.geracaoAcoes}>
              <button
                type="button"
                onClick={() => gerarOrdenamento(false)}
                disabled={Boolean(motivoBloqueio) || operacaoEmAndamento || ordemFoiAlterada}
              >
                {gerando ? 'Gerando…' : ordenamento.ordem_sugerida?.length ? 'Gerar novamente' : 'Gerar ordenamento'}
              </button>
              {ordemAtual.length > 0 && (
                <button type="button" className={styles.botaoSecundario} onClick={() => gerarOrdenamento(true)} disabled={Boolean(motivoBloqueio) || operacaoEmAndamento || ordemFoiAlterada}>
                  Recalcular sem meus ajustes
                </button>
              )}
            </div>
          </section>

          {(ordenamento.ordem_sugerida?.length ?? 0) > 0 && (
            <section className={styles.ordem} aria-labelledby="titulo-ordem-sugerida">
              <div className={styles.listaTopo}>
                <h3 id="titulo-ordem-sugerida">{temOrdemFinal ? 'Ordem final' : 'Ordem sugerida'}</h3>
                <span>{ordemAtual.length} {ordemAtual.length === 1 ? 'parada' : 'paradas'}</span>
              </div>
              <p>
                {temOrdemFinal
                  ? 'Sequência ajustada manualmente. Arraste uma rua ou use os botões para corrigir.'
                  : 'Sequência calculada pela proximidade entre as ruas, começando no CDD. Arraste uma rua ou use os botões para ajustar.'}
              </p>
              <ol>
                {ordemAtual.map((parada, indice) => (
                  <li
                    key={parada.id ?? parada.chave_agrupamento}
                    className={paradaArrastadaId === parada.id ? styles.paradaArrastada : ''}
                    draggable={!operacaoEmAndamento}
                    onDragStart={(evento) => iniciarArraste(evento, parada.id)}
                    onDragOver={(evento) => evento.preventDefault()}
                    onDrop={(evento) => soltarParada(evento, parada.id)}
                    onDragEnd={() => setParadaArrastadaId(null)}
                  >
                    <span className={styles.numeroOrdem}>{indice + 1}</span>
                    <span>{parada.nome_rua}</span>
                    <strong>{parada.quantidade_objetos} {parada.quantidade_objetos === 1 ? 'objeto' : 'objetos'}</strong>
                    <div className={styles.acoesOrdem} aria-label={`Ajustar posição de ${parada.nome_rua}`}>
                      <button
                        type="button"
                        onClick={() => moverParada(indice, -1)}
                        disabled={indice === 0 || operacaoEmAndamento}
                        aria-label={`Mover ${parada.nome_rua} para cima`}
                      >
                        ↑
                      </button>
                      <button
                        type="button"
                        onClick={() => moverParada(indice, 1)}
                        disabled={indice === ordemAtual.length - 1 || operacaoEmAndamento}
                        aria-label={`Mover ${parada.nome_rua} para baixo`}
                      >
                        ↓
                      </button>
                    </div>
                  </li>
                ))}
              </ol>
              <label className={styles.opcaoMemoria}>
                <input type="checkbox" checked={reutilizarOrdem} onChange={(evento) => setReutilizarOrdem(evento.target.checked)} disabled={operacaoEmAndamento || !ordenamento.permite_sequencia_pessoal} />
                Usar como minha sequência habitual
              </label>
              <p>
                {ordenamento.permite_sequencia_pessoal
                  ? 'Ao salvar com esta opção, a sequência será usada nas suas próximas cargas com as mesmas ruas. Ela vale apenas para você.'
                  : 'Você pode salvar o ajuste desta carga. Para reutilizar a sequência, identifique os trechos agrupados pelo CEP.'}
              </p>
              {ordenamento.referencia_pessoal_id && (
                <p>Você tem uma sequência habitual guardada para estas ruas.</p>
              )}
              {(ordemFoiAlterada || reutilizarOrdem) && (
                <div className={styles.ordemAcoes}>
                  <span>{reutilizarOrdem ? 'Salvar também para suas próximas cargas.' : 'Salvar somente para esta carga.'}</span>
                  <button type="button" className={styles.botaoSecundario} onClick={() => { setOrdemEditada(null); setReutilizarOrdem(false) }} disabled={operacaoEmAndamento}>Descartar ajuste</button>
                  <button type="button" onClick={salvarOrdemManual} disabled={operacaoEmAndamento}>
                    {salvandoOrdemFinal ? 'Salvando…' : 'Salvar ordem final'}
                  </button>
                </div>
              )}
              {temOrdemFinal && !ordemFoiAlterada && (
                <p className={styles.ordemSalva} role="status">
                  {ordemAtual.every((parada) => parada.fonte_ordem_final === 'pessoal') ? 'Sua sequência habitual foi aplicada nesta carga.' : 'Ordem final salva. Gerar novamente mantém esta sequência.'}
                </p>
              )}
            </section>
          )}
          <HistoricoOrdens
            historico={ordenamento.historico_ordens ?? []}
            paradas={ordemBase}
            ocupado={operacaoEmAndamento || ordemFoiAlterada}
            onRestaurar={(sequencia) => { setOrdemEditada(sequencia); setReutilizarOrdem(false); setErro('') }}
            onEsquecer={desativarReferencia}
          />
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
  if (motivo === 'rua_aproximada') return 'Precisa de revisão · Encontramos uma rua parecida'
  if (motivo === 'cep_nao_encontrado') return 'Precisa de revisão · CEP não encontrado no cadastro'
  return 'Precisa de revisão · Rua não encontrada no cadastro'
}

function ocultarTecladoEntrada() {
  const ativo = document.activeElement
  if (ativo instanceof HTMLElement) ativo.blur()
  try {
    navigator.virtualKeyboard?.hide?.()
  } catch {
    // A API é opcional e alguns navegadores lançam quando não há teclado.
  }
}

function extrairCepScanner(codigo) {
  const texto = String(codigo ?? '').trim()
  if (!/^[\d\s.-]+$/.test(texto)) return ''
  const digitos = texto.replace(/\D/g, '')
  if (digitos.length !== 8) return ''
  return `${digitos.slice(0, 5)}-${digitos.slice(5)}`
}

function mensagemCodigoLido(codigo) {
  if (/^\d{5}-?\d{3}$/.test(codigo)) {
    return 'CEP lido. Confira o endereço encontrado e adicione a encomenda.'
  }
  if (/^[A-Z]{2}\d{9}[A-Z]{2}$/i.test(codigo)) {
    return 'Código de rastreio lido. Informe ou complemente o endereço antes de adicionar.'
  }
  return 'Código lido. Confira ou complemente o endereço antes de adicionar.'
}

function feedbackDaAdicao(objeto, origem, entrada) {
  if (objeto?.status_resolucao === 'identificado') {
    return {
      tipo: 'Sucesso',
      texto: `✓ ${objeto.nome_rua}${objeto.cep ? ` · CEP ${objeto.cep}` : ''} adicionado ao ordenamento.`,
    }
  }
  if (origem === 'scanner' && /^\d{5}-?\d{3}$/.test(entrada)) {
    if (objeto?.motivo_pendencia === 'rua_ambigua') {
      return { tipo: 'Aviso', texto: 'CEP lido. Escolha a rua correspondente nas opções abaixo.' }
    }
    if (objeto?.motivo_pendencia === 'cep_nao_encontrado') {
      return { tipo: 'Aviso', texto: 'CEP lido, mas ele não foi encontrado no cadastro desta unidade.' }
    }
  }
  return { tipo: 'Aviso', texto: 'Encomenda adicionada para revisão. Confira o motivo indicado abaixo.' }
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
