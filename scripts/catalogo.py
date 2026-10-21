#!/usr/bin/env python3
"""Catálogo de skills de tecnología del kit: cuáles hay, qué tan probadas están y cuáles tiene el proyecto.

Las skills de tecnología (cómo se escribe código en Go, en React, en PostgreSQL…) viven en el kit, en
catalogo/skills/, y no se copian todas a los proyectos: `make install-kit` lleva a .agents/skills/ solo las
que nombra el perfil (equipo/perfil.json). Sin perfil, ninguna.

Cada skill del catálogo es una carpeta con su SKILL.md, que declara su madurez en los metadatos:

  ---
  name: go-backend
  description: ...
  metadata:
    madurez: probada        probada: usada en un proyecto real · redactada: escrita, todavía sin usar en uno
  ---

Uso:
  python3 scripts/catalogo.py              muestra el catálogo y qué tiene el proyecto
  python3 scripts/catalogo.py --verificar  falla si una skill del catálogo está mal escrita

Solo usa la biblioteca estándar de Python 3.9+.
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

RAIZ = Path(__file__).resolve().parent.parent
CARPETA = "catalogo/skills"      # dentro del kit
INSTALADAS = ".agents/skills"    # dentro del proyecto
MADUREZ = {
    "probada": "usada en un proyecto real",
    "redactada": "solo redactada: todavía no se usó en un proyecto real",
}
NOMBRE = re.compile(r"[a-z0-9]+(-[a-z0-9]+)*")


def carpeta(raiz: Path = RAIZ) -> Path | None:
    """Dónde está el catálogo: en un proyecto, dentro del submódulo del kit; en el repositorio del kit, en la raíz."""
    manifiesto = raiz / ".kit-manifest.json"
    if manifiesto.exists():
        try:
            ruta = json.loads(manifiesto.read_text(encoding="utf-8")).get("ruta_kit") or ""
        except (ValueError, OSError, AttributeError):
            ruta = ""
        candidata = raiz / ruta / CARPETA
        return candidata if ruta and candidata.is_dir() else None
    candidata = raiz / CARPETA
    return candidata if candidata.is_dir() else None


def leer(donde: Path | None) -> tuple[dict[str, dict], list[str]]:
    """({nombre: {"madurez", "descripcion", "carpeta"}}, errores). Una skill con errores no entra al catálogo."""
    skills: dict[str, dict] = {}
    errores: list[str] = []
    if donde is None or not donde.is_dir():
        return skills, errores
    for sub in sorted(p for p in donde.iterdir() if p.is_dir()):
        quien = f"{CARPETA}/{sub.name}"
        archivo = sub / "SKILL.md"
        if not archivo.is_file():
            errores.append(f"{quien}: falta SKILL.md")
            continue
        texto = archivo.read_text(encoding="utf-8").replace("\r\n", "\n")
        if not texto.startswith("---\n") or texto.count("---\n") < 2:
            errores.append(f"{quien}/SKILL.md: falta el bloque de metadatos (---)")
            continue
        cabecera = texto.split("---\n", 2)[1]
        datos: dict[str, str] = {}
        seccion = ""
        for linea in cabecera.splitlines():
            clave, _, valor = linea.partition(":")
            if not linea.startswith((" ", "\t")):
                seccion = clave.strip()
                datos[seccion] = valor.strip()
            elif seccion == "metadata":
                datos["metadata." + clave.strip()] = valor.strip().strip("\"'")
        madurez = datos.get("metadata.madurez", "")
        malos = []
        if not NOMBRE.fullmatch(sub.name):
            malos.append("el nombre de la carpeta solo admite minúsculas, números y guiones")
        if datos.get("name") != sub.name:
            malos.append(f"«name» debe ser igual al nombre de la carpeta ({sub.name})")
        if not datos.get("description"):
            malos.append("falta «description»")
        if madurez not in MADUREZ:
            malos.append(f"falta la madurez en los metadatos (metadata: madurez: {' | '.join(MADUREZ)})")
        if malos:
            errores += [f"{quien}/SKILL.md: {m}" for m in malos]
            continue
        skills[sub.name] = {"madurez": madurez, "descripcion": datos["description"], "carpeta": sub}
    return skills, errores


def main() -> int:
    import perfil as perfil_proyecto     # aquí y no arriba: perfil.py también usa este archivo

    donde = carpeta()
    if donde is None:
        print("No se encontró el catálogo de skills del kit (¿submódulo sin inicializar? git submodule update --init).",
              file=sys.stderr)
        return 1
    skills, errores = leer(donde)
    for e in errores:
        print(f"✗ {e}", file=sys.stderr)
    if "--verificar" in sys.argv[1:]:
        if not errores:
            print(f"OK: {len(skills)} skills en el catálogo.")
        return 1 if errores else 0

    en_el_kit = donde == RAIZ / CARPETA
    try:
        perfil = perfil_proyecto.cargar()
    except perfil_proyecto.ErrorPerfil:
        perfil = None
    pedidas = [] if perfil is None else perfil_proyecto.skills_nombradas(perfil)

    print(f"Catálogo de skills del kit ({donde.relative_to(RAIZ).as_posix()}/): {len(skills)}\n")
    ancho = max([len(n) for n in skills] + [5])
    for nombre, s in skills.items():
        if en_el_kit:
            estado = ""
        elif (RAIZ / INSTALADAS / nombre / "SKILL.md").exists():
            estado = "en el proyecto" if nombre in pedidas else "en el proyecto (ya no se pide: make install-kit la retira)"
        else:
            estado = "la pide el perfil: make install-kit" if nombre in pedidas else "—"
        print(f"  {nombre:<{ancho}}  {s['madurez']:<10} {estado}".rstrip())
        print(f"  {'':<{ancho}}  {s['descripcion']}")
    print()
    for clave, significado in MADUREZ.items():
        print(f"  {clave}: {significado}")
    if not en_el_kit:
        print(f"\nAl proyecto llegan con `make install-kit` las que nombra su perfil ({perfil_proyecto.RUTA}).")
        if perfil is None and not (RAIZ / perfil_proyecto.RUTA).exists():
            print("Este proyecto todavía no tiene perfil. Para crearlo, escribe en el chat: /bowser-profile")
        faltan = [s for s in pedidas if s not in skills and not (RAIZ / INSTALADAS / s / "SKILL.md").exists()]
        for s in faltan:
            print(f"⚠ el perfil pide la skill «{s}», que no está en el catálogo ni en {INSTALADAS}/", file=sys.stderr)
    return 1 if errores else 0


if __name__ == "__main__":
    sys.exit(main())
