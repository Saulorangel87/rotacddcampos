import { useEffect, useState, useMemo } from 'react'
import {
  buscarPlanoAtivo,
  criarPlano,
  reatribuirRua,
  concluirPlano,
  aplicarPlano,
  cancelarPlano,
} from '../../api/redistritamento.js'
import { apiFetchJson } from '../../api/client.js'
import { corDoDistrito } from '../../data/distritos.js'
import ConfirmModal from '../ConfirmModal.jsx'
import styles from './RedistritamentoPanel.module.css'

export default function RedistritamentoPanel() {
  const [carregando, setCarregando] = useState(true)
  const [erro, setErro] = useState('')
  const [plano, setPlano] = useState(null)
  const [distritosAtivos, setDistritosAtivos] = useState([])

  const [quantidadeAlvo, setQuantidadeAlvo] = useState('')
  const [criando, setCriando] = useState(false)

  const [salvandoLinhaId, setSalvandoLinhaId] = useState(null)
  const [concluindo, setConcluindo] = useState(false)
  const [aplicando, setAplicando] = useState(false)
  const [modalAplicarAberto, setModalAplicarAberto] = useState(false)
  const [modalCancelarAberto, setModalCancelarAberto] = useState(false)
  const [cancelando, setCancelando] = useState(false)
  const [destinoEmLote, setDestinoEmLote] = useState({}) // { [distritoOrigem]: codigoDestino }

  async function carregarTudo() {
    setCarregando(true)
    setErro('')
    try {
      const [planoAtivo, listaDistritos] = await Promise.all([
        buscarPlanoAtivo(),
        apiFetchJson('/distritos'),
      ])
      setPlano(planoAtivo)
      setDistritosAtivos(listaDistritos || [])
    } catch (e) {
      setErro(e.message)
    } finally {
      setCarregando(false)
    }
  }

  useEffect(() => {
    carregarTudo()
  }, [])

  const codigosExtintos = useMemo(
    () => new Set((plano?.distritos_extintos || '').split(',').filter(Boolean)),
    [plano]
  )

  const distritosDestinoPossiveis = useMemo(
    () => distritosAtivos.filter((d) => !codigosExtintos.has(d.codigo)),
    [distritosAtivos, codigosExtintos]
  )

  const ruas = plano?.ruas || []
  const totalRuas = ruas.length
  const totalReatribuidas = ruas.filter((r) => r.distrito_destino).length
  const progressoPct = totalRuas ? Math.round((totalReatribuidas / totalRuas) * 100) : 0
  const tudoReatribuido = totalRuas > 0 && totalReatribuidas === totalRuas

  const gruposPorOrigem = useMemo(() => {
    const mapa = new Map()
    for (const rua of ruas) {
      if (!mapa.has(rua.distrito_origem)) mapa.set(rua.distrito_origem, [])
      mapa.get(rua.distrito_origem).push(rua)
    }
    return [...mapa.entries()].sort((a, b) => b[0].localeCompare(a[0]))
  }, [ruas])

  async function handleIniciarPlano(e) {
    e.preventDefault()
    const alvo = parseInt(quantidadeAlvo, 10)
    if (!alvo || alvo <= 0) {
      setErro('Informe uma quantidade alvo válida.')
      return
    }
    setCriando(true)
    setErro('')
    try {
      const novoPlano = await criarPlano(alvo)
      setPlano(novoPlano)
      setQuantidadeAlvo('')
    } catch (e) {
      setErro(e.message)
    } finally {
      setCriando(false)
    }
  }

  async function handleMudarDestino(linha, novoDestino) {
    setSalvandoLinhaId(linha.id)
    setErro('')
    // Atualização otimista — a tela responde na hora, sem esperar o servidor
    setPlano((atual) => ({
      ...atual,
      ruas: atual.ruas.map((r) => (r.id === linha.id ? { ...r, distrito_destino: novoDestino } : r)),
    }))
    try {
      await reatribuirRua(plano.id, linha.id, novoDestino)
    } catch (e) {
      setErro(e.message)
      // desfaz a atualização otimista em caso de erro
      setPlano((atual) => ({
        ...atual,
        ruas: atual.ruas.map((r) => (r.id === linha.id ? { ...r, distrito_destino: linha.distrito_destino } : r)),
      }))
    } finally {
      setSalvandoLinhaId(null)
    }
  }

  async function handleAplicarEmLote(distritoOrigem) {
    const destino = destinoEmLote[distritoOrigem]
    if (!destino) return
    const linhasDoGrupo = ruas.filter((r) => r.distrito_origem === distritoOrigem)
    for (const linha of linhasDoGrupo) {
      await handleMudarDestino(linha, destino)
    }
  }

  async function handleConcluir() {
    setConcluindo(true)
    setErro('')
    try {
      const atualizado = await concluirPlano(plano.id)
      setPlano(atualizado)
    } catch (e) {
      setErro(e.message)
    } finally {
      setConcluindo(false)
    }
  }

  async function handleConfirmarAplicar() {
    setAplicando(true)
    setErro('')
    try {
      await aplicarPlano(plano.id)
      setModalAplicarAberto(false)
      // Depois de aplicado, o plano deixa de ser "ativo" — recarrega tudo
      // (inclusive a lista de distritos, que agora vai vir sem os extintos)
      await carregarTudo()
    } catch (e) {
      setErro(e.message)
      setModalAplicarAberto(false)
    } finally {
      setAplicando(false)
    }
  }

  async function handleConfirmarCancelar() {
    setCancelando(true)
    setErro('')
    try {
      await cancelarPlano(plano.id)
      setModalCancelarAberto(false)
      await carregarTudo()
    } catch (e) {
      setErro(e.message)
      setModalCancelarAberto(false)
    } finally {
      setCancelando(false)
    }
  }

  if (carregando) {
    return (
      <div className={styles.caixa}>
        <p className={styles.dica}>Carregando…</p>
      </div>
    )
  }

  // ─── Sem plano ativo: tela de início ────────────────────────────────
  if (!plano) {
    return (
      <div className={styles.caixa}>
        <h2 className={styles.titulo}>Redistritamento</h2>
        <p className={styles.dica}>
          Reduz a quantidade de distritos de forma controlada. O corte é sempre feito a partir
          dos distritos de maior código — as ruas deles ficam "órfãs" pra você realocar
          manualmente antes de aplicar qualquer mudança de verdade.
        </p>

        <form onSubmit={handleIniciarPlano} className={styles.formNovoPlano}>
          <div className={styles.campoNovoPlano}>
            <label htmlFor="quantidadeAlvo">Quantidade final de distritos</label>
            <input
              id="quantidadeAlvo"
              type="number"
              min="1"
              max={distritosAtivos.length - 1}
              value={quantidadeAlvo}
              onChange={(e) => setQuantidadeAlvo(e.target.value)}
              placeholder={`Hoje: ${distritosAtivos.length}`}
            />
          </div>
          <button type="submit" className={styles.botaoPrimario} disabled={criando}>
            {criando ? 'Iniciando…' : 'Iniciar plano'}
          </button>
        </form>

        {erro && <p className={styles.erro}>{erro}</p>}
      </div>
    )
  }

  // ─── Plano ativo: tela de trabalho ──────────────────────────────────
  return (
    <div className={styles.caixa}>
      <div className={styles.cabecalho}>
        <div>
          <button
            type="button"
            className={styles.botaoVoltar}
            onClick={() => setModalCancelarAberto(true)}
          >
            ← Voltar ao início
          </button>
          <h2 className={styles.titulo}>Redistritamento em andamento</h2>
          <p className={styles.dica}>
            Reduzindo de {plano.quantidade_atual} para {plano.quantidade_alvo} distritos —
            distritos extintos: {[...codigosExtintos].sort().join(', ')}
          </p>
        </div>
        <span className={plano.status === 'concluido' ? styles.tagConcluido : styles.tagRascunho}>
          {plano.status === 'concluido' ? 'Concluído (revisável)' : 'Rascunho'}
        </span>
      </div>

      <div className={styles.progresso}>
        <div className={styles.barraFundo}>
          <div className={styles.barraPreenchida} style={{ width: `${progressoPct}%` }} />
        </div>
        <span className={styles.progressoTexto}>
          {totalReatribuidas} de {totalRuas} ruas realocadas ({progressoPct}%)
        </span>
      </div>

      {erro && <p className={styles.erro}>{erro}</p>}

      <div className={styles.grupos}>
        {gruposPorOrigem.map(([origem, linhas]) => {
          const reatribuidasDoGrupo = linhas.filter((l) => l.distrito_destino).length
          return (
            <div key={origem} className={styles.grupo}>
              <div className={styles.grupoCabecalho}>
                <span
                  className={styles.chipDistrito}
                  style={{ background: corDoDistrito(origem) }}
                >
                  {origem}
                </span>
                <span className={styles.grupoTitulo}>
                  {linhas.length} ruas — {reatribuidasDoGrupo}/{linhas.length} realocadas
                </span>

                <div className={styles.loteWrapper}>
                  <select
                    value={destinoEmLote[origem] || ''}
                    onChange={(e) =>
                      setDestinoEmLote((atual) => ({ ...atual, [origem]: e.target.value }))
                    }
                  >
                    <option value="">Atribuir todas para…</option>
                    {distritosDestinoPossiveis.map((d) => (
                      <option key={d.codigo} value={d.codigo}>
                        {d.codigo}
                      </option>
                    ))}
                  </select>
                  <button
                    type="button"
                    className={styles.botaoSecundario}
                    disabled={!destinoEmLote[origem]}
                    onClick={() => handleAplicarEmLote(origem)}
                  >
                    Aplicar ao grupo
                  </button>
                </div>
              </div>

              <ul className={styles.listaRuas}>
                {linhas.map((linha) => (
                  <li key={linha.id} className={styles.linhaRua}>
                    <div className={styles.infoRua}>
                      <span className={styles.nomeRua}>{linha.nome_rua}</span>
                      {linha.bairro && <span className={styles.bairroRua}>{linha.bairro}</span>}
                    </div>
                    <select
                      className={linha.distrito_destino ? styles.selectPreenchido : styles.selectVazio}
                      value={linha.distrito_destino || ''}
                      disabled={salvandoLinhaId === linha.id || plano.status === 'aplicado'}
                      onChange={(e) => handleMudarDestino(linha, e.target.value)}
                    >
                      <option value="">Escolher destino…</option>
                      {distritosDestinoPossiveis.map((d) => (
                        <option key={d.codigo} value={d.codigo}>
                          {d.codigo} — {d.nome}
                        </option>
                      ))}
                    </select>
                  </li>
                ))}
              </ul>
            </div>
          )
        })}
      </div>

      <div className={styles.rodape}>
        {plano.status === 'rascunho' && (
          <button
            type="button"
            className={styles.botaoPrimario}
            disabled={!tudoReatribuido || concluindo}
            onClick={handleConcluir}
            title={!tudoReatribuido ? 'Realoque todas as ruas órfãs antes de concluir' : undefined}
          >
            {concluindo ? 'Concluindo…' : 'Concluir'}
          </button>
        )}

        {plano.status === 'concluido' && (
          <button
            type="button"
            className={styles.botaoPerigo}
            onClick={() => setModalAplicarAberto(true)}
          >
            Aplicar (definitivo)
          </button>
        )}
      </div>

      <ConfirmModal
        aberto={modalAplicarAberto}
        titulo="Aplicar redistritamento"
        mensagem={`Isso vai mudar o distrito de ${totalRuas} ruas de verdade no banco de dados, desativar ${codigosExtintos.size} distritos e atualizar mapa, relatórios e estatísticas pra todo mundo que usa o site. Essa ação NÃO pode ser desfeita pela tela — tem certeza que quer aplicar agora?`}
        onConfirmar={handleConfirmarAplicar}
        onCancelar={() => setModalAplicarAberto(false)}
        confirmando={aplicando}
        textoConfirmar="Aplicar definitivamente"
        textoConfirmando="Aplicando…"
      />

      <ConfirmModal
        aberto={modalCancelarAberto}
        titulo="Descartar este plano?"
        mensagem={`Isso apaga o rascunho e todo o trabalho de realocação feito até agora (${totalReatribuidas} ruas já realocadas). Nada no banco de dados real é afetado — o site continua exatamente como está. Quer voltar ao início e começar um plano novo?`}
        onConfirmar={handleConfirmarCancelar}
        onCancelar={() => setModalCancelarAberto(false)}
        confirmando={cancelando}
        textoConfirmar="Descartar e voltar"
        textoConfirmando="Descartando…"
      />
    </div>
  )
}
