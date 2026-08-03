import { useEffect, useState } from 'react'
import { listarUsuarios, criarUsuario, resetarSenhaDeUsuario } from '../api/usuarios.js'
import styles from './ColaboradoresModal.module.css'

const VISAO_LISTA = 'lista'
const VISAO_NOVO = 'novo'
const VISAO_RESETAR = 'resetar'

export default function UsuariosModal({ aberto, onFechar }) {
  const [usuarios, setUsuarios] = useState([])
  const [carregando, setCarregando] = useState(true)
  const [erroCarregar, setErroCarregar] = useState('')
  const [visao, setVisao] = useState(VISAO_LISTA)
  const [usuarioSelecionado, setUsuarioSelecionado] = useState(null)

  useEffect(() => {
    if (!aberto) return
    setVisao(VISAO_LISTA)
    carregar()
  }, [aberto])

  function carregar() {
    setCarregando(true)
    setErroCarregar('')
    listarUsuarios()
      .then(setUsuarios)
      .catch((err) => setErroCarregar(err.message))
      .finally(() => setCarregando(false))
  }

  function fecharTudo() {
    setVisao(VISAO_LISTA)
    onFechar()
  }

  function abrirReset(usuario) {
    setUsuarioSelecionado(usuario)
    setVisao(VISAO_RESETAR)
  }

  if (!aberto) return null

  const titulos = {
    [VISAO_LISTA]: 'Gerenciar usuários',
    [VISAO_NOVO]: 'Novo usuário',
    [VISAO_RESETAR]: `Resetar senha — ${usuarioSelecionado?.matricula ?? ''}`,
  }

  return (
    <div className={styles.fundo} onClick={fecharTudo}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Gerenciar usuários">
        <header className={styles.cabecalho}>
          <h2>{titulos[visao]}</h2>
          <button type="button" className={styles.fechar} onClick={fecharTudo} aria-label="Fechar">✕</button>
        </header>

        {visao === VISAO_NOVO && (
          <FormularioNovoUsuario
            onCancelar={() => setVisao(VISAO_LISTA)}
            onCriado={() => {
              setVisao(VISAO_LISTA)
              carregar()
            }}
          />
        )}

        {visao === VISAO_RESETAR && (
          <FormularioResetarSenha
            usuario={usuarioSelecionado}
            onCancelar={() => setVisao(VISAO_LISTA)}
            onResetado={() => {
              setVisao(VISAO_LISTA)
              carregar()
            }}
          />
        )}

        {visao === VISAO_LISTA && (
          <>
            <div className={styles.barraTopo}>
              <p className={styles.mensagemInline}>
                {usuarios.length} usuário(s) com acesso ao sistema
              </p>
              <button type="button" className={styles.btnNovo} onClick={() => setVisao(VISAO_NOVO)}>
                + Novo
              </button>
            </div>

            {carregando ? (
              <p className={styles.mensagem}>Carregando…</p>
            ) : erroCarregar ? (
              <p className={styles.mensagem}>{erroCarregar}</p>
            ) : usuarios.length === 0 ? (
              <p className={styles.mensagem}>Nenhum usuário cadastrado ainda.</p>
            ) : (
              <ul className={styles.lista}>
                {usuarios.map((u) => (
                  <li key={u.id}>
                    <div className={styles.itemInfo}>
                      <span className={styles.nome}>
                        {u.matricula}{' '}
                        {u.papel === 'admin' && <span className={styles.seloAdmin}>admin</span>}
                      </span>
                      <span className={styles.matricula}>
                        {u.bloqueado ? 'Bloqueado temporariamente' : u.senha_provisoria ? 'Senha provisória (ainda não trocou)' : 'Ativo'}
                      </span>
                    </div>
                    <button
                      type="button"
                      className={styles.btnExcluir}
                      onClick={() => abrirReset(u)}
                      aria-label={`Resetar senha de ${u.matricula}`}
                      title="Resetar senha"
                    >
                      🔑
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </>
        )}
      </div>
    </div>
  )
}

function FormularioNovoUsuario({ onCancelar, onCriado }) {
  const [matricula, setMatricula] = useState('')
  const [senhaTemporaria, setSenhaTemporaria] = useState('')
  const [papel, setPapel] = useState('colaborador')
  const [salvando, setSalvando] = useState(false)
  const [erro, setErro] = useState('')

  async function salvar(e) {
    e.preventDefault()
    setErro('')

    if (!matricula.trim()) {
      setErro('Informe a matrícula.')
      return
    }
    if (senhaTemporaria.length < 8) {
      setErro('A senha temporária precisa ter pelo menos 8 caracteres.')
      return
    }

    setSalvando(true)
    try {
      await criarUsuario({ matricula: matricula.trim(), senhaTemporaria, papel })
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
        Matrícula *
        <input type="text" value={matricula} onChange={(e) => setMatricula(e.target.value)} autoFocus />
      </label>

      <label className={styles.campo}>
        Senha temporária *
        <input
          type="text"
          value={senhaTemporaria}
          onChange={(e) => setSenhaTemporaria(e.target.value)}
          placeholder="Mínimo 8 caracteres"
        />
      </label>

      <label className={styles.campo}>
        Papel
        <select value={papel} onChange={(e) => setPapel(e.target.value)}>
          <option value="colaborador">Colaborador (só consulta colaboradores/relatórios)</option>
          <option value="admin">Admin (edita/cadastra/exclui)</option>
        </select>
      </label>

      <p className={styles.avisoSenha}>
        A pessoa vai ser obrigada a trocar essa senha no primeiro login — pode passar de boca ou por mensagem sem risco.
      </p>

      {erro && <p className={styles.erro}>{erro}</p>}

      <footer className={styles.rodapeForm}>
        <button type="button" className={styles.botaoSecundario} onClick={onCancelar}>
          ← Voltar
        </button>
        <button type="submit" className={styles.botaoPrimario} disabled={salvando}>
          {salvando ? 'Criando…' : 'Criar usuário'}
        </button>
      </footer>
    </form>
  )
}

function FormularioResetarSenha({ usuario, onCancelar, onResetado }) {
  const [senhaTemporaria, setSenhaTemporaria] = useState('')
  const [salvando, setSalvando] = useState(false)
  const [erro, setErro] = useState('')

  async function salvar(e) {
    e.preventDefault()
    setErro('')

    if (senhaTemporaria.length < 8) {
      setErro('A senha temporária precisa ter pelo menos 8 caracteres.')
      return
    }

    setSalvando(true)
    try {
      await resetarSenhaDeUsuario(usuario.id, senhaTemporaria)
      onResetado()
    } catch (err) {
      setErro(err.message)
    } finally {
      setSalvando(false)
    }
  }

  return (
    <form className={styles.form} onSubmit={salvar}>
      <p className={styles.avisoSenha}>
        Define uma nova senha temporária pra <strong>{usuario?.matricula}</strong>. Não precisa saber a senha antiga —
        a pessoa vai trocar no próximo login.
      </p>

      <label className={styles.campo}>
        Nova senha temporária *
        <input
          type="text"
          value={senhaTemporaria}
          onChange={(e) => setSenhaTemporaria(e.target.value)}
          placeholder="Mínimo 8 caracteres"
          autoFocus
        />
      </label>

      {erro && <p className={styles.erro}>{erro}</p>}

      <footer className={styles.rodapeForm}>
        <button type="button" className={styles.botaoSecundario} onClick={onCancelar}>
          ← Voltar
        </button>
        <button type="submit" className={styles.botaoPrimario} disabled={salvando}>
          {salvando ? 'Salvando…' : 'Resetar senha'}
        </button>
      </footer>
    </form>
  )
}
