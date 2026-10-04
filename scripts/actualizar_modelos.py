#!/usr/bin/env python3
"""Aplica al proyecto los modelos recomendados por la versión del kit instalada en el submódulo.

equipo/config.json es del proyecto (semilla) y no cambia con `make actualizar-kit`. Este script
copia SOLO la asignación de modelos del config.json del kit al del proyecto:
    modelos.<herramienta>.orquestador, .ligero, .niveles, .agentes
Todo lo demás (herramientas activas, agente_principal, temperatura, kit.excluir, esfuerzo_codex…)
se conserva tal cual.

Uso (desde la raíz del proyecto):
  python3 scripts/actualizar_modelos.py              muestra los cambios y pide confirmación
  python3 scripts/actualizar_modelos.py --si         aplica sin preguntar
  python3 scripts/actualizar_modelos.py --comprobar  solo avisa si hay diferencias (no cambia nada)
"""
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

RAIZ = Path.cwd()
CONFIG_PROYECTO = RAIZ / "equipo/config.json"
CLAVES = ("orquestador", "ligero", "niveles", "agentes")


def ruta_kit() -> Path:
    try:
        salida = subprocess.run(["bash", "scripts/ruta-kit.sh"], capture_output=True, text=True, cwd=RAIZ).stdout.strip()
    except OSError:
        salida = ""
    return RAIZ / (salida or ".bowser-spec-kit-ai")


def aplanar(modelos_herramienta: dict) -> dict[str, str]:
    """{'orquestador': x, 'niveles': {...}, 'agentes': {...}} -> {'orquestador': x, 'nivel alto': y, 'agente dev-backend': z}"""
    plano = {}
    for clave in ("orquestador", "ligero"):
        if clave in modelos_herramienta:
            plano[clave] = modelos_herramienta.get(clave) or ""
    for nivel, modelo in (modelos_herramienta.get("niveles") or {}).items():
        plano[f"nivel {nivel}"] = modelo or ""
    for agente, modelo in (modelos_herramienta.get("agentes") or {}).items():
        plano[f"agente {agente}"] = modelo or ""
    return plano


def diferencias(cfg_proyecto: dict, cfg_kit: dict) -> dict[str, list[tuple[str, str, str]]]:
    cambios: dict[str, list[tuple[str, str, str]]] = {}
    for herramienta, kit_h in (cfg_kit.get("modelos") or {}).items():
        actual = aplanar((cfg_proyecto.get("modelos") or {}).get(herramienta) or {})
        nuevo = aplanar(kit_h or {})
        filas = []
        for clave in sorted(set(actual) | set(nuevo)):
            antes, despues = actual.get(clave, "—"), nuevo.get(clave, "—")
            if antes != despues:
                filas.append((clave, antes or "(sesión)", despues or "(sesión)"))
        if filas:
            cambios[herramienta] = filas
    return cambios


def main() -> int:
    args = sys.argv[1:]
    kit = ruta_kit()
    config_kit = kit / "equipo/config.json"
    if not config_kit.exists():
        if "--comprobar" in args:
            return 0
        print(f"No encuentro el kit en {kit.relative_to(RAIZ)}/ (¿submódulo sin inicializar? git submodule update --init).", file=sys.stderr)
        return 1
    if not CONFIG_PROYECTO.exists():
        print("No existe equipo/config.json en el proyecto. Ejecuta: make instalar-kit", file=sys.stderr)
        return 1

    cfg_kit = json.loads(config_kit.read_text(encoding="utf-8"))
    cfg_proy = json.loads(CONFIG_PROYECTO.read_text(encoding="utf-8"))
    cambios = diferencias(cfg_proy, cfg_kit)

    if "--comprobar" in args:
        if cambios:
            n = sum(len(v) for v in cambios.values())
            print(f"Aviso: el kit recomienda modelos distintos a los de este proyecto ({n} diferencias). "
                  "Revísalos con: make actualizar-modelos")
        return 0

    if not cambios:
        print("Los modelos del proyecto ya coinciden con los recomendados por el kit.")
        return 0

    print("Modelos recomendados por el kit que cambiarían en equipo/config.json:\n")
    for herramienta, filas in cambios.items():
        print(f"  {herramienta}")
        ancho = max(len(c) for c, _, _ in filas)
        for clave, antes, despues in filas:
            print(f"    {clave:<{ancho}}  {antes}  →  {despues}")
        print()
    print("Se conserva el resto de la configuración (herramientas, agente principal, temperaturas, exclusiones).")

    if "--si" not in args:
        try:
            respuesta = input("¿Aplicar estos cambios? [s/N] ").strip().lower()
        except EOFError:
            respuesta = ""
        if respuesta not in ("s", "si", "sí", "y", "yes"):
            print("Sin cambios.")
            return 0

    for herramienta, kit_h in (cfg_kit.get("modelos") or {}).items():
        destino = cfg_proy.setdefault("modelos", {}).setdefault(herramienta, {})
        for clave in CLAVES:
            if clave in (kit_h or {}):
                destino[clave] = kit_h[clave]
            else:
                destino.pop(clave, None)
    CONFIG_PROYECTO.write_text(json.dumps(cfg_proy, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print("equipo/config.json actualizado. Regenerando la configuración de agentes…")
    r = subprocess.run([sys.executable, "scripts/sincronizar.py"], cwd=RAIZ)
    if r.returncode == 0:
        print("Listo. Reinicia tu herramienta (ej. OpenCode) y haz commit de los cambios.")
    return r.returncode


if __name__ == "__main__":
    sys.exit(main())
