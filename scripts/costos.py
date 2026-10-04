#!/usr/bin/env python3
"""Costo de IA por funcionalidad: tokens y dólares por agente y modelo.

Cada funcionalidad tiene su propio specs/<rama>/costos.json (versionado en git):
  - precios:  historial de precios usados en la tarea; cada cambio se AGREGA con su fecha "desde".
  - sesiones: consumo de cada sesión de la herramienta (OpenCode), por agente y modelo,
              valorizado con el precio vigente el día de cada respuesta (y la tarifa de hora pico).
  - estado:   "abierto" mientras se trabaja; "cerrado" al aprobarse el PR. Un costos.json
              cerrado no se modifica (lo bloquea el pre-commit).

Los dólares son un costo EQUIVALENTE a precio de API: con una suscripción (OpenCode Go) el gasto
real es el % de la consola. Los precios salen de models.dev (el mismo catálogo que usa OpenCode)
y se pueden fijar a mano en equipo/config.json -> "costos": {"precios_manuales": {...}}.

Uso (desde la raíz del proyecto):
  python3 scripts/costos.py              registra lo nuevo en la tarea de la rama actual y muestra su costo
  python3 scripts/costos.py --cerrar     registra lo pendiente y cierra el costo de la tarea
  python3 scripts/costos.py --todo       resumen del proyecto (todas las funcionalidades, también otras ramas)
  python3 scripts/costos.py --hoy        muestra cuánto costaría la tarea a precios de hoy (no guarda nada)
  python3 scripts/costos.py --sin-registrar   solo muestra, no registra
"""
from __future__ import annotations

import datetime as dt
import fnmatch
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import time
import urllib.request
from pathlib import Path

sys.dont_write_bytecode = True  # no dejar scripts/__pycache__ en el proyecto
sys.path.insert(0, str(Path(__file__).resolve().parent))
from estado import git, leer_estado, normal  # noqa: E402

RAIZ = Path.cwd()
SPECS = RAIZ / "specs"
CONFIG = RAIZ / "equipo/config.json"
FUENTE_PRECIOS = os.environ.get("COSTOS_FUENTE_PRECIOS", "https://models.dev/api.json")
DATOS_LOCALES = Path(os.environ.get("XDG_DATA_HOME", Path.home() / ".local/share")) / "bowser-kit"
REGISTRO = DATOS_LOCALES / "costos-registro.json"      # sesiones ya registradas en esta máquina
CACHE_PRECIOS = DATOS_LOCALES / "models-dev.json"
CACHE_MAX_SEG = 3600
MARGEN_INICIO_MS = 30 * 60 * 1000                      # respuestas hasta 30 min antes de crear la rama cuentan
TIPOS = ("input", "output", "reasoning", "cache_read", "cache_write")

# Por defecto; equipo/config.json -> "costos" puede reemplazarlos.
PICO_DEFECTO = [{
    "modelos": "opencode-go/deepseek-*",
    "multiplicador": 2,
    "dias": [0, 1, 2, 3, 4],             # lunes a viernes (UTC)
    "horas_utc": [[1, 4], [6, 10]],
    "nota": "DeepSeek cobra el doble en hora pico (opencode.ai/docs/go)",
}]


class Fallo(Exception):
    pass


# ------------------------------------------------------------------ utilidades

def ahora_ms() -> int:
    return int(time.time() * 1000)


def iso(ms: int | None) -> str:
    if not ms:
        return ""
    return dt.datetime.fromtimestamp(ms / 1000, dt.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def de_iso(texto: str) -> int:
    return int(dt.datetime.strptime(texto, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=dt.timezone.utc).timestamp() * 1000)


def leer_json(p: Path, defecto):
    try:
        return json.loads(p.read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return defecto


def escribir_json(p: Path, datos) -> None:
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(json.dumps(datos, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def config_costos() -> dict:
    cfg = leer_json(CONFIG, {}).get("costos") or {}
    return {
        "precios_manuales": cfg.get("precios_manuales") or {},
        "pico": cfg.get("pico", PICO_DEFECTO),
    }


def modelos_configurados() -> set[str]:
    oc = (leer_json(CONFIG, {}).get("modelos") or {}).get("opencode") or {}
    modelos = {oc.get("orquestador"), oc.get("ligero")}
    modelos |= set((oc.get("niveles") or {}).values()) | set((oc.get("agentes") or {}).values())
    return {m for m in modelos if m}


def fmt_tokens(n: float) -> str:
    n = float(n or 0)
    if n >= 1e6:
        return f"{n / 1e6:.1f} M"
    if n >= 1e3:
        return f"{n / 1e3:.0f} K"
    return f"{n:.0f}"


def fmt_usd(n: float | None) -> str:
    if n is None:
        return "sin precio"
    return f"${n:,.2f}" if n >= 0.01 or n == 0 else f"${n:.4f}"


# ------------------------------------------------------------------ precios

def normalizar_precio(c: dict) -> dict:
    cache = c.get("cache") or {}
    return {
        "input": float(c.get("input") or 0),
        "output": float(c.get("output") or 0),
        "cache_read": float(c.get("cache_read", cache.get("read")) or 0),
        "cache_write": float(c.get("cache_write", cache.get("write")) or 0),
    }


def catalogo_models_dev() -> tuple[dict, str]:
    """Precios por 'proveedor/modelo' (USD por millón de tokens). Usa caché de 1 hora; sin red, la última copia."""
    cache = leer_json(CACHE_PRECIOS, None)
    fresco = cache and time.time() - cache.get("_descargado", 0) < CACHE_MAX_SEG
    if not fresco:
        try:
            if FUENTE_PRECIOS.startswith(("http://", "https://")):
                # Con el User-Agent por defecto de Python, models.dev (Cloudflare) responde 403.
                pedido = urllib.request.Request(FUENTE_PRECIOS, headers={"User-Agent": "bowser-spec-kit-ai (costos.py)"})
                with urllib.request.urlopen(pedido, timeout=15) as r:
                    datos = json.loads(r.read().decode("utf-8"))
            else:
                datos = json.loads(Path(FUENTE_PRECIOS.removeprefix("file://")).read_text(encoding="utf-8"))
            precios = {}
            for prov_id, prov in datos.items():
                for mod_id, mod in ((prov or {}).get("models") or {}).items():
                    if isinstance(mod, dict) and isinstance(mod.get("cost"), dict):
                        precios[f"{prov_id}/{mod_id}"] = normalizar_precio(mod["cost"])
            cache = {"_descargado": time.time(), "precios": precios}
            escribir_json(CACHE_PRECIOS, cache)
        except Exception as e:  # sin red, sitio caído o formato inesperado
            if not cache:
                return {}, f"no se pudo consultar {FUENTE_PRECIOS} ({e.__class__.__name__}) y no hay copia local"
            return cache.get("precios", {}), f"sin conexión con models.dev: se usan los precios guardados el {iso(int(cache['_descargado'] * 1000))[:10]}"
    return cache.get("precios", {}), ""


def precios_vigentes(modelos: set[str]) -> tuple[dict, list[str]]:
    catalogo, aviso = catalogo_models_dev()
    manuales = {k: normalizar_precio(v) for k, v in config_costos()["precios_manuales"].items()}
    avisos = [aviso] if aviso else []
    resultado = {}
    for m in sorted(modelos):
        if m in manuales:
            resultado[m] = dict(manuales[m], fuente="manual")
        elif m in catalogo:
            resultado[m] = dict(catalogo[m], fuente="models.dev")
    return resultado, avisos


def precio_en(costos: dict, modelo: str, ms: int) -> dict | None:
    """Precio de un modelo vigente en un instante: la versión más reciente con desde <= ms (o la primera conocida)."""
    candidatas = [v for v in costos.get("precios", []) if modelo in v.get("modelos", {})]
    if not candidatas:
        return None
    previas = [v for v in candidatas if de_iso(v["desde"]) <= ms]
    version = max(previas, key=lambda v: v["desde"]) if previas else min(candidatas, key=lambda v: v["desde"])
    return version["modelos"][modelo]


def multiplicador_pico(modelo: str, ms: int, reglas: list) -> float:
    momento = dt.datetime.fromtimestamp(ms / 1000, dt.timezone.utc)
    for r in reglas or []:
        if not fnmatch.fnmatch(modelo, r.get("modelos", "")):
            continue
        if momento.weekday() not in r.get("dias", range(7)):
            continue
        if any(a <= momento.hour < b for a, b in r.get("horas_utc", [])):
            return float(r.get("multiplicador", 1))
    return 1.0


def costo(tokens: dict, precio: dict | None, mult: float = 1.0) -> float | None:
    """Misma fórmula que OpenCode: input sin caché, output, lectura y escritura de caché; razonamiento a precio de output."""
    if precio is None:
        return None
    total = (tokens.get("input", 0) * precio["input"]
             + (tokens.get("output", 0) + tokens.get("reasoning", 0)) * precio["output"]
             + tokens.get("cache_read", 0) * precio["cache_read"]
             + tokens.get("cache_write", 0) * precio["cache_write"]) / 1_000_000
    return total * mult


# ------------------------------------------------------------------ OpenCode

# Se soportan dos generaciones de OpenCode (verificado con 1.18.x y 2.0.22):
#   1.x: `opencode export <id>`; mensajes {info: {role, agent, providerID, modelID, ...}, parts: [...]};
#        subagentes en partes tool "task" (state.metadata.sessionId).
#   2.x: `opencode session export <id>`; mensajes planos {type, agent, model: {id, providerID}, content: [...]};
#        subagentes en content tool "subagent" (state.metadata.sessionID) y en mensajes "synthetic" (metadata.childID).

def opencode(*args: str) -> str:
    # La salida va a un archivo temporal: OpenCode 2.x corta los JSON grandes cuando escribe a una tubería.
    try:
        with tempfile.TemporaryFile("w+", encoding="utf-8") as salida:
            r = subprocess.run(["opencode", *args], stdout=salida, stderr=subprocess.PIPE, text=True, cwd=RAIZ, timeout=120)
            salida.seek(0)
            texto = salida.read()
    except (OSError, subprocess.TimeoutExpired) as e:
        raise Fallo(f"no se pudo ejecutar 'opencode {args[0]}': {e}")
    if r.returncode != 0:
        raise Fallo(f"'opencode {' '.join(args[:2])}' falló: {r.stderr.strip()[:200]}")
    return texto


_VERSION_MAYOR: int | None = None


def version_mayor() -> int:
    """Versión mayor de OpenCode ("1.18.33" → 1, "opencode v2.0.22" → 2). Si no se reconoce, se asume la más nueva."""
    global _VERSION_MAYOR
    if _VERSION_MAYOR is None:
        try:
            m = re.search(r"(\d+)\.\d+", opencode("--version"))
        except Fallo:
            m = None
        _VERSION_MAYOR = int(m.group(1)) if m else 2
    return _VERSION_MAYOR


def exportar(session_id: str) -> dict:
    orden = ("session", "export") if version_mayor() >= 2 else ("export",)
    salida = opencode(*orden, session_id)
    inicio = salida.find("{")
    if inicio < 0:
        raise Fallo(f"'opencode {' '.join(orden)} {session_id}' no devolvió JSON")
    return json.loads(salida[inicio:])


def hijas(export: dict) -> list[str]:
    """Sesiones de subagentes lanzadas desde esta sesión."""
    ids = []

    def agregar(sid) -> None:
        if sid and sid not in ids:
            ids.append(sid)

    for m in export.get("messages", []):
        for p in m.get("parts", []):                         # 1.x
            if p.get("type") == "tool" and p.get("tool") == "task":
                agregar(((p.get("state") or {}).get("metadata") or {}).get("sessionId"))
        for c in m.get("content") or []:                     # 2.x
            if isinstance(c, dict) and c.get("type") == "tool" and c.get("name") in ("subagent", "task"):
                meta = (c.get("state") or {}).get("metadata") or {}
                agregar(meta.get("sessionID") or meta.get("sessionId"))
        meta = m.get("metadata") or {}
        if m.get("type") == "synthetic" and meta.get("source") == "subagent":
            agregar(meta.get("childID"))
    return ids


def respuestas(export: dict) -> list[dict]:
    """Mensajes del asistente con su agente, modelo, hora y tokens."""
    resultado = []
    for m in export.get("messages", []):
        info = m.get("info") or m                            # 1.x: dentro de "info"; 2.x: mensaje plano
        if (info.get("role") or info.get("type")) != "assistant":
            continue
        t = info.get("tokens") or {}
        cache = t.get("cache") or {}
        tokens = {"input": t.get("input", 0), "output": t.get("output", 0), "reasoning": t.get("reasoning", 0),
                  "cache_read": cache.get("read", 0), "cache_write": cache.get("write", 0)}
        if not any(tokens.values()):
            continue
        modelo = info.get("model") if isinstance(info.get("model"), dict) else {}
        resultado.append({
            "ms": (info.get("time") or {}).get("created", 0),
            "agente": info.get("agent") or info.get("mode") or "?",
            "modelo": f"{info.get('providerID') or modelo.get('providerID') or '?'}/{info.get('modelID') or modelo.get('id') or '?'}",
            "tokens": tokens,
            "costo_opencode": float(info.get("cost") or 0),
        })
    return resultado


def sesiones_del_proyecto() -> list[dict]:
    orden = ("session", "list", "--format", "json", "-n", "1000")
    datos = json.loads(opencode(*orden).strip() or "[]")
    if not datos and version_mayor() >= 2:
        # En 2.x el servicio en segundo plano a veces no devuelve las sesiones del proyecto; leyendo directo sí.
        datos = json.loads(opencode(*orden, "--standalone").strip() or "[]")
    raiz = str(RAIZ.resolve())
    return [s for s in datos if str(Path(s.get("directory", "")).resolve()).startswith(raiz)]


# ------------------------------------------------------------------ tareas

def carpeta_de_rama(rama: str) -> Path | None:
    if not rama or not SPECS.exists():
        return None
    if (SPECS / rama).is_dir():
        return SPECS / rama
    for p in SPECS.glob("**/estado.md"):
        e = leer_estado(p)
        if e["campos"].get("rama", "") == rama:
            return p.parent
    return None


def inicio_de_rama(rama: str) -> int:
    """Momento en que se creó la rama en esta máquina (reflog), o su primer commit propio."""
    reflog = git("reflog", "show", "--date=unix", "--format=%gd", rama)
    lineas = [l for l in reflog.splitlines() if "@{" in l]
    if lineas:
        try:
            return int(lineas[-1].split("@{")[1].rstrip("}")) * 1000
        except ValueError:
            pass
    base = git("merge-base", "HEAD", "main") or git("merge-base", "HEAD", "master")
    primero = git("log", "--reverse", "--format=%ct", f"{base}..HEAD").splitlines() if base else []
    return int(primero[0]) * 1000 if primero else ahora_ms()


def nuevo_costos(rama: str) -> dict:
    return {
        "rama": rama,
        "estado": "abierto",
        "inicio": iso(max(0, inicio_de_rama(rama) - MARGEN_INICIO_MS)),
        "cerrado": None,
        "moneda": "USD",
        "nota": "Costo equivalente a precio de API (models.dev). Con suscripción, el gasto real es el % de la consola.",
        "precios": [],
        "sesiones": [],
        "totales": {},
    }


def actualizar_precios(costos: dict, modelos: set[str]) -> tuple[list[str], list[str]]:
    """Agrega una versión de precios si cambió alguno (nunca reemplaza las anteriores)."""
    vigentes, avisos = precios_vigentes(modelos)
    actuales = {}
    for v in sorted(costos["precios"], key=lambda v: v["desde"]):
        actuales.update(v["modelos"])
    cambios = {}
    for m, p in vigentes.items():
        previo = actuales.get(m)
        if previo is None or any(abs(previo.get(k, 0) - p[k]) > 1e-12 for k in ("input", "output", "cache_read", "cache_write")):
            cambios[m] = p
    mensajes = []
    if cambios:
        costos["precios"].append({"desde": iso(ahora_ms()), "modelos": cambios})
        for m, p in cambios.items():
            antes = actuales.get(m)
            if antes:
                mensajes.append(f"Precio nuevo de {m}: entrada ${antes['input']} → ${p['input']}, salida ${antes['output']} → ${p['output']} (por millón)")
    sin_precio = sorted(m for m in modelos if m not in vigentes and m not in actuales)
    if sin_precio:
        avisos.append("Sin precio para: " + ", ".join(sin_precio) + ". Agrégalo en equipo/config.json → costos.precios_manuales")
    return mensajes, avisos


def recalcular_totales(costos: dict) -> None:
    por = {}
    for s in costos["sesiones"]:
        for c in s["consumo"]:
            clave = (c["agente"], c["modelo"])
            t = por.setdefault(clave, {"agente": c["agente"], "modelo": c["modelo"], "respuestas": 0,
                                       "tokens": {k: 0 for k in TIPOS}, "costo": 0.0, "costo_opencode": 0.0,
                                       "sin_precio": False})
            t["respuestas"] += c["respuestas"]
            for k in TIPOS:
                t["tokens"][k] += c["tokens"].get(k, 0)
            if c["costo"] is None:
                t["sin_precio"] = True
            else:
                t["costo"] += c["costo"]
            t["costo_opencode"] += c.get("costo_opencode", 0)
    filas = sorted(por.values(), key=lambda t: -t["costo"])
    for t in filas:
        t["costo"] = round(t["costo"], 6)
        t["costo_opencode"] = round(t["costo_opencode"], 6)
    costos["totales"] = {
        "por_agente_modelo": filas,
        "tokens": sum(sum(t["tokens"].values()) for t in filas),
        "costo": round(sum(t["costo"] for t in filas), 6),
        "costo_opencode": round(sum(t["costo_opencode"] for t in filas), 6),
    }


def registrar(costos: dict, ruta: Path, rama: str) -> list[str]:
    """Agrega a costos.json lo consumido desde el último registro. Devuelve avisos."""
    avisos = []
    if shutil.which("opencode") is None:
        return ["No se encontró 'opencode' en esta terminal: no se registró consumo (¿estás en WSL/Ubuntu?)."]
    registro = leer_json(REGISTRO, {})
    # Lo ya guardado en cualquier costos.json (esta rama, otras ramas, tareas cerradas) también cuenta como registrado,
    # por si se perdió el registro local de esta máquina.
    for _, otro in todos_los_costos() + [("actual", costos)]:
        for s in otro.get("sesiones", []):
            try:
                hasta = int(s["hasta_ms"]) if "hasta_ms" in s else de_iso(s["hasta"]) + 999
            except (KeyError, ValueError):
                continue
            previo = registro.get(s["id"]) or {}
            if hasta > previo.get("hasta", 0):
                registro[s["id"]] = {"rama": otro.get("rama", "?"), "hasta": hasta, "proyecto": str(RAIZ.resolve())}
    inicio = de_iso(costos["inicio"])
    usuario = git("config", "user.name") or os.environ.get("USER", "?")
    pico = config_costos()["pico"]

    try:
        raices = sesiones_del_proyecto()
    except (Fallo, ValueError) as e:
        return [f"No se pudo leer las sesiones de OpenCode: {e}"]
    if not raices:
        return ["OpenCode no devolvió ninguna sesión de este proyecto (compruébalo con: opencode session list). "
                "Si trabajaste con otra herramienta o en otra carpeta, ese consumo no se registra aquí."]
    if not any(s.get("updated", 0) >= inicio for s in raices):
        avisos.append(f"Hay {len(raices)} sesión(es) de OpenCode, pero todas anteriores al inicio de la tarea ({costos['inicio']}).")

    nuevas_resp: list[tuple[dict, dict, str | None]] = []   # (sesion_info, respuesta, padre)
    hasta_por_sesion: dict[str, int] = {}
    otras_ramas: set[str] = set()
    for s in raices:
        sid = s.get("id")
        previo = registro.get(sid) or {}
        if previo.get("hasta", 0) >= s.get("updated", 0) or s.get("updated", 0) < inicio:
            continue
        pendientes = [(sid, None)]
        vistos = set()
        while pendientes:
            actual, padre = pendientes.pop()
            if actual in vistos:
                continue
            vistos.add(actual)
            try:
                exp = exportar(actual)
            except (Fallo, ValueError) as e:
                avisos.append(f"Sesión {actual}: {e}")
                continue
            info = exp.get("info") or {}
            reg = registro.get(actual) or {}
            if reg.get("rama") and reg["rama"] != rama:
                otras_ramas.add(reg["rama"])
            desde = max(reg.get("hasta", 0), inicio)
            for r in respuestas(exp):
                if r["ms"] > desde:
                    nuevas_resp.append((info, r, padre))
                    hasta_por_sesion[actual] = max(hasta_por_sesion.get(actual, 0), r["ms"])
            for h in hijas(exp):
                pendientes.append((h, actual))

    if otras_ramas:
        avisos.append("Algunas sesiones ya tenían consumo registrado en otra tarea (" + ", ".join(sorted(otras_ramas))
                      + "); lo nuevo se asigna a esta rama.")
    if not nuevas_resp:
        return avisos

    modelos = {r["modelo"] for _, r, _ in nuevas_resp} | modelos_configurados()
    cambios, av = actualizar_precios(costos, modelos)
    avisos += cambios + av
    # Sin ningún precio (models.dev caído y sin copia local) no se registra: el consumo quedaría guardado sin costo
    # para siempre. Las sesiones siguen pendientes y se registran en la próxima ejecución.
    if all(precio_en(costos, r["modelo"], r["ms"]) is None for _, r, _ in nuevas_resp):
        avisos.append("NO se registró el consumo: no hay precio para ningún modelo usado. Vuelve a ejecutar 'make costos' "
                      "cuando models.dev responda (queda pendiente, no se pierde).")
        return avisos

    # Agrupa por sesión, luego por agente y modelo
    por_sesion: dict[str, dict] = {}
    for info, r, padre in nuevas_resp:
        sid = info.get("id", "?")
        s = por_sesion.setdefault(sid, {
            "id": sid, "padre": padre, "titulo": (info.get("title") or "")[:80], "herramienta": "opencode",
            "usuario": usuario, "registrado": iso(ahora_ms()), "desde": r["ms"], "hasta": r["ms"], "consumo": {},
        })
        s["desde"], s["hasta"] = min(s["desde"], r["ms"]), max(s["hasta"], r["ms"])
        precio = precio_en(costos, r["modelo"], r["ms"])
        mult = multiplicador_pico(r["modelo"], r["ms"], pico)
        c = s["consumo"].setdefault((r["agente"], r["modelo"]), {
            "agente": r["agente"], "modelo": r["modelo"], "respuestas": 0, "respuestas_pico": 0,
            "tokens": {k: 0 for k in TIPOS}, "costo": 0.0, "costo_opencode": 0.0})
        c["respuestas"] += 1
        c["respuestas_pico"] += 1 if mult != 1 else 0
        for k in TIPOS:
            c["tokens"][k] += r["tokens"][k]
        valor = costo(r["tokens"], precio, mult)
        c["costo"] = None if valor is None or c["costo"] is None else c["costo"] + valor
        c["costo_opencode"] += r["costo_opencode"]

    for s in por_sesion.values():
        s["hasta_ms"] = s["hasta"]
        s["desde"], s["hasta"] = iso(s["desde"]), iso(s["hasta"])
        s["consumo"] = [dict(c, costo=None if c["costo"] is None else round(c["costo"], 6),
                             costo_opencode=round(c["costo_opencode"], 6)) for c in s["consumo"].values()]
        costos["sesiones"].append(s)

    recalcular_totales(costos)
    escribir_json(ruta, costos)
    for sid, hasta in hasta_por_sesion.items():
        registro[sid] = {"rama": rama, "hasta": hasta, "proyecto": str(RAIZ.resolve())}
    escribir_json(REGISTRO, registro)
    total = sum(len(s["consumo"]) for s in por_sesion.values())
    avisos.insert(0, f"Registradas {len(por_sesion)} sesión(es) nuevas o continuadas ({total} filas de consumo) en {ruta.relative_to(RAIZ)}.")
    return avisos


# ------------------------------------------------------------------ informes

def tabla(filas: list[dict], titulo_total: str = "Total") -> None:
    print(f"  {'Agente':<18} {'Modelo':<34} {'Tokens':>8} {'Entrada':>8} {'Salida':>8} {'Caché':>8} {'Equivalente':>12}")
    total_t, total_c, sin_precio = 0, 0.0, False
    for t in filas:
        tk = t["tokens"]
        suma = sum(tk.values())
        total_t += suma
        valor = None if t.get("sin_precio") else t["costo"]
        if valor is None:
            sin_precio = True
        else:
            total_c += valor
        modelo = t["modelo"].split("/", 1)[-1]
        print(f"  {t['agente'][:18]:<18} {modelo[:34]:<34} {fmt_tokens(suma):>8} {fmt_tokens(tk['input']):>8} "
              f"{fmt_tokens(tk['output'] + tk['reasoning']):>8} {fmt_tokens(tk['cache_read'] + tk['cache_write']):>8} {fmt_usd(valor):>12}")
    print(f"  {titulo_total:<53} {fmt_tokens(total_t):>8} {'':>26} {fmt_usd(total_c):>12}" + ("  (+ modelos sin precio)" if sin_precio else ""))


def informe_tarea(costos: dict, ruta: str) -> None:
    t = costos.get("totales") or {}
    estado = costos.get("estado", "abierto")
    extra = f" el {costos['cerrado'][:10]}" if costos.get("cerrado") else ""
    print(f"COSTO DE LA TAREA · {costos.get('rama', '?')} · {estado}{extra}")
    print(f"  {ruta} · {len(costos.get('sesiones', []))} registro(s) de sesión · {len(costos.get('precios', []))} versión(es) de precios")
    filas = t.get("por_agente_modelo") or []
    if not filas:
        print("  Aún no hay consumo registrado.")
        return
    tabla(filas)
    pico = sum(c.get("respuestas_pico", 0) for s in costos["sesiones"] for c in s["consumo"])
    if pico:
        print(f"  Incluye {pico} respuesta(s) en hora pico con tarifa aumentada.")
    if t.get("costo_opencode") is not None:
        print(f"  Según OpenCode (precio del momento de cada respuesta): {fmt_usd(t['costo_opencode'])}")
    print("  Equivalente a precio de API. Con OpenCode Go el gasto real es el % de tu consola.")


def informe_hoy(costos: dict) -> None:
    modelos = {c["modelo"] for s in costos["sesiones"] for c in s["consumo"]}
    vigentes, avisos = precios_vigentes(modelos)
    filas = []
    for t in (costos.get("totales") or {}).get("por_agente_modelo", []):
        valor = costo(t["tokens"], vigentes.get(t["modelo"]))
        filas.append(dict(t, costo=valor or 0.0, sin_precio=valor is None))
    print()
    print("A PRECIOS DE HOY (solo para cotizar; no modifica costos.json, sin tarifa de hora pico)")
    tabla(filas)
    for a in avisos:
        print(f"  ⚠ {a}")


def todos_los_costos() -> list[tuple[str, dict]]:
    """costos.json de esta copia y de las otras ramas (sin merge todavía)."""
    vistos, resultado = set(), []
    for p in sorted(SPECS.glob("**/costos.json")) if SPECS.exists() else []:
        c = leer_json(p, None)
        if c:
            resultado.append((str(p.relative_to(RAIZ)), c))
            vistos.add(c.get("rama"))
    refs = git("for-each-ref", "--format=%(refname:short)", "refs/heads", "refs/remotes")
    for ref in refs.splitlines():
        nombre = ref.split("/", 1)[1] if ref.startswith(("origin/", "upstream/")) else ref
        if nombre in vistos or nombre in ("HEAD", "main", "master") or "/" in nombre:
            continue
        texto = git("show", f"{ref}:specs/{nombre}/costos.json")
        if texto:
            try:
                resultado.append((f"{ref}:specs/{nombre}/costos.json", json.loads(texto)))
                vistos.add(nombre)
            except ValueError:
                pass
    return resultado


def informe_proyecto() -> None:
    todos = todos_los_costos()
    print("COSTO DEL PROYECTO POR FUNCIONALIDAD")
    if not todos:
        print("  Todavía no hay costos registrados (specs/<rama>/costos.json).")
        return
    print(f"  {'Funcionalidad':<34} {'Estado':<9} {'Tokens':>8} {'Equivalente':>12}")
    total_t, total_c = 0, 0.0
    por: dict[tuple, dict] = {}
    for _, c in todos:
        t = c.get("totales") or {}
        total_t += t.get("tokens", 0)
        total_c += t.get("costo", 0)
        print(f"  {c.get('rama', '?')[:34]:<34} {c.get('estado', '?'):<9} {fmt_tokens(t.get('tokens', 0)):>8} {fmt_usd(t.get('costo', 0)):>12}")
        for f in t.get("por_agente_modelo", []):
            a = por.setdefault((f["agente"], f["modelo"]), {"agente": f["agente"], "modelo": f["modelo"],
                                                             "tokens": {k: 0 for k in TIPOS}, "costo": 0.0})
            for k in TIPOS:
                a["tokens"][k] += f["tokens"].get(k, 0)
            a["costo"] += f["costo"]
    print(f"  {'Total':<44} {fmt_tokens(total_t):>8} {fmt_usd(total_c):>12}")
    print()
    print("POR AGENTE Y MODELO (todas las funcionalidades)")
    tabla(sorted(por.values(), key=lambda a: -a["costo"]))


# ------------------------------------------------------------------ main

def main() -> int:
    args = set(sys.argv[1:])
    if "--todo" in args:
        informe_proyecto()
        return 0

    rama = git("branch", "--show-current")
    carpeta = carpeta_de_rama(rama)
    if carpeta is None:
        if rama in ("main", "master", ""):
            print("Estás en la rama principal: el consumo se registra en la rama de cada funcionalidad.\n")
        else:
            print(f"No hay specs/{rama}/ (ni un estado.md con 'Rama | {rama}'): no se registra consumo.\n")
        informe_proyecto()
        return 0

    ruta = carpeta / "costos.json"
    costos = leer_json(ruta, None) if ruta.exists() else nuevo_costos(rama)
    if costos is None:
        print(f"✗ {ruta.relative_to(RAIZ)} no es un JSON válido.", file=sys.stderr)
        return 1

    avisos: list[str] = []
    if costos.get("estado") == "cerrado":
        avisos.append("El costo de esta tarea está cerrado: no se registra más consumo aquí.")
    elif "--sin-registrar" not in args and "--hoy" not in args:
        avisos += registrar(costos, ruta, rama)
        if "--cerrar" in args:
            costos["estado"] = "cerrado"
            costos["cerrado"] = iso(ahora_ms())
            recalcular_totales(costos)
            escribir_json(ruta, costos)
            avisos.append(f"Costo de la tarea CERRADO. Haz commit de {ruta.relative_to(RAIZ)}: desde ahora no se puede modificar.")

    informe_tarea(costos, str(ruta.relative_to(RAIZ)))
    if "--hoy" in args:
        informe_hoy(costos)
    if avisos:
        print()
        for a in avisos:
            print(f"  {'✓' if a.startswith(('Registradas', 'Costo de la tarea CERRADO')) else '⚠'} {a}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
