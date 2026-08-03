import { useEffect, useState } from 'react'
import { DISTRITOS } from '../../data/distritos.js'
import { listarRuas, moverRuasEmLote } from '../../api/ruas.js'
import DistrictMap from '../DistrictMap.jsx'
import styles from './AjustesRotasPanel.module.css'

const PASSOS = ['Selecionar Ruas', 'Alterar Distrito', 'Confirmar']
const RUAS_POR_PAGINA = 12

export default function AjustesRotasPanel({ distritoOrigem, onFechar, onConcluido }) {
  const [passo, setPasso] = useState(0)
  const [ruas, setRuas] = useState([])
  const [carregando, setCarregando] = useState(true)
  const [busca, setBusca] = useState('')
  const [pagina, setPagina] = useState(1)
  const [selecionadas, setSelecionadas] = useState(new Set())
  const [distritoDestino, setDistritoDestino] = useState('')
  const [carteiroResponsavel, setCarteiroResponsavel] = useState('')
  const [motivo, setMotivo] = useState('')
  const [resultado, setResultado] = useState(null)
  const [salvando, setSalvando] = useState(false)
  const [erroSalvar, setErroSalvar] = useState('')

  useEffect(() => {
    let ativo = true
    setCarregando(true)
    listarRuas({ distrito: distritoOrigem || '' }).then((dados) => {
      if (ativo) {
        setRuas(dados)
        setCarregando(false)
      }
    })
    return () => {
      ativo = false
    }
  }, [distritoOrigem])

  useEffect(() => {
    setPagina(1)
  }, [busca, distritoOrigem])

  const ruasFiltradas = ruas.filter((r) => r.nome_rua.toLowerCase().includes(busca.toLowerCase()))
  const totalPaginas = Math.max(1, Math.ceil(ruasFiltradas.length / RUAS_POR_PAGINA))
  const paginaAtual = Math.min(pagina, totalPaginas)
  const ruasDaPagina = ruasFiltradas.slice(
    (paginaAtual - 1) * RUAS_POR_PAGINA,
    paginaAtual * RUAS_POR_PAGINA
  )

  function alternarSelecao(id) {
    setSelecionadas((prev) => {
      const proximo = new Set(prev)
      proximo.has(id) ? proximo.delete(id) : proximo.add(id)
      return proximo
    })
  }

  async function confirmarMudanca() {
    const ids = [...selecionadas]
    setErroSalvar('')
    setSalvando(true)
    try {
      await moverRuasEmLote(ids, distritoDestino)
      setResultado({
        quantidade: ids.length,
        de: distritoOrigem,
        para: distritoDestino,
      })
      setPasso(2)
    } catch (err) {
      // Antes esse erro era engolido silenciosamente (atualizarDistritoDaRua
      // devolvia null em vez de lançar) — a tela ia direto pro "sucesso" mesmo
      // quando a API recusava com 401/403. Agora mostra o erro de verdade.
      setErroSalvar(err.message)
    } finally {
      setSalvando(false)
    }
  }

  function finalizar() {
    onConcluido?.(resultado)
  }

  return (
    <aside className={styles.painel} aria-label="Ajustes de Rotas">
      <header className={styles.cabecalho}>
        <h2>Ajustes de Rotas</h2>
        <button type="button" className={styles.fechar} onClick={onFechar} aria-label="Fechar painel">
          ✕
        </button>
      </header>

      <ol className={styles.passos}>
        {PASSOS.map((nome, i) => (
          <li key={nome} data-estado={i === passo ? 'ativo' : i < passo ? 'concluido' : 'pendente'}>
            <span className={styles.numeroPasso}>{i + 1}</span> {nome}
          </li>
        ))}
      </ol>

      {passo === 0 && (
        <div className={styles.conteudo}>
          <p className={styles.aviso}>Selecione as ruas que serão movidas para outro distrito.</p>
          <input
            type="search"
            className={styles.campoBusca}
            placeholder="Pesquisar rua..."
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
          />

          {carregando ? (
            <p className={styles.carregando}>Carregando ruas…</p>
          ) : (
            <table className={styles.tabelaSelecao}>
              <thead>
                <tr>
                  <th aria-hidden="true"></th>
                  <th>Rua</th>
                  <th>Distrito Atual</th>
                </tr>
              </thead>
              <tbody>
                {ruasDaPagina.map((r) => (
                  <tr key={r.id}>
                    <td>
                      <input
                        type="checkbox"
                        checked={selecionadas.has(r.id)}
                        onChange={() => alternarSelecao(r.id)}
                        aria-label={`Selecionar ${r.nome_rua}`}
                      />
                    </td>
                    <td>{r.nome_rua}</td>
                    <td>{r.distrito}</td>
                  </tr>
                ))}
                {ruasFiltradas.length === 0 && (
                  <tr>
                    <td colSpan={3} className={styles.vazio}>Nenhuma rua encontrada.</td>
                  </tr>
                )}
              </tbody>
            </table>
          )}

          {ruasFiltradas.length > 0 && (
            <div className={styles.paginacao}>
              <button
                type="button"
                disabled={paginaAtual <= 1}
                onClick={() => setPagina((p) => p - 1)}
              >
                ← Anterior
              </button>
              <span>Página {paginaAtual} de {totalPaginas}</span>
              <button
                type="button"
                disabled={paginaAtual >= totalPaginas}
                onClick={() => setPagina((p) => p + 1)}
              >
                Próxima →
              </button>
            </div>
          )}

          <footer className={styles.rodapePasso}>
            <span>{selecionadas.size} rua(s) selecionada(s)</span>
            <button
              type="button"
              className={styles.botaoPrimario}
              disabled={selecionadas.size === 0}
              onClick={() => setPasso(1)}
            >
              Próximo →
            </button>
          </footer>
        </div>
      )}

      {passo === 1 && (
        <div className={styles.conteudo}>
          <label className={styles.campo}>
            Mover para o distrito
            <select value={distritoDestino} onChange={(e) => setDistritoDestino(e.target.value)}>
              <option value="">Selecione…</option>
              {DISTRITOS.map((d) => (
                <option key={d.numero} value={d.numero}>{d.numero}</option>
              ))}
            </select>
          </label>

          <label className={styles.campo}>
            Carteiro responsável
            <input
              type="text"
              value={carteiroResponsavel}
              onChange={(e) => setCarteiroResponsavel(e.target.value)}
              placeholder="Nome do carteiro"
            />
          </label>

          <label className={styles.campo}>
            Motivo (opcional)
            <input
              type="text"
              value={motivo}
              onChange={(e) => setMotivo(e.target.value)}
              placeholder="Ex.: ajuste de carga operacional"
            />
          </label>

          {distritoOrigem && distritoDestino && (
            <div className={styles.comparativo}>
              <DistrictMap
                tamanho="pequeno"
                emFoco={[distritoOrigem, distritoDestino]}
                seta={{ de: distritoOrigem, para: distritoDestino }}
              />
            </div>
          )}

          {erroSalvar && <p className={styles.erro}>{erroSalvar}</p>}

          <footer className={styles.rodapePasso}>
            <button type="button" className={styles.botaoSecundario} onClick={() => setPasso(0)} disabled={salvando}>
              ← Voltar
            </button>
            <button
              type="button"
              className={styles.botaoPrimario}
              disabled={!distritoDestino || salvando}
              onClick={confirmarMudanca}
            >
              {salvando ? 'Salvando…' : '✓ Confirmar'}
            </button>
          </footer>
        </div>
      )}

      {passo === 2 && resultado && (
        <div className={styles.conteudo}>
          <div className={styles.sucesso}>
            <strong>Alteração realizada com sucesso!</strong>
            <p>
              {resultado.quantidade} rua(s) foram movidas do distrito {resultado.de || '—'} para o
              distrito {resultado.para}.
            </p>
          </div>
          <footer className={styles.rodapePasso}>
            <button type="button" className={styles.botaoPrimario} onClick={finalizar}>
              Concluir
            </button>
          </footer>
        </div>
      )}
    </aside>
  )
}
