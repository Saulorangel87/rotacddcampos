"""Gera a lista usada para revisar geometrias já existentes no banco.

O arquivo de saída contém os metadados dos cadastros do relatório de
duplicidades e uma cópia somente leitura da geometria atual. A ferramenta HTML
usa essa cópia como referência para desenhar um trecho novo; nada é alterado
no banco por este script.

Uso:
    python scripts/gerar_revisao_geometrias_existentes.py
"""

from __future__ import annotations

import csv
import hashlib
import json
import os
from pathlib import Path

import psycopg2


ROOT = Path(__file__).resolve().parents[1]
PASTA = Path(__file__).resolve().parent
ENV_PATH = ROOT / "Backend" / ".env"
RELATORIO = PASTA / "relatorio_duplicidades_geometrias.csv"
SAIDA = PASTA / "revisao_geometrias_existentes.json"


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


def conexao():
    carregar_env()
    return psycopg2.connect(
        host=os.environ.get("DB_HOST", "localhost"),
        port=int(os.environ.get("DB_PORT", "5432")),
        dbname=os.environ.get("DB_NAME", "rotas_db"),
        user=os.environ.get("DB_USER", "postgres"),
        password=os.environ.get("DB_PASSWORD", ""),
    )


def texto(valor) -> str:
    return "" if valor is None else str(valor).strip()


def ler_json(valor):
    if not valor or not str(valor).strip():
        return None
    try:
        return json.loads(valor) if isinstance(valor, str) else valor
    except (TypeError, ValueError, json.JSONDecodeError):
        return None


def hash_geometria(geometria) -> str:
    if geometria is None:
        return ""
    canonico = json.dumps(geometria, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonico.encode("utf-8")).hexdigest()[:16]


def main() -> None:
    if not RELATORIO.exists():
        raise SystemExit(f"Não achei {RELATORIO}.")

    with RELATORIO.open(encoding="utf-8-sig", newline="") as arquivo:
        relatorio = list(csv.DictReader(arquivo, delimiter=";"))
    ids = sorted({int(linha["rua_id"]) for linha in relatorio if linha.get("rua_id")})
    if not ids:
        raise SystemExit("O relatório não contém IDs válidos.")

    por_id = {int(linha["rua_id"]): linha for linha in relatorio}
    with conexao() as banco, banco.cursor() as cursor:
        cursor.execute(
            """
            SELECT id, nome_rua, bairro, cep, distrito, geometria
            FROM ruas
            WHERE ativo = true AND id = ANY(%s)
            ORDER BY id
            """,
            (ids,),
        )
        registros = cursor.fetchall()

    saida = []
    for rua_id, nome_rua, bairro, cep, distrito, geometria_bruta in registros:
        geometria = ler_json(geometria_bruta)
        linha = por_id.get(int(rua_id), {})
        saida.append(
            {
                "rua_id": int(rua_id),
                "nome": texto(nome_rua),
                "bairro": texto(bairro),
                "cep": texto(cep),
                "distrito": texto(distrito),
                "status_grupo": texto(linha.get("status_grupo")),
                "risco_casamento_osm": texto(linha.get("risco_casamento_osm")),
                "estado_geometria": "valida" if geometria else "sem_geometria",
                "hash_geometria": hash_geometria(geometria),
                "geometria": geometria,
            }
        )

    SAIDA.write_text(json.dumps(saida, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"{len(saida)} cadastros exportados para revisão.")
    print(f"Arquivo: {SAIDA}")
    print(f"Com geometria atual: {sum(bool(item['geometria']) for item in saida)}")
    print(f"Sem geometria atual: {sum(not item['geometria'] for item in saida)}")


if __name__ == "__main__":
    main()
