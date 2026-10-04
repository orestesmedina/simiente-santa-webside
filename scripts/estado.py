#!/usr/bin/env python3
"""Muestra por dónde va el proyecto: roadmap, funcionalidad actual, fase, aprobaciones, tareas y próximo paso.

Lee (no modifica nada):
  docs/producto/roadmap.md          tabla con columna "Estado"
  specs/**/estado.md                lo mantiene el orquestador (plantilla: docs/plantillas/estado.md)
  specs/<rama>/tasks.md             tareas "- [X]" / "- [ ]"
  specs/<rama>/revision-*.md        reportes de validación
  git                               rama actual, último commit, cambios sin commit

Uso (desde la raíz del proyecto):
  python3 scripts/estado.py          resumen completo   (make estado)
  python3 scripts/estado.py --todo   incluye también las funcionalidades terminadas del roadmap
"""
from __future__ import annotations

import re
import subprocess
import sys
import time
import unicodedata
from pathlib import Path

RAIZ = Path.cwd()
ROADMAP = RAIZ / "docs/producto/roadmap.md"
SPECS = RAIZ / "specs"
FASES = {1: "Especificar", 2: "Aclarar", 3: "Planificar", 4: "Tareas", 5: "Coherencia",
         6: "Implementar", 7: "Validar", 8: "Converger", 9: "Entregar"}
ABIERTOS = ("en curso", "en revision", "pausada")
MOSTRAR = {"en revision": "en revisión"}
PLURAL = {"pendiente": "pendientes", "pausada": "pausadas", "terminada": "terminadas"}


def etiqueta(estado: str, n: int = 1) -> str:
    texto = MOSTRAR.get(estado, estado)
    return PLURAL.get(estado, texto) if n > 1 else texto


# ---------- utilidades ----------

def normal(texto: str) -> str:
    """minúsculas y sin tildes, para comparar."""
    texto = unicodedata.normalize("NFD", texto or "")
    return "".join(c for c in texto if unicodedata.category(c) != "Mn").strip().lower()


def git(*args: str) -> str:
    try:
        r = subprocess.run(["git", *args], capture_output=True, text=True, cwd=RAIZ)
        if r.returncode != 0:
            return ""
        # Sin quitar espacios iniciales: en `git status --porcelain` la primera columna puede ser un espacio.
        return r.stdout.rstrip() if args[:1] == ("status",) else r.stdout.strip()
    except OSError:
        return ""


def rel(p: Path) -> str:
    try:
        return str(p.relative_to(RAIZ))
    except ValueError:
        return str(p)


def hace(ts: float | None) -> str:
    if not ts:
        return ""
    dias = int((time.time() - ts) // 86400)
    return "hoy" if dias <= 0 else ("ayer" if dias == 1 else f"hace {dias} días")


def tablas(texto: str) -> list[list[list[str]]]:
    """Devuelve las tablas markdown del texto como listas de filas (sin la fila separadora)."""
    resultado, actual = [], []
    for linea in texto.splitlines():
        s = linea.strip()
        if s.startswith("|") and s.endswith("|"):
            celdas = [c.strip() for c in s[1:-1].split("|")]
            if all(re.fullmatch(r":?-{2,}:?", c) for c in celdas if c):
                continue
            actual.append(celdas)
        elif actual:
            resultado.append(actual)
            actual = []
    if actual:
        resultado.append(actual)
    return resultado


def seccion(texto: str, titulo: str) -> str:
    m = re.search(rf"^##\s+{re.escape(titulo)}\s*$(.*?)(?=^##\s|\Z)", texto, re.M | re.S | re.I)
    return m.group(1) if m else ""


def sin_comentarios(texto: str) -> str:
    return re.sub(r"<!--.*?-->", "", texto, flags=re.S)


def ts_archivo(p: Path, modificados: set[str]) -> float:
    """Momento del último cambio: el del disco si tiene cambios sin commit; si no, el del último commit."""
    r = rel(p)
    if r in modificados:
        return p.stat().st_mtime
    ts = git("log", "-1", "--format=%ct", "--", r)
    return float(ts) if ts else p.stat().st_mtime


# ---------- lectura ----------

def leer_roadmap() -> list[dict[str, str]]:
    if not ROADMAP.exists():
        return []
    for tabla in tablas(sin_comentarios(ROADMAP.read_text(encoding="utf-8"))):
        cab = [normal(c) for c in tabla[0]]
        if "estado" not in cab:
            continue
        filas = []
        for fila in tabla[1:]:
            d = {cab[i]: (fila[i] if i < len(fila) else "") for i in range(len(cab))}
            filas.append({
                "num": d.get("#", ""),
                "nombre": d.get("funcionalidad", "") or (fila[1] if len(fila) > 1 else ""),
                "estado": normal(d.get("estado", "")) or "pendiente",
                "rama": d.get("rama / pr", "") or d.get("rama", ""),
            })
        return filas
    return []


def leer_estado(p: Path, texto: str | None = None) -> dict:
    texto = sin_comentarios(texto if texto is not None else p.read_text(encoding="utf-8"))
    datos: dict = {"ruta": p, "carpeta": p.parent, "campos": {}, "aprobaciones": {}, "hallazgos": []}
    for tabla in tablas(seccion(texto, "Resumen")):
        for fila in tabla:
            if len(fila) >= 2 and normal(fila[0]) != "campo":
                datos["campos"][normal(fila[0])] = fila[1]
    for tabla in tablas(seccion(texto, "Aprobaciones")):
        for fila in tabla[1:]:
            if fila and fila[0]:
                datos["aprobaciones"][fila[0]] = {
                    "estado": normal(fila[1]) if len(fila) > 1 else "",
                    "quien": fila[2] if len(fila) > 2 else "",
                    "fecha": fila[3] if len(fila) > 3 else "",
                }
    for linea in seccion(texto, "Hallazgos abiertos").splitlines():
        s = linea.strip()
        if s.startswith(("- ", "* ")) and not re.match(r"[-*]\s*\(?ningun", normal(s)):
            datos["hallazgos"].append(s[2:].strip())
    return datos


def num_fase(texto: str) -> int:
    m = re.match(r"\s*(\d+)", texto or "")
    return int(m.group(1)) if m else 0


def tareas(carpeta: Path, texto: str | None = None) -> tuple[int, int, str]:
    """(hechas, total, primera pendiente)"""
    p = carpeta / "tasks.md"
    if texto is None:
        if not p.exists():
            return 0, 0, ""
        texto = p.read_text(encoding="utf-8")
    hechas = total = 0
    siguiente = ""
    for linea in texto.splitlines():
        m = re.match(r"\s*[-*]\s*\[([ xX])\]\s*(.*)", linea)
        if not m:
            continue
        total += 1
        if m.group(1) in "xX":
            hechas += 1
        elif not siguiente:
            siguiente = m.group(2).strip()
    return hechas, total, siguiente


def fase_deducida(carpeta: Path) -> int:
    hechas, total, _ = tareas(carpeta)
    if list(carpeta.glob("revision-*.md")):
        return 7
    if total and hechas == total:
        return 7
    if hechas:
        return 6
    if (carpeta / "tasks.md").exists():
        return 4
    if (carpeta / "plan.md").exists():
        return 3
    if (carpeta / "spec.md").exists():
        return 1
    return 0


# ---------- informe ----------

def main() -> int:
    todo = "--todo" in sys.argv[1:]
    avisos: list[str] = []
    rama = git("branch", "--show-current")
    try:
        ruta_kit = subprocess.run(["bash", "scripts/ruta-kit.sh"], capture_output=True, text=True, cwd=RAIZ).stdout.strip()
    except OSError:
        ruta_kit = ""
    ruta_kit = ruta_kit or ".bowser-spec-kit-ai"
    sucios = [l for l in git("status", "--porcelain").splitlines() if l.strip()]
    if any(l[3:].strip() == ruta_kit for l in sucios):
        sucios = [l for l in sucios if l[3:].strip() != ruta_kit]
        avisos.append(f"El submódulo {ruta_kit}/ no está en la versión de esta rama "
                      "(pasa al cambiar de rama). Ejecuta: git submodule update")
    modificados = {l[3:].split(" -> ")[-1].strip('"') for l in sucios}

    # Roadmap
    roadmap = leer_roadmap()
    print("PROYECTO")
    if not ROADMAP.exists():
        print("  Sin roadmap (docs/producto/roadmap.md). Plantilla: docs/plantillas/roadmap.md")
    elif not roadmap:
        print("  El roadmap no tiene una tabla con columna 'Estado'. Formato: docs/plantillas/roadmap.md")
    else:
        conteo: dict[str, int] = {}
        for f in roadmap:
            conteo[f["estado"]] = conteo.get(f["estado"], 0) + 1
        terminadas = conteo.get("terminada", 0)
        resto = " · ".join(f"{n} {etiqueta(e, n)}" for e, n in conteo.items() if e != "terminada")
        print(f"  Roadmap: {terminadas}/{len(roadmap)} terminadas" + (f" · {resto}" if resto else ""))
        for f in roadmap:
            if f["estado"] in ABIERTOS or (todo and f["estado"] == "terminada"):
                extra = f"  ({f['rama']})" if f["rama"] else ""
                print(f"    {etiqueta(f['estado']):<12} {f['num']}. {f['nombre']}{extra}")
        siguiente = next((f for f in roadmap if f["estado"] == "pendiente"), None)
        if siguiente:
            print(f"  Siguiente pendiente: {siguiente['num']}. {siguiente['nombre']}")

    # Estados de funcionalidades
    estados = [leer_estado(p) for p in sorted(SPECS.glob("**/estado.md"))] if SPECS.exists() else []

    def de_rama(e: dict) -> str:
        return e["campos"].get("rama", "") or e["carpeta"].name

    actual = next((e for e in estados if rama and (de_rama(e) == rama or e["carpeta"].name == rama)), None)
    carpeta_actual = actual["carpeta"] if actual else (SPECS / rama if rama and (SPECS / rama).is_dir() else None)

    print()
    print(f"RAMA ACTUAL: {rama or '(sin git)'}")
    if carpeta_actual is None:
        if rama in ("main", "master", ""):
            print("  Rama principal: no hay una funcionalidad en curso en esta rama.")
        else:
            print(f"  No hay specs/{rama}/ ni un estado.md con 'Rama | {rama}'.")
    else:
        hechas, total, sig_tarea = tareas(carpeta_actual)
        if actual:
            c = actual["campos"]
            print(f"  Estado:        {rel(actual['ruta'])}")
            print(f"  Flujo:         {c.get('flujo', '—')}")
            print(f"  Fase:          {c.get('fase', '—')}")
        else:
            f = fase_deducida(carpeta_actual)
            print(f"  ⚠ Sin estado.md: fase deducida de los archivos (aprobaciones desconocidas).")
            print(f"  Fase:          {f}/9 · {FASES.get(f, '—')} (estimada)" if f else "  Fase:          —")
            avisos.append(f"Crea specs/{carpeta_actual.name}/estado.md: pide al orquestador \"retomemos\" (skill equipo-retomar).")
        if total:
            print(f"  Tareas:        {hechas}/{total} hechas" + (f" · siguiente: {sig_tarea[:90]}" if sig_tarea else ""))
        costos_json = carpeta_actual / "costos.json"
        if costos_json.exists():
            try:
                import json
                c = json.loads(costos_json.read_text(encoding="utf-8"))
                tot = c.get("totales") or {}
                print(f"  Costo IA:      ${tot.get('costo', 0):,.2f} equivalente · {tot.get('tokens', 0) / 1e6:.1f} M tokens"
                      f" ({c.get('estado', '?')}) · detalle: make costos")
            except ValueError:
                pass
        revisiones = sorted(carpeta_actual.glob("revision-*.md"))
        if revisiones:
            print(f"  Última revisión: {revisiones[-1].name}")
        if actual:
            c = actual["campos"]
            aprob = []
            for puerta, a in actual["aprobaciones"].items():
                marca = "✓" if a["estado"].startswith("aprobad") else ("?" if "confirmar" in a["estado"] else "—")
                detalle = f" ({a['quien']}, {a['fecha']})" if marca == "✓" and a["quien"] else ""
                aprob.append(f"{puerta} {marca}{detalle}")
            if aprob:
                print(f"  Aprobaciones:  {' · '.join(aprob)}")
            print(f"  Corrección:    ciclo {c.get('ciclo de correccion', '—')} · hallazgos abiertos: {len(actual['hallazgos'])}")
            for h in actual["hallazgos"][:5]:
                print(f"                 - {h[:100]}")
            bloqueo = c.get("bloqueado por", "")
            if bloqueo and bloqueo not in ("—", "-"):
                print(f"  Bloqueado por: {bloqueo}")
            print(f"  Próximo paso:  {c.get('proximo paso', '—')}")
            ts_est = ts_archivo(actual["ruta"], modificados)
            print(f"  Actualizado:   {c.get('actualizado', '—')} (archivo modificado {hace(ts_est)})")

            # Coherencia entre estado y evidencia
            fase = num_fase(c.get("fase", ""))
            for nombre in ("spec.md", "plan.md", "tasks.md"):
                p = carpeta_actual / nombre
                if p.exists() and ts_archivo(p, modificados) > ts_est + 60:
                    avisos.append(f"{nombre} cambió después de estado.md: el estado puede estar desactualizado.")
            puertas = {normal(k): v for k, v in actual["aprobaciones"].items()}
            if fase >= 3 and not puertas.get("spec", {}).get("estado", "").startswith("aprobad"):
                avisos.append("La fase es posterior a la spec, pero su aprobación no está registrada.")
            if fase >= 4 and not puertas.get("plan", {}).get("estado", "").startswith("aprobad"):
                avisos.append("La fase es posterior al plan, pero su aprobación no está registrada.")
            if total and hechas == total and fase and fase < 7:
                avisos.append("Todas las tareas están hechas pero la fase indica implementación o antes.")
            if 0 < fase < 6 and hechas:
                avisos.append(f"Hay {hechas} tareas hechas pero la fase indica {fase}/9.")

    # Otras funcionalidades abiertas
    otras = []
    for e in estados:
        if e is actual:
            continue
        fase_txt = e["campos"].get("fase", "")
        r = de_rama(e)
        en_roadmap = next((f for f in roadmap if r and r in f["rama"]), None)
        if "terminad" in normal(fase_txt) or (en_roadmap and en_roadmap["estado"] == "terminada"):
            continue
        otras.append(f"    {r:<32} {fase_txt or '—'} · próximo: {e['campos'].get('proximo paso', '—')[:60]}")
    if otras:
        print()
        print("OTRAS FUNCIONALIDADES SIN TERMINAR")
        print("\n".join(otras))

    for f in roadmap:
        if f["estado"] in ABIERTOS:
            nombre_rama = re.split(r"[\s,(]", f["rama"].strip())[0] if f["rama"].strip() else ""
            if nombre_rama and nombre_rama != (carpeta_actual.name if carpeta_actual else None) and not git("rev-parse", "--verify", "-q", nombre_rama) and not any(de_rama(e) == nombre_rama or e["carpeta"].name == nombre_rama for e in estados):
                avisos.append(f"El roadmap marca '{f['nombre']}' como {etiqueta(f['estado'])} pero no hay estado.md para {nombre_rama}.")

    # Trabajo en otras ramas (el estado vive en la rama de cada funcionalidad hasta el merge)
    vistas = {de_rama(e) for e in estados} | {e["carpeta"].name for e in estados} | {rama}
    refs = git("for-each-ref", "--sort=-committerdate", "--format=%(refname:short)|%(committerdate:relative)",
               "refs/heads", "refs/remotes")
    lineas_ramas, vistas_ramas = [], set()
    for linea in refs.splitlines():
        ref, _, cuando = linea.partition("|")
        nombre = ref.split("/", 1)[1] if ref.startswith(("origin/", "upstream/")) else ref
        if nombre in vistas or nombre in vistas_ramas or nombre in ("HEAD", "main", "master") or "/" in nombre:
            continue
        texto_estado = git("show", f"{ref}:specs/{nombre}/estado.md")
        if not texto_estado and not git("ls-tree", "--name-only", ref, f"specs/{nombre}/"):
            continue
        vistas_ramas.add(nombre)
        hechas, total, _ = tareas(Path(), git("show", f"{ref}:specs/{nombre}/tasks.md"))
        tareas_txt = f" · {hechas}/{total} tareas" if total else ""
        if texto_estado:
            e = leer_estado(Path(f"specs/{nombre}/estado.md"), texto_estado)
            if "terminad" in normal(e["campos"].get("fase", "")):
                continue
            lineas_ramas.append(f"    {nombre:<32} {e['campos'].get('fase', '—')}{tareas_txt} · último commit {cuando}\n"
                                f"    {'':<32} próximo: {e['campos'].get('proximo paso', '—')[:70]}")
        else:
            lineas_ramas.append(f"    {nombre:<32} (sin estado.md){tareas_txt} · último commit {cuando}")
    if lineas_ramas:
        print()
        print("TRABAJO EN OTRAS RAMAS")
        print("\n".join(lineas_ramas))
        if rama in ("main", "master"):
            print("  Para retomar: git checkout <rama> && make estado")

    if sucios:
        avisos.append(f"{len(sucios)} archivo(s) con cambios sin commit (git status).")

    ultimo = git("log", "-1", "--format=%cr · %s")
    if ultimo:
        print()
        print(f"Último commit: {ultimo}")

    if avisos:
        print()
        print("AVISOS")
        for a in avisos:
            print(f"  ⚠ {a}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
