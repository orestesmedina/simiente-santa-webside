#!/usr/bin/env python3
"""Fusiona el delta de F4 (specs/004-eventos-actividades/contracts/openapi.yaml)
en el contrato vivo backend/api/openapi.yaml de forma ADITIVA (P4-16).

Solo hace las cuatro inserciones mecánicas (tags, paths, parámetro de id y
schemas); los cambios de prosa y el enum de AdminActionItem se hacen aparte.
Reejecutable: comprueba que no exista ya el contenido.
"""
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parents[1]
DELTA = ROOT / "specs/004-eventos-actividades/contracts/openapi.yaml"
LIVE = ROOT / "backend/api/openapi.yaml"

delta_lines = DELTA.read_text(encoding="utf-8").splitlines(keepends=True)
live = LIVE.read_text(encoding="utf-8")

# --- bloques del delta (1-indexado, inclusivo) ---------------------------------
delta_paths = "".join(delta_lines[145:782])   # líneas 146-782: los 9 path items
delta_tags = "".join(delta_lines[136:144])    # líneas 137-144: las 4 etiquetas
delta_schemas = "".join(delta_lines[814:1519])  # líneas 815-1519: los schemas nuevos

assert delta_paths.startswith("  /api/v1/proximamente:"), delta_paths[:60]
assert delta_tags.startswith("  - name: proximamente"), delta_tags[:60]
assert delta_schemas.startswith("    # ------------------------------------------------------------ comunes"), delta_schemas[:80]


def quote_plain_scalars(text: str) -> str:
    """Entrecomilla los escalares simples que contienen ": ".

    El snapshot del delta (inmutable) trae cinco `description`/`summary` sin
    entrecomillar cuyo valor incluye ": " (p. ej. "Siempre `no-store` (SC-003:
    los cambios…)"). Un escalar simple de YAML no puede contener ": ": el
    documento del delta NO parsea tal cual. La fusión lo corrige entrecomillando
    el valor (sin tocar el texto) para que el contrato vivo sea YAML válido.
    """
    pattern = re.compile(r'^(\s*)(description|summary):\s+(.*)$')
    out = []
    for line in text.splitlines(keepends=True):
        match = pattern.match(line.rstrip("\n"))
        if match:
            value = match.group(3)
            if value[:1] not in ('"', "'", "|", ">") and ": " in value:
                line = f'{match.group(1)}{match.group(2)}: "{value}"\n'
        out.append(line)
    return "".join(out)


delta_paths = quote_plain_scalars(delta_paths)
delta_schemas = quote_plain_scalars(delta_schemas)

# Nombres de los parameters del documento vivo (regla anti-duplicación): la
# paginación reutiliza `Limit`/`Offset`, ya definidos por F2 con la MISMA
# semántica (defecto 20, tope 100, §8.1.3). El id de los elementos de F4 usa un
# parámetro propio, como `PortadaItemId` de F3.
delta_paths = delta_paths.replace("#/components/parameters/ListLimit", "#/components/parameters/Limit")
delta_paths = delta_paths.replace("#/components/parameters/ListOffset", "#/components/parameters/Offset")

# --- 1) etiquetas nuevas -------------------------------------------------------
tags_anchor = "  - name: portada-admin\n    description: Gestión de la portada desde el panel (permiso `portada` de F2)\n"
if "  - name: proximamente\n" in live:
    sys.exit("las etiquetas de F4 ya están en el documento vivo")
assert tags_anchor in live
live = live.replace(tags_anchor, tags_anchor + delta_tags)

# --- 2) paths nuevos (antes de `components:`) ---------------------------------
paths_anchor = "\ncomponents:\n  securitySchemes:\n"
if "  /api/v1/proximamente:" in live:
    sys.exit("los paths de F4 ya están en el documento vivo")
assert live.count(paths_anchor) == 1
live = live.replace(paths_anchor, "\n" + delta_paths + paths_anchor)

# --- 3) parámetro del id de agenda --------------------------------------------
param_anchor = ("    PortadaItemId:\n"
                "      name: id\n"
                "      in: path\n"
                "      required: true\n"
                "      description: Identificador (UUID) del elemento de la portada.\n"
                "      schema:\n"
                "        type: string\n"
                "        format: uuid\n")
agenda_param = ("    AgendaItemId:\n"
                "      name: id\n"
                "      in: path\n"
                "      required: true\n"
                "      description: Identificador (UUID) del elemento de agenda (evento, actividad o categoría).\n"
                "      schema:\n"
                "        type: string\n"
                "        format: uuid\n")
if "    AgendaItemId:" in live:
    sys.exit("el parámetro AgendaItemId ya está en el documento vivo")
assert live.count(param_anchor) == 1
live = live.replace(param_anchor, param_anchor + agenda_param)

# --- 4) schemas nuevos (al final de la sección, antes de `responses:`) --------
schemas_anchor = "\n  responses:\n    NotFound:\n"
if "    ProximamentePublica:" in live:
    sys.exit("los schemas de F4 ya están en el documento vivo")
assert live.count(schemas_anchor) == 1
live = live.replace(schemas_anchor, "\n" + delta_schemas + schemas_anchor)

LIVE.write_text(live, encoding="utf-8")
print("delta de F4 fusionado en backend/api/openapi.yaml")
