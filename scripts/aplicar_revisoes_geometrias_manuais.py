"""Valida e aplica correções manuais de geometrias existentes.

O modo padrão é simulação e não grava nada. A escrita exige ``--aplicar`` e
usa o hash da geometria original exportado pela ferramenta para evitar
sobrescrever uma alteração feita depois da revisão.

Uso:
    python scripts/aplicar_revisoes_geometrias_manuais.py \
        --arquivo scripts/revisoes_geometrias_manuais.json
    python scripts/aplicar_revisoes_geometrias_manuais.py \
        --arquivo scripts/revisoes_geometrias_manuais.json --aplicar
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import os
from pathlib import Path

import psycopg2


ROOT = Path(__file__).resolve().parents[1]
ENV_PATH = ROOT / "Backend" / ".env"
PASTA = Path(__file__).resolve().parent
ARQUIVO_PADRAO = PASTA / "revisoes_geometrias_manuais.json"
LIMITES = {"min_lat": -22.2, "max_lat": -21.1, "min_lon": -42.0, "max_lon": -40.7}


def carregar_env() -> None:
    if not ENV_PATH.exists():
        return
    for linha in ENV_PATH.read_text(encoding="utf-8").splitlines():
        linha = linha.strip()
        if not linha or linha.startswith("#") or "=" not in linha:
            continue
        chave, valor = linha.split("=", 1)
        chave = chave.strip()
        if chave and chave not in os.environ:
            os.environ[chave] = valor.strip().strip('"').strip("'")


def conectar():
    carregar_env()
    return psycopg2.connect(
        host=os.environ.get("DB_HOST", "localhost"),
        port=int(os.environ.get("DB_PORT", "5432")),
        dbname=os.environ.get("DB_NAME", "rotas_db"),
        user=os.environ.get("DB_USER", "postgres"),
        password=os.environ.get("DB_PASSWORD", ""),
    )


def hash_geometria(geometria) -> str:
    if geometria is None:
        return ""
    canonico = json.dumps(geometria, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonico.encode("utf-8")).hexdigest()[:16]


def geometria_valida(geometria) -> bool:
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
                lon, lat = float(ponto[0]), float(ponto[1])
            except (TypeError, ValueError):
                return False
            if not math.isfinite(lon) or not math.isfinite(lat):
                return False
            if not (LIMITES["min_lon"] <= lon <= LIMITES["max_lon"]):
                return False
            if not (LIMITES["min_lat"] <= lat <= LIMITES["max_lat"]):
                return False
            pontos += 1
    return pontos >= 2


def main() -> None:
    parser = argparse.ArgumentParser(description="Valida e aplica correções manuais de geometrias.")
    parser.add_argument("--arquivo", default=str(ARQUIVO_PADRAO), help="JSON exportado pela ferramenta manual.")
    parser.add_argument("--aplicar", action="store_true", help="Grava as correções aprovadas no banco.")
    args = parser.parse_args()

    arquivo = Path(args.arquivo).resolve()
    if not arquivo.exists():
        raise SystemExit(f"Não achei {arquivo}.")
    try:
        entradas = json.loads(arquivo.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as erro:
        raise SystemExit(f"Não foi possível ler {arquivo}: {erro}") from erro
    if not isinstance(entradas, list):
        raise SystemExit("O JSON precisa conter uma lista de correções.")

    ids_vistos: set[int] = set()
    invalidas = []
    candidatos = []
    for entrada in entradas:
        try:
            rua_id = int(entrada["rua_id"])
            geometria = entrada["geometria"]
        except (KeyError, TypeError, ValueError):
            invalidas.append(entrada.get("rua_id") if isinstance(entrada, dict) else None)
            continue
        if rua_id in ids_vistos or not geometria_valida(geometria):
            invalidas.append(rua_id)
            continue
        ids_vistos.add(rua_id)
        candidatos.append((rua_id, geometria, entrada))

    print(f"{len(candidatos)} correções válidas encontradas.")
    if invalidas:
        print(f"{len(invalidas)} entradas inválidas ou duplicadas serão ignoradas: {invalidas}")
    if not candidatos:
        print("Nada para revisar.")
        return

    conflitos = []
    nao_encontrados = []
    prontos = []
    with conectar() as banco, banco.cursor() as cursor:
        for rua_id, geometria, entrada in candidatos:
            cursor.execute(
                "SELECT nome_rua, geometria FROM ruas WHERE id = %s AND ativo = true",
                (rua_id,),
            )
            atual = cursor.fetchone()
            if not atual:
                nao_encontrados.append(rua_id)
                continue
            hash_original = str(entrada.get("hash_geometria_original") or "").strip()
            hash_atual = hash_geometria(json.loads(atual[1]) if atual[1] else None)
            if hash_original and hash_original != hash_atual:
                conflitos.append((rua_id, hash_original, hash_atual))
                continue
            prontos.append((rua_id, geometria, atual[0]))

        print(f"{len(prontos)} correções prontas para aplicação.")
        if nao_encontrados:
            print(f"{len(nao_encontrados)} IDs não encontrados como ativos: {nao_encontrados}")
        if conflitos:
            print(f"{len(conflitos)} conflitos de geometria atual (revisar novamente): {[item[0] for item in conflitos]}")

        if args.aplicar:
            for rua_id, geometria, _ in prontos:
                cursor.execute(
                    "UPDATE ruas SET geometria = %s, updated_at = now() WHERE id = %s AND ativo = true",
                    (json.dumps(geometria, ensure_ascii=False), rua_id),
                )
            banco.commit()
            print(f"{len(prontos)} correções aplicadas no banco configurado.")
        else:
            banco.rollback()
            print("Modo simulação: nenhuma geometria foi alterada. Use --aplicar depois da revisão.")


if __name__ == "__main__":
    main()
