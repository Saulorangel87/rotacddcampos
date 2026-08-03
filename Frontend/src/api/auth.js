import { apiFetchJson } from './client.js'

/**
 * Login (POST /auth/login). Lança erro com a mensagem da API
 * (credenciais inválidas, conta bloqueada, etc.) em caso de falha.
 */
export async function login(matricula, senha) {
  return apiFetchJson('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ matricula, senha }),
  })
}

/**
 * Troca a própria senha (POST /auth/trocar-senha). Exige estar autenticado
 * (apiFetch já injeta o token). Usado tanto pra troca voluntária quanto
 * pra sair do estado de senha provisória no primeiro acesso.
 */
export async function trocarSenha(senhaAtual, senhaNova) {
  await apiFetchJson('/auth/trocar-senha', {
    method: 'POST',
    body: JSON.stringify({ senha_atual: senhaAtual, senha_nova: senhaNova }),
  })
}

/**
 * Cadastra um novo usuário de acesso (POST /auth/usuarios). Só admin pode
 * chamar — o backend recusa com 403 se quem estiver logado não for admin.
 */
export async function criarUsuario({ matricula, senhaTemporaria, papel, colaboradorId }) {
  return apiFetchJson('/auth/usuarios', {
    method: 'POST',
    body: JSON.stringify({
      matricula,
      senha_temporaria: senhaTemporaria,
      papel,
      colaborador_id: colaboradorId,
    }),
  })
}
