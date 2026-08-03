import { apiFetch, apiFetchJson } from './client.js'

/**
 * Lista usuários de acesso (GET /auth/usuarios — exige admin).
 */
export async function listarUsuarios() {
  const res = await apiFetch('/auth/usuarios')
  if (!res.ok) {
    const dados = await res.json().catch(() => ({}))
    throw new Error(dados.error || `API respondeu ${res.status}`)
  }
  return await res.json()
}

/**
 * Cadastra um novo usuário de acesso (POST /auth/usuarios — exige admin).
 */
export async function criarUsuario({ matricula, senhaTemporaria, papel }) {
  return apiFetchJson('/auth/usuarios', {
    method: 'POST',
    body: JSON.stringify({
      matricula,
      senha_temporaria: senhaTemporaria,
      papel,
    }),
  })
}

/**
 * Reseta a senha de outro usuário (POST /auth/usuarios/:id/resetar-senha — exige admin).
 * Não pede a senha atual — o admin não sabe e não deveria saber.
 */
export async function resetarSenhaDeUsuario(id, senhaTemporaria) {
  await apiFetchJson(`/auth/usuarios/${id}/resetar-senha`, {
    method: 'POST',
    body: JSON.stringify({ senha_temporaria: senhaTemporaria }),
  })
}
