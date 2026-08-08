"""
Lista as ruas ativas sem geometria que NUNCA apareceram em nenhuma leva de
casamento com o OpenStreetMap (nem na automática de alta confiança, nem no
CSV de revisão manual de confiança média). São os casos mais difíceis:
provavelmente não estão mapeadas no OSM com esse nome, ou não estão
mapeadas de jeito nenhum — candidatas a desenho manual.

Cruza:
  1. Todas as ruas ativas com geometria IS NULL no banco (produção).
  2. Todos os rua_id que já apareceram em algum revisao_matches_baixos.csv
     gerado até agora (aceitos ou rejeitados — já foram vistos por um
     humano, não interessa aqui).

O que sobra é escrito em ruas_sem_nenhum_match.csv, ordenado por nome.

Requisitos: pip install psycopg2-binary (mesmo do casar_ruas_osm.py)

Uso:
  export DB_PASSWORD=...
  python listar_ruas_sem_match.py
"""

import csv
import os

import psycopg2

DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = int(os.environ.get("DB_PORT", "5432"))
DB_NAME = os.environ.get("DB_NAME", "rotas_db")
DB_USER = os.environ.get("DB_USER", "postgres")
DB_PASSWORD = os.environ["DB_PASSWORD"]

PASTA = os.path.dirname(os.path.abspath(__file__))
CSV_REVISADAS = os.path.join(PASTA, "revisao_matches_baixos.csv")
CSV_SAIDA = os.path.join(PASTA, "ruas_sem_nenhum_match.csv")


def ids_ja_revisados():
    """rua_id que já apareceram no CSV de revisão mais recente (aceitos ou
    rejeitados — os dois já passaram por decisão humana)."""
    if not os.path.exists(CSV_REVISADAS):
        return set()
    with open(CSV_REVISADAS, encoding="utf-8") as f:
        reader = csv.DictReader(f, delimiter=";")
        return {int(row["rua_id"]) for row in reader}


def main():
    ja_revisados = ids_ja_revisados()
    print(f"{len(ja_revisados)} ruas já passaram por alguma leva de revisão manual (ignoradas aqui).")

    conexao = psycopg2.connect(
        host=DB_HOST, port=DB_PORT, dbname=DB_NAME, user=DB_USER, password=DB_PASSWORD,
    )
    cursor = conexao.cursor()
    cursor.execute(
        "SELECT id, nome_rua, distrito FROM ruas WHERE ativo = true AND geometria IS NULL ORDER BY nome_rua"
    )
    todas_sem_geometria = cursor.fetchall()
    conexao.close()

    nunca_vistas = [r for r in todas_sem_geometria if r[0] not in ja_revisados]

    print(f"{len(todas_sem_geometria)} ruas ativas sem geometria no total.")
    print(f"{len(nunca_vistas)} nunca passaram por nenhuma leva de casamento com o OSM.")

    with open(CSV_SAIDA, "w", encoding="utf-8", newline="") as f:
        writer = csv.writer(f, delimiter=";")
        writer.writerow(["rua_id", "nome_rua", "distrito"])
        for rua_id, nome_rua, distrito in nunca_vistas:
            writer.writerow([rua_id, nome_rua, distrito])

    print(f"Lista salva em {CSV_SAIDA}")


if __name__ == "__main__":
    main()
