"""
Audita possíveis trocas de geometria causadas por nomes de rua repetidos.

O script é somente leitura: não altera o banco e não consulta o OSM. Ele usa
as geometrias já gravadas em ``ruas`` para encontrar os casos que precisam de
uma nova associação por bairro, distrito ou CEP.

Saídas padrão:
  - relatorio_duplicidades_geometrias.csv: uma linha por cadastro;
  - relatorio_duplicidades_geometrias_grupos.csv: uma linha por nome exato repetido;
  - relatorio_ambiguidades_casamento_osm.csv: colisões da chave-base usada pelo OSM.

Para executar:
    python scripts/auditar_duplicidades_geometrias.py

As credenciais podem vir do ambiente ou de Backend/.env. Nenhum valor sensível
é exibido no terminal.
"""

from __future__ import annotations

import csv
import hashlib
import json
import math
import os
import re
import unicodedata
from collections import defaultdict
from pathlib import Path
from typing import Any, Iterable

import psycopg2


ROOT = Path(__file__).resolve().parents[1]
ENV_PATH = ROOT / "Backend" / ".env"
OUT_ROWS = Path(__file__).resolve().with_name("relatorio_duplicidades_geometrias.csv")
OUT_GROUPS = Path(__file__).resolve().with_name("relatorio_duplicidades_geometrias_grupos.csv")
OUT_BASE = Path(__file__).resolve().with_name("relatorio_ambiguidades_casamento_osm.csv")

PREFIXOS = re.compile(
    r"^(RUA|AVENIDA|AV|TRAVESSA|TV|PRA[CÇ]A|ESTRADA|RODOVIA|ALAMEDA|LARGO|BOULEVARD)\.?\s+"
)


def carregar_env() -> None:
    """Carrega apenas variáveis ainda ausentes, sem substituir o ambiente."""

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


def _normalizar(nome: str, remover_faixa: bool) -> str:
    nome = unicodedata.normalize("NFKD", nome.upper()).encode("ascii", "ignore").decode()
    nome = PREFIXOS.sub("", nome)
    if remover_faixa:
        nome = re.split(r"\s*-\s*(ATE|DE|LADO)\b", nome)[0]
    nome = re.sub(r"[^A-Z0-9]+", " ", nome)
    return re.sub(r"\s+", " ", nome).strip()


def normalizar(nome: str) -> str:
    """Chave-base compatível com o casamento OSM atual."""

    return _normalizar(nome, remover_faixa=True)


def normalizar_exato(nome: str) -> str:
    """Normaliza acentos e tipo de via, preservando faixa/lado/número."""

    return _normalizar(nome, remover_faixa=False)


def iterar_pontos(coordenadas: Any) -> Iterable[tuple[float, float]]:
    """Percorre coordenadas GeoJSON sem depender do tipo exato da geometria."""

    if isinstance(coordenadas, (list, tuple)):
        if (
            len(coordenadas) >= 2
            and isinstance(coordenadas[0], (int, float))
            and isinstance(coordenadas[1], (int, float))
        ):
            lon, lat = float(coordenadas[0]), float(coordenadas[1])
            if math.isfinite(lon) and math.isfinite(lat) and -180 <= lon <= 180 and -90 <= lat <= 90:
                yield lon, lat
            return
        for item in coordenadas:
            yield from iterar_pontos(item)


def extrair_geometria(valor: Any) -> tuple[list[tuple[float, float]], str, str]:
    if valor is None or not str(valor).strip():
        return [], "sem_geometria", ""
    try:
        objeto = json.loads(valor) if isinstance(valor, str) else valor
        pontos = list(iterar_pontos(objeto.get("coordinates")))
        if not pontos:
            return [], "geometria_invalida", ""
        canonico = json.dumps(objeto, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
        digest = hashlib.sha256(canonico.encode("utf-8")).hexdigest()[:16]
        return pontos, "valida", digest
    except (TypeError, ValueError, json.JSONDecodeError):
        return [], "geometria_invalida", ""


def distancia_metros(a: tuple[float, float], b: tuple[float, float]) -> float:
    raio = 6_371_000.0
    lon1, lat1 = map(math.radians, a)
    lon2, lat2 = map(math.radians, b)
    dlat = lat2 - lat1
    dlon = lon2 - lon1
    hav = math.sin(dlat / 2) ** 2 + math.cos(lat1) * math.cos(lat2) * math.sin(dlon / 2) ** 2
    return 2 * raio * math.asin(min(1.0, math.sqrt(hav)))


def ponto_medio(pontos: list[tuple[float, float]]) -> tuple[float, float]:
    return (
        sum(ponto[0] for ponto in pontos) / len(pontos),
        sum(ponto[1] for ponto in pontos) / len(pontos),
    )


def conectar():
    carregar_env()
    return psycopg2.connect(
        host=os.environ.get("DB_HOST", "localhost"),
        port=int(os.environ.get("DB_PORT", "5432")),
        dbname=os.environ.get("DB_NAME", "rotas_db"),
        user=os.environ.get("DB_USER", "postgres"),
        password=os.environ.get("DB_PASSWORD", ""),
    )


def texto(valor: Any) -> str:
    return "" if valor is None else str(valor).strip()


def auditoria() -> tuple[list[dict[str, str]], list[dict[str, str]], list[dict[str, str]]]:
    with conectar() as conexao, conexao.cursor() as cursor:
        cursor.execute(
            """
            SELECT id, nome_rua, bairro, cep, distrito, geometria
            FROM ruas
            WHERE ativo = true
            ORDER BY id
            """
        )
        registros = cursor.fetchall()

    grupos_base: dict[str, list[dict[str, Any]]] = defaultdict(list)
    grupos_exatos: dict[str, list[dict[str, Any]]] = defaultdict(list)
    for rua_id, nome_rua, bairro, cep, distrito, geometria in registros:
        pontos, estado, digest = extrair_geometria(geometria)
        centro = ponto_medio(pontos) if pontos else None
        extensao = 0.0
        if len(pontos) > 1:
            extensao = distancia_metros(
                (min(ponto[0] for ponto in pontos), min(ponto[1] for ponto in pontos)),
                (max(ponto[0] for ponto in pontos), max(ponto[1] for ponto in pontos)),
            )
        registro = {
            "id": int(rua_id),
            "nome_rua": texto(nome_rua),
            "nome_normalizado": normalizar_exato(texto(nome_rua)),
            "nome_base_normalizado": normalizar(texto(nome_rua)),
            "bairro": texto(bairro),
            "cep": texto(cep),
            "distrito": texto(distrito),
            "estado_geometria": estado,
            "hash_geometria": digest,
            "centro": centro,
            "extensao_m": extensao,
        }
        grupos_base[registro["nome_base_normalizado"]].append(registro)
        grupos_exatos[registro["nome_normalizado"]].append(registro)

    # A chave-base é a que o importador OSM usa e pode colidir entre bairros.
    # Guardamos o diagnóstico em cada linha do relatório exato para tornar a
    # revisão acionável sem misturar faixas legítimas da mesma avenida.
    informacoes_base: dict[int, dict[str, str]] = {}
    resumos_base: list[dict[str, str]] = []
    for nome_base, grupo_base in grupos_base.items():
        if len(grupo_base) < 2:
            continue
        contextos_base = {(item["bairro"], item["distrito"], item["cep"]) for item in grupo_base}
        risco = "ambiguidade_por_contexto" if len(contextos_base) > 1 else "segmentos_do_mesmo_contexto"
        ids_base = ",".join(str(item["id"]) for item in grupo_base)
        for item in grupo_base:
            informacoes_base[item["id"]] = {
                "quantidade_cadastros_base": str(len(grupo_base)),
                "quantidade_contextos_base": str(len(contextos_base)),
                "risco_casamento_osm": risco,
                "ids_chave_base": ids_base,
            }
        resumos_base.append(
            {
                "nome_base_normalizado": nome_base,
                "nomes_cadastrados": " | ".join(sorted({item["nome_rua"] for item in grupo_base})),
                "quantidade_cadastros": str(len(grupo_base)),
                "quantidade_contextos": str(len(contextos_base)),
                "status": risco,
                "ids": ids_base,
                "bairros": ",".join(sorted({item["bairro"] for item in grupo_base if item["bairro"]})),
                "ceps": ",".join(sorted({item["cep"] for item in grupo_base if item["cep"]})),
                "distritos": ",".join(sorted({item["distrito"] for item in grupo_base if item["distrito"]})),
            }
        )

    linhas: list[dict[str, str]] = []
    resumos: list[dict[str, str]] = []
    for nome_normalizado, grupo in grupos_exatos.items():
        if len(grupo) < 2:
            continue

        contextos = {(item["bairro"], item["distrito"], item["cep"]) for item in grupo}
        hashes_validos = [item["hash_geometria"] for item in grupo if item["hash_geometria"]]
        contagem_hash = defaultdict(int)
        for digest in hashes_validos:
            contagem_hash[digest] += 1
        hashes_compartilhados = [digest for digest, quantidade in contagem_hash.items() if quantidade > 1]

        centros = [item["centro"] for item in grupo if item["centro"] is not None]
        distancia_max_centros = 0.0
        for indice, centro in enumerate(centros):
            for outro in centros[indice + 1 :]:
                distancia_max_centros = max(distancia_max_centros, distancia_metros(centro, outro))

        if hashes_compartilhados and len(contextos) > 1:
            status_grupo = "geometria_compartilhada_contextos_distintos"
        elif distancia_max_centros > 500:
            status_grupo = "geometrias_distantes_mesmo_nome"
        elif len(contextos) > 1:
            status_grupo = "nome_repetido_revisar"
        else:
            status_grupo = "cadastros_repetidos_mesmo_contexto"

        ids = ",".join(str(item["id"]) for item in grupo)
        bairros = ",".join(sorted({item["bairro"] for item in grupo if item["bairro"]}))
        ceps = ",".join(sorted({item["cep"] for item in grupo if item["cep"]}))
        distritos = ",".join(sorted({item["distrito"] for item in grupo if item["distrito"]}))
        resumo = {
            "nome_normalizado": nome_normalizado,
            "nome_exemplo": grupo[0]["nome_rua"],
            "quantidade_cadastros": str(len(grupo)),
            "quantidade_contextos": str(len(contextos)),
            "quantidade_geometrias_validas": str(sum(item["estado_geometria"] == "valida" for item in grupo)),
            "quantidade_geometrias_unicas": str(len(set(hashes_validos))),
            "quantidade_geometrias_compartilhadas": str(sum(contagem_hash[digest] for digest in hashes_compartilhados)),
            "distancia_maxima_entre_centros_m": f"{distancia_max_centros:.1f}",
            "extensao_maxima_da_geometria_m": f"{max(item["extensao_m"] for item in grupo):.1f}",
            "status": status_grupo,
            "ids": ids,
            "bairros": bairros,
            "ceps": ceps,
            "distritos": distritos,
        }
        resumos.append(resumo)

        for item in grupo:
            base_info = informacoes_base.get(
                item["id"],
                {
                    "quantidade_cadastros_base": "1",
                    "quantidade_contextos_base": "1",
                    "risco_casamento_osm": "chave_base_unica",
                    "ids_chave_base": str(item["id"]),
                },
            )
            outras_distancias = [
                distancia_metros(item["centro"], outro["centro"])
                for outro in grupo
                if outro is not item and item["centro"] is not None and outro["centro"] is not None
            ]
            linhas.append(
                {
                    "rua_id": str(item["id"]),
                    "nome_rua": item["nome_rua"],
                    "nome_normalizado": nome_normalizado,
                    "nome_base_normalizado": item["nome_base_normalizado"],
                    "bairro": item["bairro"],
                    "cep": item["cep"],
                    "distrito": item["distrito"],
                    "estado_geometria": item["estado_geometria"],
                    "hash_geometria": item["hash_geometria"],
                    "extensao_geometria_m": f"{item["extensao_m"]:.1f}",
                    "distancia_minima_outra_geometria_m": (
                        f"{min(outras_distancias):.1f}" if outras_distancias else ""
                    ),
                    "status_grupo": status_grupo,
                    "ids_do_grupo": ids,
                    **base_info,
                }
            )

    resumos.sort(key=lambda item: (item["status"], -float(item["distancia_maxima_entre_centros_m"])))
    linhas.sort(key=lambda item: (item["status_grupo"], item["nome_normalizado"], int(item["rua_id"])))
    resumos_base.sort(key=lambda item: (item["status"], item["nome_base_normalizado"]))
    return linhas, resumos, resumos_base


def gravar_csv(caminho: Path, linhas: list[dict[str, str]]) -> None:
    if not linhas:
        caminho.write_text("", encoding="utf-8")
        return
    with caminho.open("w", newline="", encoding="utf-8-sig") as arquivo:
        escritor = csv.DictWriter(arquivo, fieldnames=list(linhas[0]), delimiter=";")
        escritor.writeheader()
        escritor.writerows(linhas)


def main() -> None:
    linhas, resumos, resumos_base = auditoria()
    gravar_csv(OUT_ROWS, linhas)
    gravar_csv(OUT_GROUPS, resumos)
    gravar_csv(OUT_BASE, resumos_base)

    contagem_status = defaultdict(int)
    for resumo in resumos:
        contagem_status[resumo["status"]] += 1
    grupos_base_ambiguos = sum(item["status"] == "ambiguidade_por_contexto" for item in resumos_base)
    print(f"{len(linhas)} cadastros pertencem a {len(resumos)} grupos de nomes exatos repetidos.")
    print(f"{grupos_base_ambiguos} chaves-base do casamento OSM têm contextos diferentes e precisam de revisão.")
    for status, quantidade in sorted(contagem_status.items()):
        print(f"  {status}: {quantidade} grupos")
    print(f"Relatório por cadastro: {OUT_ROWS}")
    print(f"Resumo por grupo: {OUT_GROUPS}")
    print(f"Colisões da chave-base OSM: {OUT_BASE}")


if __name__ == "__main__":
    main()
