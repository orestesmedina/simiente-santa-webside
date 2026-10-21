#!/usr/bin/env python3
"""Comprueba que la integración continua del proyecto ejecute los verbos del perfil (equipo/perfil.json).

La integración continua es del proyecto: sus workflows (.github/workflows/) los redacta el proyecto desde su
perfil (comando bowser-ci), porque instalar un lenguaje o levantar una base de datos depende de la tecnología.
El kit solo trae bowser.yml, con los controles que no dependen de ella. Este script mira los demás workflows y
dice qué verbos de qué partes ejecutan, para que una parte no quede sin revisar sin que nadie lo note.

Reconoce las dos formas de ejecutar un verbo:
  make ci | lint | test | coverage | security | check-generated          (con PART=<nombre>, solo esa parte)
  python3 scripts/verbos.py <verbo>...                                    (con --parte <nombre>, solo esa parte)

Uso:
  python3 scripts/ci.py              muestra qué se ejecuta y avisa de lo que falta (termina bien)
  python3 scripts/ci.py --verificar  además termina con error si falta algo

Solo usa la biblioteca estándar de Python 3.9+.
"""
from __future__ import annotations

import os
import re
import sys

from perfil import MAKE_DE_VERBOS, RAIZ, RUTA, VERBOS, comando_de, perfil_valido

CARPETA = ".github/workflows"
DEL_KIT = {"bowser.yml"}        # el workflow del kit no ejecuta verbos
# Qué verbos ejecuta cada comando de make. `make ci` los ejecuta todos; `make generate` no cuenta: en la
# integración continua lo que comprueba es check-generated.
DE_MAKE = {comando: set(verbos) for verbos, comando in MAKE_DE_VERBOS if comando != "generate"}
DE_MAKE.update({"ci": set(VERBOS), "check-generated": {"generar"}})
# Los workflows escritos antes de la 1.22.0 usan los nombres anteriores de los comandos y de la opción.
DE_MAKE.update({"cobertura": DE_MAKE["coverage"], "verificar-generados": DE_MAKE["check-generated"]})
CON_PARTE = ("PART=", "PARTE=")
CON_MAKE = re.compile(r"\bmake\b([^\n|;&]*)")
CON_SCRIPT = re.compile(r"scripts/verbos\.py\b([^\n|;&]*)")
TODAS = "*"


def partes_de(nombre: str | None, partes: list[str]) -> list[str]:
    """Las partes a las que se refiere una llamada. Sin nombre, o con uno que se resuelve al ejecutar
    (${{ matrix.parte }}, $PART), no se puede saber: se toma como todas."""
    if not nombre or "$" in nombre:
        return partes
    nombre = nombre.strip("\"'")
    return [nombre] if nombre in partes else []


def ejecutados(texto: str, partes: list[str]) -> set[tuple[str, str]]:
    """Los pares (parte, verbo) que un workflow ejecuta."""
    pares: set[tuple[str, str]] = set()
    for linea in texto.splitlines():
        linea = re.sub(r"(^|\s)#.*", "", linea)      # lo comentado no se ejecuta
        for m in CON_MAKE.finditer(linea):
            palabras = m.group(1).split()
            parte = next((p.split("=", 1)[1] for p in palabras if p.startswith(CON_PARTE)), None)
            verbos = set().union(*(DE_MAKE.get(p, set()) for p in palabras))
            pares |= {(p, v) for p in partes_de(parte, partes) for v in verbos}
        for m in CON_SCRIPT.finditer(linea):
            palabras = m.group(1).split()
            parte = palabras[palabras.index("--parte") + 1] if "--parte" in palabras[:-1] else None
            pares |= {(p, v) for p in partes_de(parte, partes) for v in palabras if v in VERBOS}
    return pares


def workflows_del_proyecto(raiz=RAIZ) -> list:
    carpeta = raiz / CARPETA
    return sorted(f for f in carpeta.glob("*.y*ml") if f.name not in DEL_KIT) if carpeta.is_dir() else []


def sin_ejecutar(perfil: dict, raiz=RAIZ) -> set[tuple[str, str]]:
    """Los pares (parte, verbo) que el perfil define y ningún workflow del proyecto ejecuta."""
    definidos = {(p["nombre"], v) for p in perfil["partes"] for v in VERBOS if comando_de(p, v)}
    nombres = [p["nombre"] for p in perfil["partes"]]
    for f in workflows_del_proyecto(raiz):
        definidos -= ejecutados(f.read_text(encoding="utf-8", errors="replace"), nombres)
    return definidos


def main() -> int:
    perfil = perfil_valido()
    if perfil is None:
        print(f"Este proyecto no tiene perfil ({RUTA}): no hay verbos que comprobar.")
        return 0
    definidos = {(p["nombre"], v) for p in perfil["partes"] for v in VERBOS if comando_de(p, v)}
    if not definidos:
        print("El perfil no define ningún verbo todavía: no hay nada que la integración continua deba ejecutar.")
        return 0

    nombres = [p["nombre"] for p in perfil["partes"]]
    cubiertos: set[tuple[str, str]] = set()
    for f in workflows_del_proyecto():
        pares = ejecutados(f.read_text(encoding="utf-8", errors="replace"), nombres) & definidos
        if pares:
            print(f"{CARPETA}/{f.name}")
            for nombre in nombres:
                verbos = [v for v in VERBOS if (nombre, v) in pares]
                if verbos:
                    print(f"  {nombre}: {', '.join(verbos)}")
        cubiertos |= pares

    faltan = definidos - cubiertos
    if not faltan:
        print(f"✓ La integración continua ejecuta los {len(definidos)} verbos que el perfil define.")
        return 0
    detalle = "; ".join(f"{n}: {', '.join(v for v in VERBOS if (n, v) in faltan)}"
                        for n in nombres if any(p == n for p, _ in faltan))
    mensaje = (f"Ningún workflow del proyecto ejecuta estos verbos del perfil ({detalle}). "
               "Lo que no corre en la integración continua nadie lo revisa en un Pull Request. "
               "Para redactarla desde el perfil, escribe en el chat: /bowser-ci")
    print(f"::warning::{mensaje}" if os.environ.get("GITHUB_ACTIONS") else f"⚠ {mensaje}")
    return 1 if "--verificar" in sys.argv[1:] else 0


if __name__ == "__main__":
    sys.exit(main())
