#!/usr/bin/env python3
"""Ejecuta los verbos del perfil del proyecto (equipo/perfil.json).

El núcleo del kit solo sabe pedir verbos (formato, revisar, probar, cobertura, auditar, generar). Qué comando
corresponde a cada uno lo declara cada parte del proyecto en su perfil. Este script lo busca y lo ejecuta
dentro de la carpeta de la parte, con las variables PERFIL_RAIZ (raíz del proyecto) y PERFIL_TERCEROS (carpetas
de terceros de la parte, una por línea). Una parte que no define un verbo se omite con un aviso: no es un error.

Uso:
  python3 scripts/verbos.py probar                 el verbo en todas las partes
  python3 scripts/verbos.py formato revisar        varios verbos, en orden
  python3 scripts/verbos.py probar --parte api     solo en una parte
  python3 scripts/verbos.py generar --verificar    además falla si la parte quedó con cambios sin commit
  python3 scripts/verbos.py formato probar --resumen   al final, qué se comprobó y qué quedó sin comprobar (make ci)
  python3 scripts/verbos.py --pre-commit           controles del perfil en cada commit (lo llama el hook de git)
  python3 scripts/verbos.py --inmutables origin/main   falla si la rama modifica archivos inmutables (integración continua)
  python3 scripts/verbos.py --inmutable <archivo>  termina con 1 si el archivo es inmutable y ya está versionado (hook del agente)

Solo usa la biblioteca estándar de Python 3.9+.
"""
from __future__ import annotations

import fnmatch
import os
import subprocess
import sys

from perfil import RAIZ, RUTA, VERBOS, ErrorPerfil, cargar, comando_de, perfil_valido, validar


def correr(comando: str, parte: dict) -> int:
    """Ejecuta el comando dentro de la carpeta de la parte. Le pasa la raíz del proyecto y las carpetas de
    terceros de la parte (una por línea), para que no tenga que repetirlas."""
    sys.stdout.flush()
    sys.stderr.flush()
    entorno = dict(os.environ, PERFIL_RAIZ=str(RAIZ),
                   PERFIL_TERCEROS="\n".join(t.rstrip("/") for t in parte.get("terceros", [])))
    return subprocess.run(["bash", "-o", "pipefail", "-c", comando], cwd=RAIZ / parte["carpeta"], env=entorno).returncode


def git(*args: str) -> list[str]:
    r = subprocess.run(["git", *args], cwd=RAIZ, capture_output=True)
    return [l for l in r.stdout.decode("utf-8", "replace").splitlines() if l]


def dentro_de(archivo: str, carpeta: str) -> str | None:
    """La ruta del archivo relativa a la carpeta de la parte, o None si no le pertenece."""
    carpeta = os.path.normpath(carpeta).replace("\\", "/")
    if carpeta == ".":
        return archivo
    return archivo[len(carpeta) + 1:] if archivo.startswith(carpeta + "/") else None


def es_de_terceros(rel: str, parte: dict) -> bool:
    return any(rel == t.rstrip("/") or rel.startswith(t.rstrip("/") + "/") for t in parte.get("terceros", []))


# ---------------------------------------------------------------- ejecutar verbos

def resumir(verbos: list[str], resultados: dict[str, dict[str, list[str]]]) -> None:
    """Qué se comprobó y qué no. Que nada falle no significa que todo esté comprobado: un verbo sin definir
    no se ejecutó."""
    sys.stderr.flush()
    print("\nResumen de lo comprobado")
    total = hechas = 0
    for verbo in verbos:
        r = resultados[verbo]
        total += sum(len(nombres) for nombres in r.values())
        hechas += len(r["bien"]) + len(r["falló"])
        if not r["bien"] and not r["falló"]:
            print(f"  {verbo:<10} sin definir en ninguna parte")
            continue
        grupos = [f"{titulo}: {', '.join(r[clave])}"
                  for clave, titulo in (("falló", "FALLÓ"), ("bien", "bien"), ("sin definir", "sin definir")) if r[clave]]
        print(f"  {verbo:<10} {'  ·  '.join(grupos)}")
    if any(resultados[v]["falló"] for v in verbos):
        return
    if hechas < total:
        print(f"⚠ Pasó lo que se comprobó: {hechas} de {total} comprobaciones. Las otras {total - hechas} no existen "
              f"todavía en este proyecto (arriba, «sin definir»): nadie las revisó.")
    else:
        print(f"✓ Todo comprobado y sin fallos ({total} comprobaciones).")


def ejecutar(perfil: dict, verbos: list[str], solo: str | None, verificar: bool, resumen: bool = False) -> int:
    partes = [p for p in perfil["partes"] if solo in (None, p["nombre"])]
    if not partes:
        nombres = ", ".join(p["nombre"] for p in perfil["partes"])
        print(f"✗ No hay ninguna parte llamada «{solo}» en {RUTA} (partes: {nombres}).", file=sys.stderr)
        return 2

    fallos: list[str] = []
    resultados = {v: {"bien": [], "falló": [], "sin definir": []} for v in verbos}
    for verbo in verbos:
        definido = False
        for parte in partes:
            comando = comando_de(parte, verbo)
            if not comando:
                print(f"⚠ La parte «{parte['nombre']}» no define el verbo «{verbo}»: se omite.")
                resultados[verbo]["sin definir"].append(parte["nombre"])
                continue
            definido = True
            print(f"▶ {verbo} · {parte['nombre']} ({parte['carpeta']}): {comando}")
            codigo = correr(comando, parte)
            if codigo != 0:
                print(f"✗ {verbo} falló en la parte «{parte['nombre']}» (código {codigo}).", file=sys.stderr)
                fallos.append(f"{verbo} en «{parte['nombre']}»")
                resultados[verbo]["falló"].append(parte["nombre"])
                continue
            if verificar and verbo == "generar":
                cambios = git("status", "--porcelain", "--", parte["carpeta"])
                if cambios:
                    print(f"✗ El código generado de la parte «{parte['nombre']}» no estaba al día (o hay cambios sin commit):",
                          file=sys.stderr)
                    for c in cambios:
                        print(f"  {c}", file=sys.stderr)
                    print("  Ejecuta 'make generate' y agrega el resultado al commit.", file=sys.stderr)
                    fallos.append(f"código generado de «{parte['nombre']}»")
                    resultados[verbo]["falló"].append(parte["nombre"])
                    continue
                print(f"✓ Código generado al día en la parte «{parte['nombre']}».")
            resultados[verbo]["bien"].append(parte["nombre"])
        if not definido:
            print(f"⚠ Ninguna parte define «{verbo}» ({VERBOS[verbo]}). Se agrega en {RUTA}.")

    if resumen:
        resumir(verbos, resultados)
    if fallos:
        print(f"\n✗ Falló: {'; '.join(fallos)}.", file=sys.stderr)
        return 1
    return 0


# ---------------------------------------------------------------- controles del commit

def inmutables_tocados(perfil: dict, tocados: list[str]) -> list[str]:
    """Un mensaje por cada parte con archivos inmutables (ej. migraciones) entre los modificados o eliminados."""
    mensajes = []
    for parte in perfil["partes"]:
        patrones = parte.get("inmutables", [])
        violan = []
        for archivo in tocados:
            rel = dentro_de(archivo, parte["carpeta"])
            if rel is not None and any(fnmatch.fnmatchcase(rel, patron) for patron in patrones):
                violan.append(archivo)
        if violan:
            mensajes.append(f"Estos archivos de la parte «{parte['nombre']}» no se modifican una vez versionados; "
                            "crea uno nuevo: " + " ".join(violan))
    return mensajes


def inmutables_desde(base: str) -> int:
    """Para la integración continua: lo mismo que revisa el commit, sobre todo lo que la rama cambió desde `base`.
    Sin perfil no hay archivos inmutables declarados."""
    perfil = perfil_valido()
    if perfil is None:
        print(f"Este proyecto no tiene perfil ({RUTA}): no declara archivos inmutables.")
        return 0
    r = subprocess.run(["git", "diff", "--name-only", "--no-renames", "--diff-filter=MD", f"{base}...HEAD"],
                       cwd=RAIZ, capture_output=True)
    if r.returncode != 0:
        print(f"✗ No se pudo comparar con «{base}»: {r.stderr.decode('utf-8', 'replace').strip()}", file=sys.stderr)
        return 2
    mensajes = inmutables_tocados(perfil, [l for l in r.stdout.decode("utf-8", "replace").splitlines() if l])
    for m in mensajes:
        print(f"::error::{m}" if os.environ.get("GITHUB_ACTIONS") else f"✗ {m}", file=sys.stderr)
    if not mensajes:
        print("OK: ningún archivo inmutable modificado.")
    return 1 if mensajes else 0


def es_inmutable(archivo: str) -> int:
    """Para el hook del agente (Claude Code): 1 si el archivo es inmutable y ya está versionado, con el motivo;
    0 si se puede editar. Sin perfil no hay inmutables. Con un perfil que no se puede leer no se bloquea nada:
    lo dirá el commit, y bloquear aquí impediría hasta corregir el perfil."""
    try:
        perfil = cargar()
        if perfil is None or validar(perfil)[0]:
            return 0
    except ErrorPerfil:
        return 0
    if os.path.isabs(archivo):
        archivo = os.path.relpath(os.path.realpath(archivo), os.path.realpath(RAIZ))
    archivo = archivo.replace("\\", "/")
    mensajes = inmutables_tocados(perfil, [archivo])
    if not mensajes:
        return 0
    versionado = subprocess.run(["git", "ls-files", "--error-unmatch", "--", archivo], cwd=RAIZ, capture_output=True)
    if versionado.returncode != 0:      # todavía no está en git: se puede escribir y corregir
        return 0
    print(mensajes[0], file=sys.stderr)
    return 1


def pre_commit() -> int:
    """Controles del perfil en cada commit. Escribe un «✗» por problema y termina con 1 si hay alguno."""
    errores = 0

    def fallar(mensaje: str) -> None:
        nonlocal errores
        print(f"✗ {mensaje}", file=sys.stderr)
        errores += 1

    # 1. Los comandos del perfil se ejecutan en la máquina de cada persona y en la integración continua:
    #    un cambio en él lo confirma una persona.
    if RUTA in git("diff", "--cached", "--name-only", "--no-renames") and "1" not in (os.environ.get("APPROVED_PROFILE"), os.environ.get("APROBADO_PERFIL")):   # el segundo: nombre anterior
        fallar(f"El perfil del proyecto ({RUTA}) cambió. Sus comandos se ejecutan en tu máquina y en la integración "
               "continua: revísalo (make profile) y confirma con: APPROVED_PROFILE=1 git commit ...")

    try:
        perfil = cargar()
    except ErrorPerfil as e:
        fallar(str(e))
        return 1
    if perfil is None:          # sin perfil (o el commit lo está eliminando): no declara nada que revisar
        return 1 if errores else 0
    problemas, _ = validar(perfil)
    for p in problemas:
        fallar(f"{RUTA}: {p}")
    if problemas:
        return 1

    # 2. Archivos que no se modifican una vez versionados (ej. migraciones): se crea uno nuevo.
    tocados = git("diff", "--cached", "--name-only", "--no-renames", "--diff-filter=MD")
    for mensaje in inmutables_tocados(perfil, tocados):
        fallar(mensaje)

    # 3. Formato, solo en las partes que el commit toca.
    nuevos = git("diff", "--cached", "--name-only", "--no-renames", "--diff-filter=ACM")
    for parte in perfil["partes"]:
        comando = comando_de(parte, "formato")
        if not comando:
            continue
        propios = [rel for rel in (dentro_de(a, parte["carpeta"]) for a in nuevos)
                   if rel is not None and not es_de_terceros(rel, parte)]
        if propios and correr(comando, parte) != 0:
            fallar(f"La parte «{parte['nombre']}» no pasa la revisión de formato ({comando}).")

    return 1 if errores else 0


# ---------------------------------------------------------------- principal

def main() -> int:
    args = sys.argv[1:]
    if "--pre-commit" in args:
        return pre_commit()
    if "--inmutables" in args:
        i = args.index("--inmutables")
        if i + 1 >= len(args):
            print("Uso: scripts/verbos.py --inmutables <rama o commit de base>", file=sys.stderr)
            return 2
        return inmutables_desde(args[i + 1])
    if "--inmutable" in args:
        i = args.index("--inmutable")
        if i + 1 >= len(args):
            print("Uso: scripts/verbos.py --inmutable <archivo>", file=sys.stderr)
            return 2
        return es_inmutable(args[i + 1])

    solo = None
    if "--parte" in args:
        i = args.index("--parte")
        if i + 1 >= len(args):
            print("Uso: scripts/verbos.py <verbo>... [--parte <nombre>] [--verificar]", file=sys.stderr)
            return 2
        solo = args[i + 1]
        del args[i:i + 2]
    verificar = "--verificar" in args
    verbos = [a for a in args if not a.startswith("--")]
    desconocidos = [v for v in verbos if v not in VERBOS]
    if not verbos or desconocidos:
        if desconocidos:
            print(f"✗ Verbo desconocido: {', '.join(desconocidos)}.", file=sys.stderr)
        print("Uso: scripts/verbos.py <verbo>... [--parte <nombre>] [--verificar]", file=sys.stderr)
        print("Verbos:", file=sys.stderr)
        for verbo, que in VERBOS.items():
            print(f"  {verbo:<10} {que}", file=sys.stderr)
        return 2

    perfil = perfil_valido()
    if perfil is None:
        print(f"✗ Este proyecto no tiene perfil ({RUTA}). Para crearlo, escribe en el chat: /bowser-profile", file=sys.stderr)
        return 1
    return ejecutar(perfil, verbos, solo, verificar, "--resumen" in args)


if __name__ == "__main__":
    sys.exit(main())
