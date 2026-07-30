import { useEffect, useState } from 'react'
import { listarRuas } from '../api/ruas.js'
import { corDoDistrito } from '../data/distritos.js'
import styles from './CepLookup.module.css'

export default function CepLookup() {
  const [termo, setTermo] = useState('')
  const [resultados, setResultados] = useState([])
  const [buscando, setBuscando] = useState(false)
  const [jaBuscou, setJaBuscou] = useState(false)

  useEffect(() => {
    if (termo.trim().length < 3) {
      setResultados([])
      setJaBuscou(false)
      return
    }
    setBuscando(true)
    const timer = setTimeout(() => {
      listarRuas({ nome: termo }).then((dados) => {
        setResultados(dados)
        setBuscando(false)
        setJaBuscou(true)
      })
    }, 400)
    return () => clearTimeout(timer)
  }, [termo])

  return (
    <section className={styles.caixa} aria-label="Consulta de CEP por rua">
      <h2 className={styles.titulo}>📮 Consultar CEP por rua</h2>
      <p className={styles.dica}>Digite pelo menos 3 letras do nome da rua.</p>

      <input
        type="search"
        className={styles.campo}
        placeholder="Ex: Alberto Torres"
        value={termo}
        onChange={(e) => setTermo(e.target.value)}
        autoFocus
      />

      {buscando && <p className={styles.mensagem}>Buscando…</p>}

      {!buscando && jaBuscou && resultados.length === 0 && (
        <p className={styles.mensagem}>Nenhuma rua encontrada com esse nome.</p>
      )}

      {!buscando && resultados.length > 0 && (
        <ul className={styles.lista}>
          {resultados.map((r) => (
            <li key={r.id}>
              <div>
                <strong>{r.nome_rua}</strong>
                <span className={styles.bairro}>{r.bairro}</span>
              </div>
              <div className={styles.direita}>
                <span className={styles.cep}>{r.cep}</span>
                <span className={styles.chipDistrito} style={{ background: corDoDistrito(r.distrito) }}>
                  {r.distrito}
                </span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
