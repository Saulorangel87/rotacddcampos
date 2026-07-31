import { useEffect, useState } from 'react'
import { listarColaboradores, criarColaborador } from '../api/colaboradores.js'
import styles from './ColaboradoresModal.module.css'

export default function ColaboradoresModal({ aberto, onFechar }) {
  const [colaboradores, setColaboradores] = useState([])
  const [carregando, setCarregando] = useState(true)
  const [busca, setBusca] = useState('')
  const [formAberto, setFormAberto] = useState(false)

  useEffect(() => {
    if (!aberto) return
    carregar()
  }, [aberto])

  function carregar() {
    setCarregando(true)
    listarColaboradores().then((dados) => {
      setColaboradores(dados)
      setCarregando(false)
    })
  }

  function fecharTudo() {
    setFormAberto(false)
    onFechar()
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
              <button type="button" className={styles.btnNovo} onClick={() => setFormAberto(true)}>
                + Novo
              </button>
            </div>

            {carregando ? (
              <p className={styles.mensagem}>Carregando…</p>
            ) : filtrados.length === 0 ? (
              <p className={styles.mensagem}>Nenhum colaborador encontrado.</p>
            ) : (
              <ul className={styles.lista}>
                {filtrados.map((c) => (
                  <li key={c.id}>
                    <span className={styles.nome}>{c.nome}</span>
                    <span className={styles.matricula}>{c.matricula}</span>
                  </li>
                ))}
              </ul>
            )}

            <footer className={styles.rodape}>{filtrados.length} colaborador(es)</footer>
          </>
        )}
      </div>
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
        <input type="text" value={funcao} onChange={(e) => setFuncao(e.target.value)} placeholder="Ex: MOTORIZADO (M), CICLISTA..." />
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
