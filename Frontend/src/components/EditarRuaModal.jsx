import { useState } from 'react'
import { atualizarRua } from '../api/ruas.js'
import { DISTRITOS } from '../data/distritos.js'
import styles from './NovaRuaModal.module.css'

export default function EditarRuaModal({ rua, onFechar, onAtualizada }) {
  const [nomeRua, setNomeRua] = useState(rua.nome_rua || '')
  const [cep, setCep] = useState(rua.cep || '')
  const [distrito, setDistrito] = useState(rua.distrito || '')
  const [bairro, setBairro] = useState(rua.bairro || '')
  const [rota, setRota] = useState(rua.rota || '')
  const [observacao, setObservacao] = useState(rua.observacao || '')
  const [salvando, setSalvando] = useState(false)
  const [erro, setErro] = useState('')

  async function salvar(e) {
    e.preventDefault()
    setErro('')

    if (!nomeRua.trim() || !cep.trim() || !distrito) {
      setErro('Nome da rua, CEP e distrito são obrigatórios.')
      return
    }

    setSalvando(true)
    try {
      const atualizada = await atualizarRua(rua.id, {
        nome_rua: nomeRua.trim(),
        cep: cep.trim(),
        distrito,
        bairro,
        rota,
        observacao,
      })
      onAtualizada(atualizada)
      onFechar()
    } catch (err) {
      setErro(err.message)
    } finally {
      setSalvando(false)
    }
  }

  return (
    <div className={styles.fundo} onClick={onFechar}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Editar rua">
        <header className={styles.cabecalho}>
          <h2>Editar rua</h2>
          <button type="button" className={styles.fechar} onClick={onFechar} aria-label="Fechar">✕</button>
        </header>

        <form className={styles.form} onSubmit={salvar}>
          <label className={styles.campo}>
            Nome da rua *
            <input type="text" value={nomeRua} onChange={(e) => setNomeRua(e.target.value)} />
          </label>

          <label className={styles.campo}>
            CEP *
            <input type="text" value={cep} onChange={(e) => setCep(e.target.value)} />
          </label>

          <label className={styles.campo}>
            Distrito *
            <select value={distrito} onChange={(e) => setDistrito(e.target.value)}>
              <option value="">Selecione…</option>
              {DISTRITOS.map((d) => (
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
            <button type="button" className={styles.botaoSecundario} onClick={onFechar}>
              Cancelar
            </button>
            <button type="submit" className={styles.botaoPrimario} disabled={salvando}>
              {salvando ? 'Salvando…' : 'Salvar alterações'}
            </button>
          </footer>
        </form>
      </div>
    </div>
  )
}
