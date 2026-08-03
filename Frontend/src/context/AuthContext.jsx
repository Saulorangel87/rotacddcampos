import { createContext, useContext, useEffect, useState, useCallback } from 'react'
import { login as apiLogin } from '../api/auth.js'
import { salvarToken, limparToken, getToken, aoExpirarSessao } from '../api/client.js'

const CHAVE_SESSAO = 'rotas_sessao'
const AuthContext = createContext(null)

function lerSessaoSalva() {
  const token = getToken()
  if (!token) return null
  try {
    const bruto = localStorage.getItem(CHAVE_SESSAO)
    if (!bruto) return null
    return JSON.parse(bruto)
  } catch {
    return null
  }
}

export function AuthProvider({ children }) {
  const [sessao, setSessao] = useState(lerSessaoSalva)

  useEffect(() => {
    // Se a API responder 401 em qualquer chamada (token expirado/inválido),
    // desloga e volta pra tela pública — sem isso, o usuário ficaria vendo
    // como "logado" uma sessão que o backend já não aceita mais.
    return aoExpirarSessao(() => setSessao(null))
  }, [])

  const entrar = useCallback(async (matricula, senha) => {
    const resposta = await apiLogin(matricula, senha)
    salvarToken(resposta.token)
    const novaSessao = {
      matricula: resposta.matricula,
      papel: resposta.papel,
      senhaProvisoria: resposta.senha_provisoria,
    }
    localStorage.setItem(CHAVE_SESSAO, JSON.stringify(novaSessao))
    setSessao(novaSessao)
    return novaSessao
  }, [])

  const sair = useCallback(() => {
    limparToken()
    localStorage.removeItem(CHAVE_SESSAO)
    setSessao(null)
  }, [])

  const marcarSenhaTrocada = useCallback(() => {
    setSessao((prev) => {
      if (!prev) return prev
      const atualizada = { ...prev, senhaProvisoria: false }
      localStorage.setItem(CHAVE_SESSAO, JSON.stringify(atualizada))
      return atualizada
    })
  }, [])

  const valor = {
    sessao,
    autenticado: !!sessao,
    admin: sessao?.papel === 'admin',
    entrar,
    sair,
    marcarSenhaTrocada,
  }

  return <AuthContext.Provider value={valor}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const contexto = useContext(AuthContext)
  if (!contexto) {
    throw new Error('useAuth precisa ser usado dentro de <AuthProvider>')
  }
  return contexto
}
