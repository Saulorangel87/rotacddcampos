"""Gera o lote e a ferramenta de desenho manual iniciada do zero.

O lote exclui as decisões OSM aplicadas localmente e mantém os cinco IDs que
ficaram pendentes na aplicação atual. A ferramenta resultante é separada da
lista histórica de ruas sem geometria e não carrega o traçado atual.

Uso:
    python scripts/gerar_lote_revisao_manual.py
"""

from __future__ import annotations

import csv
import json
import re
from pathlib import Path


PASTA = Path(__file__).resolve().parent
RELATORIO = PASTA / "relatorio_duplicidades_geometrias.csv"
DECISOES = PASTA / "decisoes_duplicidades_osm.json"
MODELO = PASTA / "desenhar-ruas-manual.html"
LOTE = PASTA / "lote_revisao_manual_pendente.json"
SAIDA_HTML = PASTA / "desenhar-revisoes-pendentes.html"

# Resultado registrado pelo último aplicador OSM. Esses IDs continuam no lote
# porque a escolha não pôde ser aplicada com segurança.
IDS_PENDENTES_APLICACAO = {969, 974, 1618, 1812, 1822}


def carregar_lote() -> list[dict]:
    with RELATORIO.open(encoding="utf-8-sig", newline="") as arquivo:
        relatorio = list(csv.DictReader(arquivo, delimiter=";"))
    decisoes = json.loads(DECISOES.read_text(encoding="utf-8"))
    ids_decididos = {
        int(item["rua_id"])
        for item in decisoes
        if item.get("decisao") == "aceito" and int(item["rua_id"]) not in IDS_PENDENTES_APLICACAO
    }
    lote = []
    for linha in relatorio:
        rua_id = int(linha["rua_id"])
        if rua_id in ids_decididos:
            continue
        lote.append(
            {
                "ruaId": rua_id,
                "nome": linha.get("nome_rua", ""),
                "bairro": linha.get("bairro", ""),
                "cep": linha.get("cep", ""),
                "distrito": linha.get("distrito", ""),
                "status": "revisao_manual_pendente",
                "confianca": "",
                "nomeOsmSugerido": "",
            }
        )
    return sorted(lote, key=lambda item: item["ruaId"])


def gerar_html(lote: list[dict]) -> None:
    html = MODELO.read_text(encoding="utf-8")
    dados = json.dumps(lote, ensure_ascii=False, separators=(",", ":"))
    html, quantidade = re.subn(
        r"const DADOS = .*?;\nconst CHAVE_STORAGE",
        f"const DADOS = {dados};\nconst CHAVE_STORAGE",
        html,
        count=1,
        flags=re.DOTALL,
    )
    if quantidade != 1:
        raise RuntimeError("Não foi possível substituir a lista DADOS do modelo.")
    substituicoes = {
        "<title>Editor manual de ruas — CDD Campos</title>": "<title>Desenho manual de revisões — CDD Campos</title>",
        "<h1>Editor manual de geometrias</h1>": "<h1>Desenho manual de revisões pendentes</h1>",
        "<p class=\"subtitulo\">CDD Campos dos Goytacazes · coordenadas para o ordenamento</p>": "<p class=\"subtitulo\">CDD Campos dos Goytacazes · desenho do trecho postal do zero</p>",
        "Ruas sem geometria": "Cadastros pendentes de revisão",
        "const CHAVE_STORAGE = 'geometrias_manuais_v2';": "const CHAVE_STORAGE = 'revisao_manual_pendente_v1';",
        "document.getElementById('status-rua').textContent = item.status === 'nunca_revisada_osm' ? 'Sem correspondência anterior no OSM. Confirme a rua visualmente antes de desenhar.' : `A busca OSM anterior foi rejeitada. Sugestão anterior: ${item.nomeOsmSugerido || 'nenhuma'}.`;": "document.getElementById('status-rua').textContent = 'Desenho iniciado do zero. Confira bairro, CEP e trace somente o trecho correto.';",
        "link.download = 'geometrias_manuais.json';": "link.download = 'revisoes_geometrias_manuais.json';",
        "Geometria salva neste navegador. Exporte o JSON para revisão e aplicação no banco.": "Correção salva neste navegador. Exporte o JSON para revisão e aplicação no banco.",
    }
    for antigo, novo in substituicoes.items():
        if antigo not in html:
            raise RuntimeError(f"Trecho esperado não encontrado no modelo: {antigo[:80]}")
        html = html.replace(antigo, novo)
    html = html.replace(
        "<p class=\"ajuda\"><strong>Como usar:</strong> localize a rua, confirme o bairro no mapa e clique no traçado da via em sequência.",
        "<p class=\"ajuda\"><strong>Como usar:</strong> esta lista começa com o desenho vazio. Localize o bairro e o CEP, confirme a área no mapa e clique somente no trecho postal correto.",
    )
    SAIDA_HTML.write_text(html, encoding="utf-8")


def main() -> None:
    lote = carregar_lote()
    LOTE.write_text(json.dumps(lote, ensure_ascii=False, indent=2), encoding="utf-8")
    gerar_html(lote)
    print(f"{len(lote)} cadastros pendentes incluídos no lote manual.")
    print(f"Lote: {LOTE}")
    print(f"Ferramenta: {SAIDA_HTML}")


if __name__ == "__main__":
    main()
