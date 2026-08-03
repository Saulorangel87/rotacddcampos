import { useEffect, useState } from 'react'
import { listarColaboradores, criarColaborador, excluirColaborador } from '../api/colaboradores.js'
import { FUNCOES } from '../data/funcoes.js'
import { useAuth } from '../context/AuthContext.jsx'
import ConfirmModal from './ConfirmModal.jsx'
import styles from './ColaboradoresModal.module.css'

export default function ColaboradoresModal({ aberto, onFechar, onAlterado }) {
  const { admin } = useAuth()
  const [colaboradores, setColaboradores] = useState([])
  const [carregando, setCarregando] = useState(true)
  const [erroCarregar, setErroCarregar] = useState('')
  const [busca, setBusca] = useState('')
  const [formAberto, setFormAberto] = useState(false)
  const [excluindoId, setExcluindoId] = useState(null)
  const [colaboradorParaExcluir, setColaboradorParaExcluir] = useState(null)

  useEffect(() => {
    if (!aberto) return
    carregar()
  }, [aberto])

  function carregar() {
    setCarregando(true)
    setErroCarregar('')
    listarColaboradores()
      .then(setColaboradores)
      .catch((err) => setErroCarregar(err.message))
      .finally(() => setCarregando(false))
  }

  function fecharTudo() {
    setFormAberto(false)
    onFechar()
  }

  function aoExcluir(colaborador) {
    setColaboradorParaExcluir(colaborador)
  }

  async function confirmarExclusao() {
    if (!colaboradorParaExcluir) return
    setExcluindoId(colaboradorParaExcluir.id)
    try {
      await excluirColaborador(colaboradorParaExcluir.id)
      setColaboradores((prev) => prev.filter((c) => c.id !== colaboradorParaExcluir.id))
      setColaboradorParaExcluir(null)
      onAlterado?.()
    } catch (err) {
      alert(err.message)
    } finally {
      setExcluindoId(null)
    }
  }

  if (!aberto) return null

  const filtrados = colaboradores.filter((c) => c.nome.toLowerCase().includes(busca.toLowerCase()))

  return (
    <div className={styles.fundo} onClick={fecharTudo}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Colaboradores da unidade">
        <header className={styles.cabecalho}>
          <h2>{formAberto ? 'Novo colaborador' : 'Colaboradores'}</h2>
          <button type="button" className={styles.fechar} onClick={fecharTudo} aria-label="Fechar">✕</button>
        </header>

        {formAberto ? (
          <FormularioColaborador
            onCancelar={() => setFormAberto(false)}
            onCriado={() => {
              setFormAberto(false)
              carregar()
              onAlterado?.()
            }}
          />
        ) : (
          <>
            <div className={styles.barraTopo}>
              <input
                type="search"
                className={styles.campoBusca}
                placeholder="Pesquisar colaborador..."
                value={busca}
                onChange={(e) => setBusca(e.target.value)}
              />
              {admin && (
                <button type="button" className={styles.btnNovo} onClick={() => setFormAberto(true)}>
                  + Novo
                </button>
              )}
            </div>

            {carregando ? (
              <p className={styles.mensagem}>Carregando…</p>
            ) : erroCarregar ? (
              <p className={styles.mensagem}>{erroCarregar}</p>
            ) : filtrados.length === 0 ? (
              <p className={styles.mensagem}>Nenhum colaborador encontrado.</p>
            ) : (
              <ul className={styles.lista}>
                {filtrados.map((c) => (
                  <li key={c.id}>
                    <div className={styles.itemInfo}>
                      <span className={styles.nome}>{c.nome}</span>
                      <span className={styles.matricula}>{c.matricula}</span>
                    </div>
                    <button
                      type="button"
                      className={styles.btnExcluir}
                      onClick={() => aoExcluir(c)}
                      disabled={excluindoId === c.id}
                      aria-label={`Excluir ${c.nome}`}
                      title="Excluir colaborador"
                      hidden={!admin}
                    >
                      {excluindoId === c.id ? '…' : '🗑️'}
                    </button>
                  </li>
                ))}
              </ul>
            )}

            <footer className={styles.rodape}>{filtrados.length} colaborador(es)</footer>
          </>
        )}
      </div>

      <ConfirmModal
        aberto={!!colaboradorParaExcluir}
        titulo="Excluir colaborador"
        mensagem={colaboradorParaExcluir ? `Tem certeza que quer excluir "${colaboradorParaExcluir.nome}" (matrícula ${colaboradorParaExcluir.matricula})? Essa ação não pode ser desfeita.` : ''}
        onConfirmar={confirmarExclusao}
        onCancelar={() => setColaboradorParaExcluir(null)}
        confirmando={excluindoId === colaboradorParaExcluir?.id}
      />
    </div>
  )
}

function FormularioColaborador({ onCancelar, onCriado }) {
  const [nome, setNome] = useState('')
  const [matricula, setMatricula] = useState('')
  const [funcao, setFuncao] = useState('')
  const [cargo, setCargo] = useState('')
  const [dataAdmissao, setDataAdmissao] = useState('')
  const [dataNascimento, setDataNascimento] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [erro, setErro] = useState('')

  async function salvar(e) {
    e.preventDefault()
    setErro('')

    if (!nome.trim() || !matricula.trim()) {
      setErro('Nome e matrícula são obrigatórios.')
      return
    }

    setSalvando(true)
    try {
      await criarColaborador({ nome: nome.trim(), matricula: matricula.trim(), funcao, cargo, dataAdmissao, dataNascimento })
      onCriado()
    } catch (err) {
      setErro(err.message)
    } finally {
      setSalvando(false)
    }
  }

  return (
    <form className={styles.form} onSubmit={salvar}>
      <label className={styles.campo}>
        Nome *
        <input type="text" value={nome} onChange={(e) => setNome(e.target.value)} />
      </label>

      <label className={styles.campo}>
        Matrícula *
        <input type="text" value={matricula} onChange={(e) => setMatricula(e.target.value)} />
      </label>

      <label className={styles.campo}>
        Função
        <select value={funcao} onChange={(e) => setFuncao(e.target.value)}>
          <option value="">Selecione…</option>
          {FUNCOES.map((f) => (
            <option key={f} value={f}>{f}</option>
          ))}
        </select>
      </label>

      <label className={styles.campo}>
        Cargo
        <input type="text" value={cargo} onChange={(e) => setCargo(e.target.value)} placeholder="Ex: AGENTE DE CORREIOS" />
      </label>

      <label className={styles.campo}>
        Data de admissão (opcional)
        <input type="text" value={dataAdmissao} onChange={(e) => setDataAdmissao(e.target.value)} placeholder="DD/MM/AAAA" />
      </label>

      <label className={styles.campo}>
        Data de nascimento (opcional)
        <input type="text" value={dataNascimento} onChange={(e) => setDataNascimento(e.target.value)} placeholder="DD/MM/AAAA" />
      </label>

      {erro && <p className={styles.erro}>{erro}</p>}

      <footer className={styles.rodapeForm}>
        <button type="button" className={styles.botaoSecundario} onClick={onCancelar}>
          ← Voltar
        </button>
        <button type="submit" className={styles.botaoPrimario} disabled={salvando}>
          {salvando ? 'Salvando…' : 'Cadastrar'}
        </button>
      </footer>
    </form>
  )
}
