import { useState } from 'react'
import { criarRua } from '../api/ruas.js'
import { useDistritosAtivos } from '../hooks/useDistritosAtivos.js'
import styles from './NovaRuaModal.module.css'

export default function NovaRuaModal({ aberto, onFechar, onCriada }) {
  const { distritosAtivos } = useDistritosAtivos()
  const [nomeRua, setNomeRua] = useState('')
  const [cep, setCep] = useState('')
  const [distrito, setDistrito] = useState('')
  const [bairro, setBairro] = useState('')
  const [rota, setRota] = useState('')
  const [observacao, setObservacao] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [erro, setErro] = useState('')

  if (!aberto) return null

  function limparEfechar() {
    setNomeRua('')
    setCep('')
    setDistrito('')
    setBairro('')
    setRota('')
    setObservacao('')
    setErro('')
    onFechar()
  }

  async function salvar(e) {
    e.preventDefault()
    setErro('')

    if (!nomeRua.trim() || !cep.trim() || !distrito) {
      setErro('Nome da rua, CEP e distrito são obrigatórios.')
      return
    }

    setSalvando(true)
    try {
      const nova = await criarRua({ nomeRua: nomeRua.trim(), cep: cep.trim(), distrito, bairro, rota, observacao })
      onCriada(nova)
      limparEfechar()
    } catch (err) {
      setErro(err.message)
    } finally {
      setSalvando(false)
    }
  }

  return (
    <div className={styles.fundo} onClick={limparEfechar}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Cadastrar nova rua">
        <header className={styles.cabecalho}>
          <h2>Nova rua</h2>
          <button type="button" className={styles.fechar} onClick={limparEfechar} aria-label="Fechar">✕</button>
        </header>

        <form className={styles.form} onSubmit={salvar}>
          <label className={styles.campo}>
            Nome da rua *
            <input type="text" value={nomeRua} onChange={(e) => setNomeRua(e.target.value)} placeholder="Ex: Rua das Acácias" />
          </label>

          <label className={styles.campo}>
            CEP *
            <input type="text" value={cep} onChange={(e) => setCep(e.target.value)} placeholder="28000-000" />
          </label>

          <label className={styles.campo}>
            Distrito *
            <select value={distrito} onChange={(e) => setDistrito(e.target.value)}>
              <option value="">Selecione…</option>
              {distritosAtivos.map((d) => (
                <option key={d.numero} value={d.numero}>{d.numero}</option>
              ))}
            </select>
          </label>

          <label className={styles.campo}>
            Bairro
            <input type="text" value={bairro} onChange={(e) => setBairro(e.target.value)} />
          </label>

          <label className={styles.campo}>
            Carteiro/rota (opcional)
            <input type="text" value={rota} onChange={(e) => setRota(e.target.value)} />
          </label>

          <label className={styles.campo}>
            Observação (opcional)
            <input type="text" value={observacao} onChange={(e) => setObservacao(e.target.value)} />
          </label>

          {erro && <p className={styles.erro}>{erro}</p>}

          <footer className={styles.rodape}>
            <button type="button" className={styles.botaoSecundario} onClick={limparEfechar}>
              Cancelar
            </button>
            <button type="submit" className={styles.botaoPrimario} disabled={salvando}>
              {salvando ? 'Salvando…' : 'Cadastrar'}
            </button>
          </footer>
        </form>
      </div>
    </div>
  )
}
