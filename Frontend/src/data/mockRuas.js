// Dados de exemplo para desenvolvimento da interface, no mesmo formato do
// model Rua da API Go (Backend/models/rua.go). Quando GET /ruas responder,
// src/api/ruas.js troca automaticamente para os dados reais do PostgreSQL.
let seq = 1
const rua = (nome, cep, distrito, carteiro, bairro) => ({
  id: seq++,
  nome_rua: nome,
  cep,
  distrito,
  bairro,
  rota: carteiro,
  observacao: '',
  atualizado_em: '2026-07-27',
})

export const MOCK_RUAS = [
  rua('Rua das Acácias', '28035-120', '602', 'Carlos Silva', 'Centro'),
  rua('Rua das Palmeiras', '28035-130', '602', 'Carlos Silva', 'Centro'),
  rua('Rua João Batista', '28035-140', '602', 'Carlos Silva', 'Centro'),
  rua('Rua Santo Antônio', '28035-150', '602', 'Carlos Silva', 'Centro'),
  rua('Rua Almirante Barroso', '28035-160', '602', 'Carlos Silva', 'Centro'),
  rua('Rua XV de Novembro', '28035-170', '602', 'Carlos Silva', 'Centro'),
  rua('Rua Marechal Floriano', '28035-180', '602', 'Carlos Silva', 'Centro'),
  rua('Avenida Alberto Torres', '28035-200', '601', 'Maria Souza', 'Parque Pecuária'),
  rua('Rua Barão de Miracema', '28035-210', '601', 'Maria Souza', 'Parque Pecuária'),
  rua('Rua Formosa', '28035-220', '603', 'João Pereira', 'Rogino'),
  rua('Rua Coronel Rangel', '28035-230', '603', 'João Pereira', 'Rogino'),
  rua('Estrada da Aurora', '28035-300', '604', 'Ana Costa', 'Aurora'),
  rua('Rua dos Cravos', '28035-310', '605', 'Ana Costa', 'Guarus'),
  rua('Avenida Rui Barbosa', '28035-400', '606', 'Pedro Lima', 'Guarus Plaza'),
  rua('Rua Tenente Coronel Cardoso', '28035-410', '607', 'Pedro Lima', 'Horto'),
  rua('Rua Estação das Plantas', '28035-420', '607', 'Pedro Lima', 'Horto'),
  rua('Avenida Carlos Chagas', '28035-500', '608', 'Fábio Nunes', 'Parque Carmelitas'),
  rua('Rua da Barra', '28035-510', '609', 'Fábio Nunes', 'Vila Nova'),
]

export const CARTEIROS = [...new Set(MOCK_RUAS.map((r) => r.rota))]
