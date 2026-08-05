import { createContext, useContext, useEffect, useState, useCallback, useRef } from 'react'
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

// Lê o "exp" (data de expiração) direto do JWT, sem precisar perguntar pro
// backend. Não valida assinatura — é só leitura, pra decidir quando deslogar
// na tela; a validação de verdade continua sendo feita pelo backend em toda
// chamada. Sem isso, quem loga e não navega mais nada continua "aparecendo"
// logado na tela indefinidamente, mesmo com o token já expirado, porque só
// descobríamos isso quando uma chamada à API falhava com 401.
function lerExpiracaoDoToken(token) {
  try {
    const payload = token.split('.')[1]
    const json = atob(payload.replace(/-/g, '+').replace(/_/g, '/'))
    const dados = JSON.parse(json)
    return dados.exp ? dados.exp * 1000 : null
  } catch {
    return null
  }
}

export function AuthProvider({ children }) {
  const [sessao, setSessao] = useState(lerSessaoSalva)
  const timerExpiracao = useRef(null)

  const sair = useCallback(() => {
    limparToken()
    localStorage.removeItem(CHAVE_SESSAO)
    setSessao(null)
    if (timerExpiracao.current) {
      clearTimeout(timerExpiracao.current)
      timerExpiracao.current = null
    }
  }, [])

  // Agenda o logout automático pro exato momento em que o token expira,
  // em vez de esperar a próxima chamada à API falhar com 401.
  const agendarExpiracao = useCallback((token) => {
    if (timerExpiracao.current) clearTimeout(timerExpiracao.current)
    const expiraEm = lerExpiracaoDoToken(token)
    if (!expiraEm) return
    const faltam = expiraEm - Date.now()
    if (faltam <= 0) {
      sair()
      return
    }
    // setTimeout tem limite de ~24.8 dias; nosso token expira em poucas
    // horas, então nunca deve bater nesse teto, mas por segurança limitamos.
    timerExpiracao.current = setTimeout(sair, Math.min(faltam, 2 ** 31 - 1))
  }, [sair])

  useEffect(() => {
    // Se a API responder 401 em qualquer chamada (token expirado/inválido),
    // desloga e volta pra tela pública — sem isso, o usuário ficaria vendo
    // como "logado" uma sessão que o backend já não aceita mais.
    return aoExpirarSessao(() => setSessao(null))
  }, [])

  // Ao carregar a página (ou trocar de sessão), reagenda o timer de expiração
  // com base no token que está salvo agora.
  useEffect(() => {
    const token = getToken()
    if (token && sessao) {
      agendarExpiracao(token)
    }
    return () => {
      if (timerExpiracao.current) clearTimeout(timerExpiracao.current)
    }
  }, [sessao, agendarExpiracao])

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
