import { useEffect, useState } from 'react'
import { listarObservacoes, adicionarObservacao, excluirObservacao, rotuloCategoria, CATEGORIAS } from '../api/observacoes.js'
import { useAuth } from '../context/AuthContext.jsx'
import ConfirmModal from './ConfirmModal.jsx'
import styles from './ObservacoesRuaModal.module.css'

export default function ObservacoesRuaModal({ rua, onFechar }) {
  const { admin } = useAuth()
  const [observacoes, setObservacoes] = useState([])
  const [carregando, setCarregando] = useState(true)
  const [erro, setErro] = useState('')
  const [formAberto, setFormAberto] = useState(false)
  const [categoria, setCategoria] = useState(CATEGORIAS[0])
  const [texto, setTexto] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [paraExcluir, setParaExcluir] = useState(null)
  const [excluindoId, setExcluindoId] = useState(null)

  async function carregar() {
    setCarregando(true)
    setErro('')
    try {
      setObservacoes(await listarObservacoes(rua.id))
    } catch (err) {
      setErro(err.message)
    } finally {
      setCarregando(false)
    }
  }

  useEffect(() => {
    carregar()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [rua.id])

  async function salvar(e) {
    e.preventDefault()
    if (!texto.trim()) return
    setSalvando(true)
    try {
      await adicionarObservacao(rua.id, { categoria, texto: texto.trim() })
      setTexto('')
      setFormAberto(false)
      await carregar()
    } catch (err) {
      alert(err.message)
    } finally {
      setSalvando(false)
    }
  }

  async function confirmarExclusao() {
    if (!paraExcluir) return
    setExcluindoId(paraExcluir.id)
    try {
      await excluirObservacao(paraExcluir.id)
      setParaExcluir(null)
      await carregar()
    } catch (err) {
      alert(err.message)
    } finally {
      setExcluindoId(null)
    }
  }

  return (
    <div className={styles.fundo} onClick={onFechar}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Observações da rua">
        <header className={styles.cabecalho}>
          <div>
            <h2>Observações de campo</h2>
            <span className={styles.nomeRua}>{rua.nome_rua}</span>
          </div>
          <button type="button" className={styles.fechar} onClick={onFechar} aria-label="Fechar">✕</button>
        </header>

        {carregando && <p className={styles.mensagem}>Carregando…</p>}
        {erro && <p className={styles.mensagemErro}>{erro}</p>}

        {!carregando && !erro && (
          <>
            {observacoes.length === 0 ? (
              <p className={styles.mensagem}>Nenhuma observação registrada ainda pra essa rua.</p>
            ) : (
              <ul className={styles.lista}>
                {observacoes.map((o) => (
                  <li key={o.id} className={styles.item}>
                    <span className={styles.categoriaSelo} data-categoria={o.categoria}>
                      {rotuloCategoria(o.categoria)}
                    </span>
                    <div className={styles.itemInfo}>
                      <p>{o.texto}</p>
                      <span className={styles.itemMeta}>Por {o.criado_por || '—'}</span>
                    </div>
                    {admin && (
                      <button
                        type="button"
                        className={styles.btnExcluir}
                        onClick={() => setParaExcluir(o)}
                        disabled={excluindoId === o.id}
                        aria-label="Excluir observação"
                      >
                        {excluindoId === o.id ? '…' : '🗑️'}
                      </button>
                    )}
                  </li>
                ))}
              </ul>
            )}

            {admin && !formAberto && (
              <button type="button" className={styles.btnNovo} onClick={() => setFormAberto(true)}>
                + Adicionar observação
              </button>
            )}

            {admin && formAberto && (
              <form className={styles.form} onSubmit={salvar}>
                <label className={styles.campo}>
                  Categoria
                  <select value={categoria} onChange={(e) => setCategoria(e.target.value)}>
                    {CATEGORIAS.map((c) => (
                      <option key={c} value={c}>{rotuloCategoria(c)}</option>
                    ))}
                  </select>
                </label>
                <label className={styles.campo}>
                  Texto
                  <textarea
                    value={texto}
                    onChange={(e) => setTexto(e.target.value)}
                    placeholder="Ex: rua sem saída, entrada só pela Av. Principal"
                    rows={3}
                  />
                </label>
                <div className={styles.rodapeForm}>
                  <button type="button" className={styles.botaoSecundario} onClick={() => setFormAberto(false)}>
                    Cancelar
                  </button>
                  <button type="submit" className={styles.botaoPrimario} disabled={salvando || !texto.trim()}>
                    {salvando ? 'Salvando…' : 'Salvar'}
                  </button>
                </div>
              </form>
            )}
          </>
        )}
      </div>

      <ConfirmModal
        aberto={!!paraExcluir}
        titulo="Excluir observação"
        mensagem={paraExcluir ? `Excluir a observação "${paraExcluir.texto}"?` : ''}
        onConfirmar={confirmarExclusao}
        onCancelar={() => setParaExcluir(null)}
        confirmando={excluindoId === paraExcluir?.id}
      />
    </div>
  )
}
