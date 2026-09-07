"""Funções comuns para separar ways do OSM por continuidade geográfica."""

from __future__ import annotations

import math
from collections import defaultdict
from typing import Any


LIMIAR_CONEXAO_METROS = 80.0


def distancia_metros(a: list[float], b: list[float]) -> float:
    """Distância Haversine entre pontos no formato GeoJSON [lon, lat]."""

    raio = 6_371_000.0
    lon1, lat1 = map(math.radians, a[:2])
    lon2, lat2 = map(math.radians, b[:2])
    dlat = lat2 - lat1
    dlon = lon2 - lon1
    hav = math.sin(dlat / 2) ** 2 + math.cos(lat1) * math.cos(lat2) * math.sin(dlon / 2) ** 2
    return 2 * raio * math.asin(min(1.0, math.sqrt(hav)))


def _pontos_extremos(segmento: dict[str, Any]) -> list[list[float]]:
    coordenadas = segmento.get("coordenadas") or []
    if len(coordenadas) < 2:
        return []
    return [coordenadas[0], coordenadas[-1]]


def _unir(parent: list[int], esquerda: int, direita: int) -> None:
    raiz_esquerda = esquerda
    while parent[raiz_esquerda] != raiz_esquerda:
        parent[raiz_esquerda] = parent[parent[raiz_esquerda]]
        raiz_esquerda = parent[raiz_esquerda]
    raiz_direita = direita
    while parent[raiz_direita] != raiz_direita:
        parent[raiz_direita] = parent[parent[raiz_direita]]
        raiz_direita = parent[raiz_direita]
    if raiz_esquerda != raiz_direita:
        parent[raiz_direita] = raiz_esquerda


def agrupar_segmentos(
    segmentos: list[dict[str, Any]],
    limiar_metros: float = LIMIAR_CONEXAO_METROS,
) -> list[dict[str, Any]]:
    """Retorna componentes conectados pelos extremos dos ways.

    Ways adjacentes normalmente compartilham um nó no OSM. O pequeno limiar
    cobre diferenças de alguns metros e mantém separados trechos do mesmo nome
    que estão em bairros diferentes. A ordem e os IDs são determinísticos.
    """

    if not segmentos:
        return []

    parent = list(range(len(segmentos)))
    extremos = [_pontos_extremos(segmento) for segmento in segmentos]
    for indice, pontos in enumerate(extremos):
        if not pontos:
            continue
        for outro_indice in range(indice + 1, len(extremos)):
            outros = extremos[outro_indice]
            if not outros:
                continue
            if min(distancia_metros(a, b) for a in pontos for b in outros) <= limiar_metros:
                _unir(parent, indice, outro_indice)

    componentes: dict[int, list[dict[str, Any]]] = defaultdict(list)
    for indice, segmento in enumerate(segmentos):
        raiz = indice
        while parent[raiz] != raiz:
            parent[raiz] = parent[parent[raiz]]
            raiz = parent[raiz]
        componentes[raiz].append(segmento)

    resultado = []
    for grupo in componentes.values():
        grupo = sorted(grupo, key=lambda item: int(item.get("id", 0)))
        resultado.append(
            {
                "segmentos": grupo,
                "ids": [int(item["id"]) for item in grupo if item.get("id") is not None],
            }
        )
    return sorted(resultado, key=lambda item: (item["ids"] or [0])[0])
