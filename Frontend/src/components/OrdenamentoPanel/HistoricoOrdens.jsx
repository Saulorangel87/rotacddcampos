import { restaurarSequencia } from './ordemManual.js'
import styles from './OrdenamentoPanel.module.css'

export default function HistoricoOrdens({ historico = [], paradas, ocupado, onRestaurar, onEsquecer }) {
  if (!historico.length) return null
  return (
    <details className={styles.historico}>
      <summary>Meu histórico de correções</summary>
      <p>Últimos 20 salvamentos. Limpar a carga mantém este histórico.</p>
      {historico.map((versao) => {
        const sequencia = restaurarSequencia(paradas, versao)
        const habitualAtiva = versao.reutilizar && !versao.desativada_em
        return (
          <details key={versao.id} className={styles.versaoHistorico}>
            <summary>
              {new Date(versao.created_at).toLocaleString('pt-BR')} · {versao.paradas.length} ruas
              <span>{habitualAtiva ? 'Sequência habitual' : versao.reutilizar ? 'Sequência habitual desativada' : 'Ajuste da carga'}</span>
            </summary>
            <ol>
              {versao.paradas.map((parada) => (
                <li key={parada.chave_agrupamento}>
                  <strong>{parada.nome_rua}</strong>
                  <span>Sugestão original: {parada.ordem_sugerida} · Ordem escolhida: {parada.ordem_final}</span>
                </li>
              ))}
            </ol>
            <div className={styles.historicoAcoes}>
              <button type="button" disabled={ocupado || !sequencia} onClick={() => onRestaurar(sequencia)}>
                Usar esta versão na carga
              </button>
              {habitualAtiva && (
                <button type="button" disabled={ocupado} onClick={() => onEsquecer(versao.id)}>
                  Deixar de usar como habitual
                </button>
              )}
            </div>
            {!sequencia && <p>Gere uma lista com as mesmas ruas para usar esta versão.</p>}
          </details>
        )
      })}
    </details>
  )
}
