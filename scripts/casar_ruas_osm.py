"""
Casa as ruas do banco (rotas_db) com o traçado real delas no OpenStreetMap,
usando a Overpass API — sem precisar desenhar nada na mão.

Como funciona:
  1. Busca todas as ruas nomeadas dentro do município de Campos dos Goytacazes
     no OpenStreetMap (via Overpass API).
  2. Normaliza os nomes dos dois lados (remove "RUA"/"AVENIDA", remove trechos
     tipo "- ATÉ 137 - LADO ÍMPAR", tira acento) pra comparar de forma justa.
  3. Pra cada rua do banco, acha o nome do OSM mais parecido.
  4. Só grava automaticamente as que baterem com confiança alta (>= 85%).
  5. As de confiança média (60-84%) NÃO são aplicadas — viram um CSV
     (revisao_matches_baixos.csv) pra você olhar com calma depois.

Requisitos (instala uma vez):
    pip install requests psycopg2-binary

Antes de rodar, ajusta os dados de conexão do banco logo abaixo
(mesmos dados que estão no seu Backend/.env).
"""

import json
import re
import time
import os
import unicodedata
from difflib import SequenceMatcher

import requests
import psycopg2

# ── Lê do ambiente — passe via variáveis de ambiente na hora de rodar,
# nunca deixe senha real gravada aqui no arquivo (esse script já vazou uma
# vez sem querer; ver nota-de-status-site-correios.md).
DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = int(os.environ.get("DB_PORT", "5432"))
DB_NAME = os.environ.get("DB_NAME", "rotas_db")
DB_USER = os.environ.get("DB_USER", "postgres")
DB_PASSWORD = os.environ["DB_PASSWORD"]
# ────────────────────────────────────────────────────────────────────────

OVERPASS_URLS = [
    "https://overpass-api.de/api/interpreter",
    "https://overpass.kumi.systems/api/interpreter",
]
HEADERS = {
    "User-Agent": "CDD-Campos-AjustesDeRotas/1.0 (uso interno Correios, contato: saulo)",
    "Accept": "*/*",
}
MUNICIPIO = "Campos dos Goytacazes"

TIPOS_VIA_RELEVANTES = (
    "residential|primary|secondary|tertiary|unclassified|living_street"
    "|primary_link|secondary_link|tertiary_link"
)

QUERY_OVERPASS = f"""
[out:json][timeout:180];
area["name"="{MUNICIPIO}"]["boundary"="administrative"]->.a;
(
  way["highway"~"^({TIPOS_VIA_RELEVANTES})$"]["name"](area.a);
);
out geom;
"""

PREFIXOS = r"^(RUA|AVENIDA|AV|TRAVESSA|TV|PRA[CÇ]A|ESTRADA|RODOVIA|ALAMEDA|LARGO|BOULEVARD)\.?\s+"


def normalizar(nome: str) -> str:
    nome = nome.upper()
    nome = unicodedata.normalize("NFKD", nome).encode("ascii", "ignore").decode()
    nome = re.sub(PREFIXOS, "", nome)
    # remove qualquer coisa depois de " - ATÉ", " - DE", " - LADO"
    nome = re.split(r"\s*-\s*(ATE|DE|LADO)\b", nome)[0]
    nome = re.sub(r"[.,]", "", nome)
    return nome.strip()


def buscar_ruas_osm() -> dict:
    print(f"Consultando Overpass API pra {MUNICIPIO}... (pode levar 1-2 minutos)")

    resposta = None
    ultimo_erro = None
    for url in OVERPASS_URLS:
        for tentativa in range(1, 4):
            try:
                print(f"  tentando {url} (tentativa {tentativa}/3) ...")
                resposta = requests.post(
                    url, data={"data": QUERY_OVERPASS}, headers=HEADERS, timeout=200
                )
                resposta.raise_for_status()
                break
            except requests.exceptions.RequestException as erro:
                ultimo_erro = erro
                print(f"  falhou ({erro})")
                resposta = None
                if tentativa < 3:
                    espera = 10 * tentativa
                    print(f"  esperando {espera}s antes de tentar de novo...")
                    time.sleep(espera)
        if resposta is not None:
            break

    if resposta is None:
        raise RuntimeError(f"Nenhum servidor Overpass respondeu. Último erro: {ultimo_erro}")

    dados = resposta.json()

    ruas_osm = {}
    for elemento in dados["elements"]:
        if elemento.get("type") != "way":
            continue
        tags = elemento.get("tags", {})
        nome = tags.get("name")
        geometria = elemento.get("geometry")
        if not nome or not geometria:
            continue

        nome_norm = normalizar(nome)
        coordenadas = [[ponto["lon"], ponto["lat"]] for ponto in geometria]
        ruas_osm.setdefault(nome_norm, {"nome_original": nome, "segmentos": []})
        ruas_osm[nome_norm]["segmentos"].append(coordenadas)

    print(f"{len(ruas_osm)} nomes de rua distintos encontrados no OpenStreetMap.")
    return ruas_osm


def main():
    ruas_osm = buscar_ruas_osm()
    nomes_osm = list(ruas_osm.keys())

    # Nada de client_encoding manual aqui — foi essa gambiarra (herdada de um
    # workaround antigo pra Windows) que causava os acentos saírem dobrados
    # no CSV ("ANTÃƒÂ”NIO" em vez de "ANTÔNIO"). psycopg2 conversa em UTF-8
    # com o Postgres sem ajuda nenhuma, então não mexe.
    conexao = psycopg2.connect(
        host=DB_HOST, port=DB_PORT, dbname=DB_NAME, user=DB_USER, password=DB_PASSWORD,
    )
    cursor = conexao.cursor()
    cursor.execute("SELECT id, nome_rua FROM ruas WHERE ativo = true AND geometria IS NULL")
    ruas_banco = cursor.fetchall()
    print(f"{len(ruas_banco)} ruas ativas no banco pra tentar casar.")

    matches_altos = []
    matches_baixos = []

    for rua_id, nome_rua in ruas_banco:
        nome_norm = normalizar(nome_rua)
        melhor_nome, melhor_score = None, 0.0
        for nome_osm in nomes_osm:
            score = SequenceMatcher(None, nome_norm, nome_osm).ratio()
            if score > melhor_score:
                melhor_score, melhor_nome = score, nome_osm

        if melhor_score >= 0.85:
            matches_altos.append((rua_id, nome_rua, melhor_nome, melhor_score))
        elif melhor_score >= 0.6:
            matches_baixos.append((rua_id, nome_rua, melhor_nome, melhor_score))

    print(f"{len(matches_altos)} correspondências de alta confiança (>=85%) — serão aplicadas.")
    print(f"{len(matches_baixos)} de confiança média (60-84%) — NÃO aplicadas, vão pro CSV de revisão.")

    aplicados = 0
    for rua_id, nome_rua, nome_osm, score in matches_altos:
        segmentos = ruas_osm[nome_osm]["segmentos"]
        geometria_geojson = {"type": "MultiLineString", "coordinates": segmentos}
        cursor.execute(
            "UPDATE ruas SET geometria = %s, updated_at = now() WHERE id = %s",
            (json.dumps(geometria_geojson, ensure_ascii=False), rua_id),
        )
        aplicados += 1

    conexao.commit()
    print(f"{aplicados} ruas atualizadas com geometria do OpenStreetMap.")

    caminho_csv = os.path.join(os.path.dirname(os.path.abspath(__file__)), "revisao_matches_baixos.csv")
    with open(caminho_csv, "w", encoding="utf-8") as arquivo:
        arquivo.write("rua_id;nome_banco;nome_osm_sugerido;confianca\n")
        for rua_id, nome_rua, nome_osm, score in matches_baixos:
            arquivo.write(f"{rua_id};{nome_rua};{nome_osm};{score:.2f}\n")

    print(f"Lista de revisão salva em {caminho_csv}")

    cursor.close()
    conexao.close()


if __name__ == "__main__":
    main()
