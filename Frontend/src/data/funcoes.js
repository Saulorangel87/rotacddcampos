// Lista fechada de funções, batendo exatamente com os padrões que o backend usa
// pra contar por categoria (Backend/handlers/estatisticas_handler.go). Usar texto
// livre aqui quebra a contagem — por isso isso vira um <select>, não input de texto.
export const FUNCOES = [
  'MOTORIZADO (M)',
  'MOTORIZADO (V)',
  'CICLISTA',
  'INTERNO',
  'ADMINISTRATIVO',
  'OTT',
  'OT',
  'SUPERVISOR',
  'GERENTE',
]
