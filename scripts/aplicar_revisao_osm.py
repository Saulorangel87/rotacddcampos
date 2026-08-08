"""
Aplica as decisões tomadas na ferramenta de revisão (revisao-ruas-osm.html)
direto no banco: pra cada rua marcada como "aceito", busca a geometria real
dela no OpenStreetMap e grava em ruas.geometria — igual o casar_ruas_osm.py
já faz pras de alta confiança, só que agora pra essas que você confirmou
manualmente.

Como usar:
  1. Termine (ou pare quando quiser) a revisão na ferramenta HTML.
  2. Clique em "Baixar decisões (JSON)" — salva decisoes_revisao_osm.json.
  3. Coloque esse arquivo na mesma pasta deste script (Backend/../scripts).
  4. Rode:
       python aplicar_revisao_osm.py

Requisitos: as mesmas do casar_ruas_osm.py (pip install requests psycopg2-binary).
"""

import json
import re
import os
import time
import unicodedata

import requests
import psycopg2

# ── Lê do ambiente — mesmo motivo do casar_ruas_osm.py ────────────────────
DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = int(os.environ.get("DB_PORT", "5432"))
DB_NAME = os.environ.get("DB_NAME", "rotas_db")
DB_USER = os.environ.get("DB_USER", "postgres")
DB_PASSWORD = os.environ["DB_PASSWORD"]
# ────────────────────────────────────────────────────────────────────────

ARQUIVO_DECISOES = os.path.join(os.path.dirname(os.path.abspath(__file__)), "decisoes_revisao_osm.json")

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
                resposta = requests.post(url, data={"data": QUERY_OVERPASS}, headers=HEADERS, timeout=200)
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
    try:
        with open(ARQUIVO_DECISOES, encoding="utf-8") as f:
            decisoes = json.load(f)
    except FileNotFoundError:
        print(f"Não achei {ARQUIVO_DECISOES} nesta pasta.")
        print("Baixe o JSON pela ferramenta de revisão e coloque ele aqui do lado do script.")
        return

    aceitos = [d for d in decisoes if d["decisao"] == "aceito"]
    rejeitados = [d for d in decisoes if d["decisao"] == "rejeitado"]
    print(f"{len(aceitos)} ruas aceitas na revisão, {len(rejeitados)} rejeitadas (essas ficam sem geometria, sem problema).")

    if not aceitos:
        print("Nada aceito ainda — nada pra aplicar.")
        return

    ruas_osm = buscar_ruas_osm()

    conexao = psycopg2.connect(
        host=DB_HOST, port=DB_PORT, dbname=DB_NAME, user=DB_USER, password=DB_PASSWORD,
        client_encoding="latin1",
    )
    cursor_setup = conexao.cursor()
    cursor_setup.execute("SET client_encoding TO 'UTF8';")
    cursor_setup.close()
    cursor = conexao.cursor()

    aplicados = 0
    nao_encontrados = []
    for d in aceitos:
        nome_norm = normalizar(d["nome_osm"])
        entrada = ruas_osm.get(nome_norm)
        if not entrada:
            nao_encontrados.append(d)
            continue
        geometria_geojson = {"type": "MultiLineString", "coordinates": entrada["segmentos"]}
        cursor.execute(
            "UPDATE ruas SET geometria = %s, updated_at = now() WHERE id = %s",
            (json.dumps(geometria_geojson, ensure_ascii=False), d["rua_id"]),
        )
        aplicados += 1

    conexao.commit()
    cursor.close()
    conexao.close()

    print(f"{aplicados} ruas atualizadas com geometria real do OpenStreetMap.")
    if nao_encontrados:
        print(f"{len(nao_encontrados)} não foram encontradas no OSM nessa consulta (nome pode ter mudado lá) — ficaram sem geometria:")
        for d in nao_encontrados[:20]:
            print(f"  - rua_id {d['rua_id']}: {d['nome_osm']}")


if __name__ == "__main__":
    main()
