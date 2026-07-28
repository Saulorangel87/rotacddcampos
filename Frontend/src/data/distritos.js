// Metadados dos 9 distritos do CDD Campos dos Goytacazes.
// As cores são as mesmas já usadas na legenda do site atual (style.css),
// mantidas aqui para não quebrar a identidade visual que os carteiros já reconhecem.
export const DISTRITOS = [
  { numero: '601', cor: 'var(--d601)', nome: 'Distrito 601' },
  { numero: '602', cor: 'var(--d602)', nome: 'Distrito 602' },
  { numero: '603', cor: 'var(--d603)', nome: 'Distrito 603' },
  { numero: '604', cor: 'var(--d604)', nome: 'Distrito 604' },
  { numero: '605', cor: 'var(--d605)', nome: 'Distrito 605' },
  { numero: '606', cor: 'var(--d606)', nome: 'Distrito 606' },
  { numero: '607', cor: 'var(--d607)', nome: 'Distrito 607' },
  { numero: '608', cor: 'var(--d608)', nome: 'Distrito 608' },
  { numero: '609', cor: 'var(--d609)', nome: 'Distrito 609' },
]

export function corDoDistrito(numero) {
  return DISTRITOS.find((d) => d.numero === String(numero))?.cor || 'var(--cinza-400)'
}

// Layout aproximado (não geográfico) só para desenhar o mini-mapa em SVG,
// seguindo a disposição relativa do mapa do Google My Maps atual:
// 601 acima-esquerda, 606 acima-direita, 608/609 ponta direita,
// 602 centro, 603/604/605 faixa inferior, 607 direita.
export const LAYOUT_DISTRITOS = {
  601: { x: 40, y: 20, w: 150, h: 130 },
  606: { x: 210, y: 15, w: 140, h: 110 },
  608: { x: 360, y: 10, w: 120, h: 95 },
  602: { x: 130, y: 150, w: 140, h: 120 },
  607: { x: 280, y: 125, w: 150, h: 130 },
  609: { x: 440, y: 110, w: 90, h: 130 },
  603: { x: 20, y: 260, w: 130, h: 110 },
  604: { x: 160, y: 270, w: 110, h: 100 },
  605: { x: 280, y: 255, w: 130, h: 115 },
}
