import { useState } from 'react'
import { consultarSaldoFolgas, lancarFolga, excluirLancamentoFolga } from '../api/folgas.js'
import { useAuth } from '../context/AuthContext.jsx'
import ConfirmModal from './ConfirmModal.jsx'
import styles from './FolgasModal.module.css'

export default function FolgasModal({ aberto, onFechar, onAlterado }) {
  const { admin } = useAuth()
  const [matricula, setMatricula] = useState('')
  const [consultando, setConsultando] = useState(false)
  const [erro, setErro] = useState('')
  const [resultado, setResultado] = useState(null)
  const [formAberto, setFormAberto] = useState(false)
  const [lancamentoParaExcluir, setLancamentoParaExcluir] = useState(null)
  const [excluindoId, setExcluindoId] = useState(null)

  function fecharTudo() {
    setMatricula('')
    setResultado(null)
    setErro('')
    setFormAberto(false)
    onFechar()
  }

  async function pesquisar(e) {
    e.preventDefault()
    const m = matricula.trim()
    if (!m) return

    setConsultando(true)
    setErro('')
    setResultado(null)
    try {
      const dados = await consultarSaldoFolgas(m)
      setResultado(dados)
    } catch (err) {
      setErro(err.message)
    } finally {
      setConsultando(false)
    }
  }

  function aoExcluir(lancamento) {
    setLancamentoParaExcluir(lancamento)
  }

  async function confirmarExclusao() {
    if (!lancamentoParaExcluir) return
    setExcluindoId(lancamentoParaExcluir.id)
    try {
      await excluirLancamentoFolga(lancamentoParaExcluir.id)
      setLancamentoParaExcluir(null)
      // Refaz a consulta pra recalcular o saldo certo, sem confiar em subtração local
      const dados = await consultarSaldoFolgas(resultado.matricula)
      setResultado(dados)
      onAlterado?.()
    } catch (err) {
      alert(err.message)
    } finally {
      setExcluindoId(null)
    }
  }

  if (!aberto) return null

  return (
    <div className={styles.fundo} onClick={fecharTudo}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Consulta de folgas">
        <header className={styles.cabecalho}>
          <h2>{formAberto ? 'Lançar folga' : 'Consulta de folgas'}</h2>
          <button type="button" className={styles.fechar} onClick={fecharTudo} aria-label="Fechar">✕</button>
        </header>

        {formAberto ? (
          <FormularioLancamento
            matriculaInicial={resultado?.matricula || matricula}
            onCancelar={() => setFormAberto(false)}
            onLancado={async (matriculaLancada) => {
              setFormAberto(false)
              const dados = await consultarSaldoFolgas(matriculaLancada)
              setMatricula(matriculaLancada)
              setResultado(dados)
              onAlterado?.()
            }}
          />
        ) : (
          <>
            <form className={styles.barraBusca} onSubmit={pesquisar}>
              <input
                type="text"
                inputMode="numeric"
                className={styles.campoBusca}
                placeholder="Digite a matrícula..."
                value={matricula}
                onChange={(e) => setMatricula(e.target.value)}
              />
              <button type="submit" className={styles.btnBuscar} disabled={consultando || !matricula.trim()}>
                {consultando ? 'Buscando…' : 'Buscar'}
              </button>
            </form>

            {erro && <p className={styles.mensagemErro}>{erro}</p>}

            {resultado && (
              <>
                <div className={styles.saldoCard}>
                  <div>
                    <strong className={styles.nome}>{resultado.nome}</strong>
                    <span className={styles.matriculaTexto}>Matrícula {resultado.matricula}</span>
                  </div>
                  <div className={styles.saldoValor} data-negativo={resultado.saldo < 0}>
                    {resultado.saldo}
                    <span>folga(s)</span>
                  </div>
                </div>

                {admin && (
                  <button type="button" className={styles.btnNovo} onClick={() => setFormAberto(true)}>
                    + Lançar folga
                  </button>
                )}

                {resultado.lancamentos.length === 0 ? (
                  <p className={styles.mensagem}>Nenhum lançamento registrado ainda.</p>
                ) : (
                  <ul className={styles.extrato}>
                    {resultado.lancamentos.map((l) => (
                      <li key={l.id} className={styles.item}>
                        <span className={styles.tipoSelo} data-tipo={l.tipo}>
                          {l.tipo === 'credito' ? `+${l.quantidade}` : `−${l.quantidade}`}
                        </span>
                        <div className={styles.itemInfo}>
                          <span className={styles.motivo}>{l.motivo}</span>
                          <span className={styles.itemMeta}>
                            {l.data_referencia && `${formatarData(l.data_referencia)} · `}
                            lançado por {l.criado_por || '—'}
                          </span>
                        </div>
                        {admin && (
                          <button
                            type="button"
                            className={styles.btnExcluir}
                            onClick={() => aoExcluir(l)}
                            disabled={excluindoId === l.id}
                            aria-label="Excluir lançamento"
                            title="Excluir lançamento"
                          >
                            {excluindoId === l.id ? '…' : '🗑️'}
                          </button>
                        )}
                      </li>
                    ))}
                  </ul>
                )}
              </>
            )}

            {!resultado && !erro && (
              <p className={styles.mensagem}>Digite a matrícula pra ver o saldo e o histórico de folgas.</p>
            )}
          </>
        )}
      </div>

      <ConfirmModal
        aberto={!!lancamentoParaExcluir}
        titulo="Excluir lançamento"
        mensagem={lancamentoParaExcluir ? `Excluir o lançamento "${lancamentoParaExcluir.motivo}"? Essa ação não pode ser desfeita.` : ''}
        onConfirmar={confirmarExclusao}
        onCancelar={() => setLancamentoParaExcluir(null)}
        confirmando={excluindoId === lancamentoParaExcluir?.id}
      />
    </div>
  )
}

function formatarData(dataIso) {
  // A API devolve data_referencia em ISO (AAAA-MM-DDTHH:mm:ssZ); exibe em DD/MM/AAAA.
  const d = new Date(dataIso)
  if (Number.isNaN(d.getTime())) return dataIso
  return d.toLocaleDateString('pt-BR', { timeZone: 'UTC' })
}

function FormularioLancamento({ matriculaInicial, onCancelar, onLancado }) {
  const [matricula, setMatricula] = useState(matriculaInicial || '')
  const [tipo, setTipo] = useState('credito')
  const [quantidade, setQuantidade] = useState('1')
  const [motivo, setMotivo] = useState('')
  const [dataReferencia, setDataReferencia] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [erro, setErro] = useState('')

  async function salvar(e) {
    e.preventDefault()
    setErro('')

    const m = matricula.trim()
    const qtd = Number(quantidade)
    if (!m || !motivo.trim()) {
      setErro('Matrícula e motivo/justificativa são obrigatórios.')
      return
    }
    if (!qtd || qtd <= 0) {
      setErro('Quantidade deve ser maior que zero.')
      return
    }

    setSalvando(true)
    try {
      await lancarFolga({ matricula: m, tipo, quantidade: qtd, motivo: motivo.trim(), dataReferencia })
      onLancado(m)
    } catch (err) {
      setErro(err.message)
    } finally {
      setSalvando(false)
    }
  }

  return (
    <form className={styles.form} onSubmit={salvar}>
      <label className={styles.campo}>
        Matrícula *
        <input type="text" value={matricula} onChange={(e) => setMatricula(e.target.value)} />
      </label>

      <label className={styles.campo}>
        Tipo *
        <select value={tipo} onChange={(e) => setTipo(e.target.value)}>
          <option value="credito">Crédito (folga ganha)</option>
          <option value="debito">Débito (folga tirada)</option>
        </select>
      </label>

      <label className={styles.campo}>
        Quantidade *
        <input type="number" min="1" value={quantidade} onChange={(e) => setQuantidade(e.target.value)} />
      </label>

      <label className={styles.campo}>
        Motivo / justificativa *
        <input
          type="text"
          value={motivo}
          onChange={(e) => setMotivo(e.target.value)}
          placeholder={tipo === 'credito' ? 'Ex: Feriado trabalhado 25/12' : 'Ex: Folga tirada'}
        />
      </label>

      <label className={styles.campo}>
        Data de referência (opcional)
        <input
          type="text"
          value={dataReferencia}
          onChange={(e) => setDataReferencia(e.target.value)}
          placeholder="DD/MM/AAAA"
        />
      </label>

      {erro && <p className={styles.erro}>{erro}</p>}

      <footer className={styles.rodapeForm}>
        <button type="button" className={styles.botaoSecundario} onClick={onCancelar}>
          ← Voltar
        </button>
        <button type="submit" className={styles.botaoPrimario} disabled={salvando}>
          {salvando ? 'Salvando…' : 'Lançar'}
        </button>
      </footer>
    </form>
  )
}
