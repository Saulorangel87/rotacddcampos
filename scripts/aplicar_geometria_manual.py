"""
Aplica no banco as geometrias desenhadas manualmente na ferramenta
desenhar-ruas-manual.html — pra ruas que nunca bateram com nada no
OpenStreetMap (ver listar_ruas_sem_match.py).

Lê geometrias_manuais.json (baixado da ferramenta HTML) e grava somente as
geometrias de ruas ativas que ainda estão vazias, no mesmo formato GeoJSON MultiLineString que
casar_ruas_osm.py e aplicar_revisao_osm.py já usam — então o mapa do site
lê tudo do mesmo jeito, não importa se veio do OSM automático ou de desenho
manual.

Registros que já possuem geometria, IDs inexistentes e desenhos fora dos
limites amplos de Campos dos Goytacazes são ignorados e listados no final;
nenhum desenho existente é sobrescrito automaticamente.

Requisitos: pip install psycopg2-binary

Uso:
  export DB_PASSWORD=...
  python aplicar_geometria_manual.py
"""

import json
import math
import os

import psycopg2

DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = int(os.environ.get("DB_PORT", "5432"))
DB_NAME = os.environ.get("DB_NAME", "rotas_db")
DB_USER = os.environ.get("DB_USER", "postgres")
DB_PASSWORD = os.environ["DB_PASSWORD"]

PASTA = os.path.dirname(os.path.abspath(__file__))
ARQUIVO_GEOMETRIAS = os.path.join(PASTA, "geometrias_manuais.json")

LIMITES_CAMPOS = {
    "min_lat": -22.2,
    "max_lat": -21.1,
    "min_lon": -42.0,
    "max_lon": -40.7,
}


def geometria_valida(geometria):
    """Aceita somente MultiLineString com dois ou mais pontos válidos."""
    if not isinstance(geometria, dict) or geometria.get("type") != "MultiLineString":
        return False
    linhas = geometria.get("coordinates")
    if not isinstance(linhas, list):
        return False

    pontos = 0
    for linha in linhas:
        if not isinstance(linha, list):
            return False
        if len(linha) < 2:
            continue
        for ponto in linha:
            if not isinstance(ponto, list) or len(ponto) < 2:
                return False
            try:
                longitude = float(ponto[0])
                latitude = float(ponto[1])
            except (TypeError, ValueError):
                return False
            if not math.isfinite(longitude) or not math.isfinite(latitude):
                return False
            if not (LIMITES_CAMPOS["min_lon"] <= longitude <= LIMITES_CAMPOS["max_lon"]):
                return False
            if not (LIMITES_CAMPOS["min_lat"] <= latitude <= LIMITES_CAMPOS["max_lat"]):
                return False
            pontos += 1
    return pontos >= 2


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
    ja_preenchidos = []
    invalidas = []
    ids_processados = set()
    for entrada in entradas:
        try:
            rua_id = int(entrada["rua_id"])
        except (KeyError, TypeError, ValueError):
            invalidas.append(entrada.get("rua_id"))
            continue
        geometria = entrada.get("geometria")

        if rua_id in ids_processados or not geometria_valida(geometria):
            invalidas.append(rua_id)
            continue
        ids_processados.add(rua_id)

        cursor.execute(
            "SELECT COALESCE(btrim(geometria), '') FROM ruas WHERE id = %s AND ativo = true",
            (rua_id,),
        )
        registro = cursor.fetchone()
        if not registro:
            nao_encontrados.append(rua_id)
            continue
        if registro[0]:
            ja_preenchidos.append(rua_id)
            continue

        cursor.execute(
            """
            UPDATE ruas
            SET geometria = %s, updated_at = now()
            WHERE id = %s AND ativo = true AND (geometria IS NULL OR btrim(geometria) = '')
            """,
            (json.dumps(geometria, ensure_ascii=False), rua_id),
        )
        aplicados += 1

    conexao.commit()
    cursor.close()
    conexao.close()

    print(f"{aplicados} ruas atualizadas com geometria desenhada manualmente.")
    if nao_encontrados:
        print(f"{len(nao_encontrados)} rua_id não encontrados como ativos no banco (ignorados): {nao_encontrados}")
    if ja_preenchidos:
        print(f"{len(ja_preenchidos)} ruas já tinham geometria e não foram sobrescritas: {ja_preenchidos}")
    if invalidas:
        print(f"{len(invalidas)} entradas inválidas ou duplicadas (ignoradas): {invalidas}")


if __name__ == "__main__":
    main()
