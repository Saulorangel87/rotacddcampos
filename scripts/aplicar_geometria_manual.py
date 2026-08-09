"""
Aplica no banco as geometrias desenhadas manualmente na ferramenta
desenhar-ruas-manual.html — pra ruas que nunca bateram com nada no
OpenStreetMap (ver listar_ruas_sem_match.py).

Lê geometrias_manuais.json (baixado da ferramenta HTML) e grava cada
geometria em ruas.geometria, no mesmo formato GeoJSON MultiLineString que
casar_ruas_osm.py e aplicar_revisao_osm.py já usam — então o mapa do site
lê tudo do mesmo jeito, não importa se veio do OSM automático ou de desenho
manual.

Requisitos: pip install psycopg2-binary

Uso:
  export DB_PASSWORD=...
  python aplicar_geometria_manual.py
"""

import json
import os

import psycopg2

DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = int(os.environ.get("DB_PORT", "5432"))
DB_NAME = os.environ.get("DB_NAME", "rotas_db")
DB_USER = os.environ.get("DB_USER", "postgres")
DB_PASSWORD = os.environ["DB_PASSWORD"]

PASTA = os.path.dirname(os.path.abspath(__file__))
ARQUIVO_GEOMETRIAS = os.path.join(PASTA, "geometrias_manuais.json")


def main():
    if not os.path.exists(ARQUIVO_GEOMETRIAS):
        print(f"Não achei {ARQUIVO_GEOMETRIAS} nesta pasta.")
        return

    with open(ARQUIVO_GEOMETRIAS, encoding="utf-8") as f:
        entradas = json.load(f)

    print(f"{len(entradas)} geometrias desenhadas manualmente pra aplicar.")

    conexao = psycopg2.connect(
        host=DB_HOST, port=DB_PORT, dbname=DB_NAME, user=DB_USER, password=DB_PASSWORD,
    )
    cursor = conexao.cursor()

    aplicados = 0
    nao_encontrados = []
    for entrada in entradas:
        rua_id = entrada["rua_id"]
        geometria = entrada["geometria"]

        cursor.execute("SELECT 1 FROM ruas WHERE id = %s AND ativo = true", (rua_id,))
        if not cursor.fetchone():
            nao_encontrados.append(rua_id)
            continue

        cursor.execute(
            "UPDATE ruas SET geometria = %s, updated_at = now() WHERE id = %s",
            (json.dumps(geometria, ensure_ascii=False), rua_id),
        )
        aplicados += 1

    conexao.commit()
    cursor.close()
    conexao.close()

    print(f"{aplicados} ruas atualizadas com geometria desenhada manualmente.")
    if nao_encontrados:
        print(f"{len(nao_encontrados)} rua_id não encontrados como ativos no banco (ignorados): {nao_encontrados}")


if __name__ == "__main__":
    main()
