#!/usr/bin/env python3
"""Revisa lo que le falta a un proyecto para aprovechar la versión 2.0 del kit, y arregla lo mecánico.

Un proyecto que viene de la 1.x queda con pendientes que la actualización no puede resolver sola: crear el perfil,
redactar su integración continua, revisar las reglas heredadas de la constitución… Este script los junta en una
lista. Lo que tiene un solo arreglo posible lo hace él, después de mostrar los cambios y pedir confirmación:

  - los nombres anteriores de los comandos de make, sus opciones y las confirmaciones del commit (`make cobertura`,
    `FORZAR=1`, `APROBADO_PERFIL`), en los workflows del proyecto y en proyecto.mk;
  - los comandos retirados que tienen reemplazo (`make test-backend` → `make test PART=backend`, si esa parte existe);
  - los archivos de scripts/ que el kit dejó de traer y el proyecto conservó: se mueven a tools/ (scripts/ es del kit)
    y se corrige quien los nombra (el perfil, proyecto.mk, los workflows).

Lo demás necesita criterio y lo resuelve un agente (/bowser-profile, /bowser-ci) o una persona. Nunca hace commit.

Uso (desde la raíz del proyecto):
  python3 scripts/pasar_a_2.py             muestra lo pendiente y, en una terminal, ofrece aplicar los arreglos
  python3 scripts/pasar_a_2.py --si        los aplica sin preguntar
  python3 scripts/pasar_a_2.py --resumen   una sola línea, o nada si no falta nada (lo usa make install-kit)

Solo usa la biblioteca estándar de Python 3.9+.
"""
from __future__ import annotations

import difflib
import json
import re
import shutil
import subprocess
import sys
from pathlib import Path

import ci
import perfil as perfil_proyecto

RAIZ = Path.cwd()
MANIFIESTO = ".kit-manifest.json"
PROYECTO_MK = "proyecto.mk"
# Las confirmaciones del commit se llamaban así hasta la 1.21.0. Los comandos y las opciones anteriores se leen del
# Makefile del kit (lista ANTERIORES y las líneas `NUEVA ?= $(ANTERIOR)`), para no repetirlos aquí.
CONFIRMACIONES = {"APROBADO_PERFIL": "APPROVED_PROFILE", "APROBADO_CONSTITUCION": "APPROVED_CONSTITUTION",
                  "APROBADO_COSTOS": "APPROVED_COSTS"}
# Comandos de make retirados en la 1.23.0. Los dos primeros tienen reemplazo si el perfil tiene esa parte.
RETIRADOS = {"test-backend": "backend", "test-frontend": "frontend", "db-migrate": None}
# Lo que la actualización dejó en el proyecto y alguien tiene que revisar.
REGLAS = "equipo/reglas.md"
REGLAS_HEREDADAS = ("Estas reglas venían en la constitución que el kit instalaba antes",
                    "Esta es, completa, la constitución que el proyecto tenía")
GITIGNORE_HEREDADAS = "# Estas líneas venían en el bloque del kit; desde ahora son del proyecto"
# Los grupos del ci.yml que instalaba el kit hasta la 1.17 y que hoy repite bowser.yml.
GRUPOS_REPETIDOS = ("name: Kit y configuración de agentes al día", "name: Migraciones y constitución",
                    "name: Búsqueda de secretos")
# Archivos que el kit gestionaba y dejaron de serlo: excluirlos (kit.excluir) ya no hace nada.
YA_NO_SON_DEL_KIT = (".github/workflows/ci.yml", ".specify/memory/constitution.md")
INVOCACION = re.compile(r"(\bmake\b|\$\(MAKE\))([^\n|;&#]*)")


# ---------------------------------------------------------------- lectura

def leer(ruta: Path) -> str:
    try:
        return ruta.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError):
        return ""


def nombres_anteriores(raiz: Path) -> tuple[dict[str, str], dict[str, str]]:
    """(comandos, opciones): nombre anterior -> nuevo, leídos del Makefile del kit."""
    makefile = leer(raiz / "Makefile").replace("\\\n", " ")
    comandos: dict[str, str] = {}
    m = re.search(r"^ANTERIORES\s*:=(.*)$", makefile, re.MULTILINE)
    for par in (m.group(1).split() if m else []):
        viejo, _, nuevo = par.partition(":")
        if viejo and nuevo:
            comandos[viejo] = nuevo
    opciones = {viejo: nuevo for nuevo, viejo in re.findall(r"^([A-Z]+) \?= \$\(([A-Z]+)\)$", makefile, re.MULTILINE)}
    return comandos, opciones


def kit_excluir(raiz: Path) -> list[str]:
    try:
        excluir = json.loads(leer(raiz / "equipo/config.json") or "{}").get("kit", {}).get("excluir", [])
    except (json.JSONDecodeError, AttributeError):
        return []
    return [e for e in excluir if isinstance(e, str)] if isinstance(excluir, list) else []


def excluido(rel: str, excluir: list[str]) -> bool:
    """El proyecto se quedó con su propia versión de ese archivo del kit (misma regla que instalar_kit.py)."""
    return any(rel == e or (e.endswith("/") and rel.startswith(e)) for e in excluir)


def definidos_en_proyecto_mk(raiz: Path) -> set[str]:
    return set(re.findall(r"^([A-Za-z0-9_-]+)\s*:(?!=)", leer(raiz / PROYECTO_MK), re.MULTILINE))


def que_nombran(raiz: Path, del_kit: dict) -> list[Path]:
    """Los archivos del proyecto donde se nombran comandos y scripts: el perfil, proyecto.mk y sus workflows."""
    rutas = [raiz / perfil_proyecto.RUTA, raiz / PROYECTO_MK, *ci.workflows_del_proyecto(raiz)]
    return [r for r in rutas if r.is_file() and r.relative_to(raiz).as_posix() not in del_kit]


# ---------------------------------------------------------------- arreglos mecánicos

def cambiar_invocaciones(texto: str, cambiar) -> str:
    """Aplica `cambiar` a las palabras de cada llamada a make (`make …` o `$(MAKE) …`), sin tocar lo demás."""
    def una(m: re.Match) -> str:
        return m.group(1) + re.sub(r"[^\s]+", lambda p: cambiar(p.group(0)), m.group(2))
    return "\n".join(INVOCACION.sub(una, linea) if not linea.lstrip().startswith("#") else linea
                     for linea in texto.split("\n"))


def arreglos(raiz: Path, perfil: dict | None, del_kit: dict) -> tuple[dict[Path, str], dict[str, str], list[str]]:
    """(textos nuevos por archivo, scripts a mover origen -> destino, avisos de lo que no tiene arreglo)."""
    comandos, opciones = nombres_anteriores(raiz)
    propios = definidos_en_proyecto_mk(raiz)
    partes = {p["nombre"] for p in (perfil or {}).get("partes", [])}
    avisos: list[str] = []

    mover: dict[str, str] = {}
    excluir = kit_excluir(raiz)
    carpeta = raiz / "scripts"
    for f in sorted(carpeta.rglob("*")) if carpeta.is_dir() else []:
        rel = f.relative_to(raiz).as_posix()
        if f.is_file() and rel not in del_kit and not excluido(rel, excluir) and "__pycache__" not in f.parts:
            destino = "tools/" + f.relative_to(carpeta).as_posix()
            if (raiz / destino).exists():
                avisos.append(f"{rel} no es del kit y debería estar en tools/, pero ya existe {destino}: muévelo a mano.")
            else:
                mover[rel] = destino

    def cambiar(palabra: str) -> str:
        nombre, igual, valor = palabra.partition("=")
        if igual and nombre in opciones:
            return f"{opciones[nombre]}={valor}"
        if not igual and palabra in comandos and palabra not in propios:
            return comandos[palabra]
        if not igual and palabra in RETIRADOS and palabra not in propios and RETIRADOS[palabra] in partes:
            return f"test PART={RETIRADOS[palabra]}"
        return palabra

    nuevos: dict[Path, str] = {}
    for ruta in que_nombran(raiz, del_kit):
        antes = leer(ruta)
        texto = antes
        if ruta.suffix != ".json":
            texto = cambiar_invocaciones(texto, cambiar)
            for viejo, nuevo in CONFIRMACIONES.items():
                texto = re.sub(rf"\b{viejo}\b", nuevo, texto)
        for origen, destino in mover.items():
            texto = texto.replace(origen, destino)
        if texto != antes:
            nuevos[ruta] = texto
        nombre = ruta.relative_to(raiz).as_posix()
        for retirado, parte in RETIRADOS.items():
            if retirado not in propios and re.search(rf"(\bmake\b|\$\(MAKE\))[^\n|;&#]*\s{retirado}\b", texto):
                como = f"usa `make test PART=<parte>` o defínelo" if parte else "defínelo"
                avisos.append(f"{nombre} usa `make {retirado}`, que se retiró en la 1.23.0: {como} en {PROYECTO_MK}.")
    return nuevos, mover, avisos


# ---------------------------------------------------------------- revisión

def revisar(raiz: Path) -> dict:
    """Todo lo pendiente, agrupado por quién lo resuelve."""
    manifiesto = json.loads(leer(raiz / MANIFIESTO) or "{}")
    del_kit = manifiesto.get("archivos", {})
    r = {"agente": [], "persona": [], "avisos": []}

    try:
        perfil = perfil_proyecto.cargar(raiz)
        errores = perfil_proyecto.validar(perfil, raiz)[0] if perfil is not None else []
    except perfil_proyecto.ErrorPerfil as e:
        perfil, errores = None, [str(e)]
    if errores:
        r["persona"].append(f"El perfil ({perfil_proyecto.RUTA}) tiene errores: `make profile` dice qué corregir.")
        perfil = None
    elif perfil is None:
        r["agente"].append("El proyecto no tiene perfil: sin él no funcionan make test, make lint ni make ci, y los "
                           "agentes no saben qué carpeta es de quién. En el chat: /bowser-profile")
    else:
        if not perfil.get("requisitos"):
            r["agente"].append("El perfil no declara `requisitos`: make doctor no sabe qué herramientas necesita una "
                               "máquina para este proyecto. En el chat: /bowser-profile")
        faltan = ci.sin_ejecutar(perfil, raiz)
        if faltan:
            detalle = "; ".join(f"{p['nombre']}: {', '.join(v for v in perfil_proyecto.VERBOS if (p['nombre'], v) in faltan)}"
                                for p in perfil["partes"] if any(n == p["nombre"] for n, _ in faltan))
            r["agente"].append(f"La integración continua no ejecuta estos verbos del perfil ({detalle}). En el chat: /bowser-ci")
        sin_convenciones = [p["nombre"] for p in perfil["partes"]
                            if (p.get("roles") or p.get("rol"))
                            and not any(s.startswith(perfil_proyecto.CONVENCIONES) for r_ in perfil_proyecto.roles_de(p)
                                        for s in r_["skills"])]
        if sin_convenciones:
            cuales = (f"la parte {sin_convenciones[0]} no tiene" if len(sin_convenciones) == 1 else
                      f"las partes {', '.join(sin_convenciones)} no tienen")
            r["avisos"].append(f"Recomendado: {cuales} convenciones propias (cómo nombra y organiza las cosas este "
                               "proyecto). En el chat: /bowser-conventions")

    for f in ci.workflows_del_proyecto(raiz):
        texto = leer(f)
        if sum(g in texto for g in GRUPOS_REPETIDOS) >= 2:
            r["agente"].append(f".github/workflows/{f.name} es el que instalaba el kit: repite lo que ya hace bowser.yml "
                               "(kit, agentes y secretos) y copia los comandos en vez de ejecutar los verbos. "
                               "En el chat: /bowser-ci")
    if any(m in leer(raiz / REGLAS) for m in REGLAS_HEREDADAS):
        r["persona"].append(f"{REGLAS} todavía tiene las reglas que traía la constitución anterior (de Go, React y "
                            "PostgreSQL, o la constitución completa del proyecto). Conserva las que le sirven a este "
                            "proyecto, borra las demás y quita el párrafo que lo explica. El commit pide "
                            "APPROVED_CONSTITUTION=1.")
    gitignore = leer(raiz / ".gitignore")
    if GITIGNORE_HEREDADAS in gitignore:
        r["persona"].append(".gitignore tiene, debajo del bloque del kit, reglas que el kit traía antes (de Go y Node). "
                            "Conserva las que le sirven a este proyecto, borra las demás y quita el comentario que las "
                            "presenta.")
    for e in kit_excluir(raiz):
        if e in YA_NO_SON_DEL_KIT:
            r["avisos"].append(f"`kit.excluir` nombra {e}, que ya no es del kit: esa línea no hace nada. Quítala de "
                               "equipo/config.json.")

    nuevos, mover, avisos = arreglos(raiz, perfil, del_kit)
    r["avisos"] = avisos + r["avisos"]
    r["nuevos"], r["mover"] = nuevos, mover
    return r


def cuantos(r: dict) -> int:
    return len(r["agente"]) + len(r["persona"]) + len(r["avisos"]) + len(r["nuevos"]) + len(r["mover"])


# ---------------------------------------------------------------- salida

def mostrar_arreglos(raiz: Path, r: dict) -> None:
    for origen, destino in r["mover"].items():
        print(f"  {origen} → {destino}")
    for ruta, texto in r["nuevos"].items():
        nombre = ruta.relative_to(raiz).as_posix()
        print(f"  {nombre}")
        for linea in difflib.unified_diff(leer(ruta).splitlines(), texto.splitlines(), n=0, lineterm=""):
            if not linea.startswith(("@@", "---", "+++")):
                print(f"    {linea}")


def aplicar(raiz: Path, r: dict) -> None:
    for ruta, texto in r["nuevos"].items():
        ruta.write_text(texto, encoding="utf-8")
    for origen, destino in r["mover"].items():
        (raiz / destino).parent.mkdir(parents=True, exist_ok=True)
        shutil.move(str(raiz / origen), str(raiz / destino))
    carpeta = raiz / "scripts"
    for vacia in sorted((d for d in carpeta.rglob("*") if d.is_dir()), reverse=True) if carpeta.is_dir() else []:
        if not any(vacia.iterdir()):
            vacia.rmdir()


def resumen(raiz: Path) -> str:
    """Una línea para el final de make install-kit, o nada si el proyecto no tiene pendientes."""
    if not (raiz / MANIFIESTO).exists():
        return ""
    n = cuantos(revisar(raiz))
    if not n:
        return ""
    return (f"Al proyecto le {'queda' if n == 1 else 'quedan'} {n} "
            f"{'pendiente' if n == 1 else 'pendientes'} para aprovechar el kit: make upgrade")


def main() -> int:
    args = sys.argv[1:]
    if not (RAIZ / MANIFIESTO).exists():
        print("Aquí no está instalado el kit (falta .kit-manifest.json): no hay nada que revisar.")
        return 0
    if "--resumen" in args:
        linea = resumen(RAIZ)
        if linea:
            print(f"  ! {linea}")
        return 0

    r = revisar(RAIZ)
    if not cuantos(r):
        print("✓ Nada pendiente: el proyecto aprovecha todo lo que el kit trae.")
        return 0

    print("Lo que le falta al proyecto para aprovechar el kit\n")
    for clave, titulo in (("agente", "Lo resuelve un agente (escribe el comando en el chat)"),
                          ("persona", "Lo resuelve una persona"), ("avisos", "Avisos")):
        if r[clave]:
            print(titulo)
            for linea in r[clave]:
                print(f"  • {linea}")
            print()
    if not (r["nuevos"] or r["mover"]):
        return 0

    print("Arreglos mecánicos (nombres anteriores, comandos retirados y scripts que pasan a tools/)")
    mostrar_arreglos(RAIZ, r)
    print()
    if "--si" not in args and not sys.stdin.isatty():      # un agente o un script: muestra y no pregunta
        print("Para aplicarlos: make upgrade YES=1")
        return 0
    if "--si" not in args:
        try:
            respuesta = input("¿Aplicar estos arreglos? [s/N] ").strip().lower()
        except EOFError:
            respuesta = ""
        if respuesta not in ("s", "si", "sí", "y", "yes"):
            print("Sin cambios. Para aplicarlos sin preguntar: make upgrade YES=1")
            return 0
    aplicar(RAIZ, r)
    print("Arreglos aplicados.")
    if any(ruta.name == "perfil.json" for ruta in r["nuevos"]):
        subprocess.run([sys.executable, "scripts/sincronizar.py"], cwd=RAIZ)
        print(f"Cambió el perfil: el commit pide confirmarlo con APPROVED_PROFILE=1.")
    print("Revisa los cambios con 'git diff' y haz commit.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
