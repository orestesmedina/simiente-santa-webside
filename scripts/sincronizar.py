#!/usr/bin/env python3
"""Genera la configuración específica de cada agente de código a partir de la fuente única.

Fuente única (lo que el equipo edita):
  AGENTS.md                 instrucciones del proyecto
  equipo/agentes/*.md       definición neutral de cada subagente
  .agents/skills/           skills compartidas (estándar SKILL.md)
  equipo/config.json        herramientas activas y modelos (por nivel, por agente y del orquestador)
  equipo/adaptadores/       archivos propios de cada herramienta (ej. permisos y hooks de Claude Code)

Salida (generada, no editar a mano):
  Claude Code  -> CLAUDE.md, .claude/agents/, .claude/skills/, .claude/settings.json
  Codex        -> .codex/agents/*.toml         (lee AGENTS.md y .agents/skills/ directamente)
  OpenCode     -> .opencode/agents/*.md, opencode.json (si hay modelo de orquestador o adaptador)

Uso:
  python3 scripts/sincronizar.py              genera los archivos
  python3 scripts/sincronizar.py --verificar  falla si lo generado no está al día (para CI)
  python3 scripts/sincronizar.py --modelos    muestra qué modelo usa cada agente en cada herramienta

Solo usa la biblioteca estándar de Python 3.9+.
"""
from __future__ import annotations

import json
import shutil
import subprocess
import sys
from pathlib import Path

RAIZ = Path(__file__).resolve().parent.parent
AVISO = "GENERADO por scripts/sincronizar.py desde {fuente}. No editar: cambia la fuente y ejecuta `make sincronizar`."

# Carpetas y archivos que este script controla por completo, por herramienta.
GESTIONADOS = {
    "claude": {"dirs": [".claude/agents", ".claude/skills"], "files": ["CLAUDE.md", ".claude/settings.json"]},
    "codex": {"dirs": [".codex/agents"], "files": []},
    "opencode": {"dirs": [".opencode/agents"], "files": ["opencode.json"]},
}
PREFIJOS = {"claude": (".claude", "CLAUDE"), "codex": (".codex",), "opencode": (".opencode", "opencode.json")}

ACCESOS = {"lectura", "documentos", "completo"}
NIVELES = {"alto", "medio", "bajo"}


# ---------------------------------------------------------------- lectura de la fuente

def leer_agente(ruta: Path) -> dict:
    texto = ruta.read_text(encoding="utf-8")
    ruta_txt = ruta.relative_to(RAIZ).as_posix()
    if not texto.startswith("---\n"):
        raise ValueError(f"{ruta_txt}: falta el bloque de metadatos (---)")
    _, cabecera, cuerpo = texto.split("---\n", 2)
    datos = {}
    for linea in cabecera.strip().splitlines():
        clave, _, valor = linea.partition(":")
        datos[clave.strip()] = valor.strip()

    for campo in ("nombre", "descripcion", "acceso", "nivel"):
        if not datos.get(campo):
            raise ValueError(f"{ruta_txt}: falta el campo '{campo}'")
    if datos["acceso"] not in ACCESOS:
        raise ValueError(f"{ruta_txt}: acceso debe ser uno de {sorted(ACCESOS)}")
    if datos["nivel"] not in NIVELES:
        raise ValueError(f"{ruta_txt}: nivel debe ser uno de {sorted(NIVELES)}")
    if datos["nombre"] != ruta.stem:
        raise ValueError(f"{ruta_txt}: 'nombre' debe coincidir con el nombre del archivo")

    texto_temp = datos.get("temperatura", "")
    datos["temperatura"] = None
    if texto_temp:
        try:
            datos["temperatura"] = float(texto_temp)
        except ValueError:
            raise ValueError(f"{ruta_txt}: temperatura debe ser un número (ej. 0.1)")
        if not 0 <= datos["temperatura"] <= 2:
            raise ValueError(f"{ruta_txt}: temperatura debe estar entre 0 y 2")

    datos["web"] = datos.get("web", "no") == "si"
    datos["skills"] = [s.strip() for s in datos.get("skills", "").split(",") if s.strip()]
    datos["fuente"] = ruta.relative_to(RAIZ).as_posix()

    instrucciones = cuerpo.strip()
    if datos["skills"]:
        lista = ", ".join(f"`{s}`" for s in datos["skills"])
        instrucciones += f"\n\n## Skills que debes aplicar\n{lista} (en `.agents/skills/`)."
    datos["instrucciones"] = instrucciones
    return datos


def config_modelos(config: dict, herramienta: str) -> dict:
    """Normaliza la sección de modelos de una herramienta.

    Formato:
      "opencode": {
        "orquestador": "opencode-go/glm-5.3",       # modelo de la sesión principal (opcional)
        "ligero": "opencode-go/glm-5.3-flash",      # tareas auxiliares, solo OpenCode (opcional)
        "niveles": {"alto": "...", "medio": "...", "bajo": "..."},
        "agentes": {"revisor-codigo": "..."}        # excepciones por agente (opcional)
      }
    También acepta el formato antiguo con alto/medio/bajo directamente.
    """
    crudo = config.get("modelos", {}).get(herramienta, {}) or {}
    niveles = crudo.get("niveles") or {k: crudo.get(k, "") for k in NIVELES}
    return {
        "orquestador": crudo.get("orquestador", ""),
        "ligero": crudo.get("ligero", ""),
        "niveles": niveles,
        "agentes": crudo.get("agentes", {}) or {},
    }


def modelo_de(agente: dict, modelos: dict) -> tuple[str, str]:
    """Devuelve (modelo, origen). Prioridad: excepción por agente > nivel > heredado de la sesión."""
    if modelos["agentes"].get(agente["nombre"]):
        return modelos["agentes"][agente["nombre"]], "agente"
    if modelos["niveles"].get(agente["nivel"]):
        return modelos["niveles"][agente["nivel"]], f"nivel {agente['nivel']}"
    return "", "sesión"


# Herramientas cuyos subagentes aceptan temperatura.
SOPORTA_TEMPERATURA = {"opencode"}


def config_temperatura(config: dict, herramienta: str) -> dict:
    """Sección opcional modelos.<herramienta>.temperatura:
      "temperatura": {
        "agentes": {"revisor-codigo": 0.0},          # excepciones sobre el valor por defecto del agente
        "sin_temperatura": ["opencode-go/kimi-k3"]    # modelos con temperatura fija: no se envía
      }
    """
    crudo = (config.get("modelos", {}).get(herramienta, {}) or {}).get("temperatura", {}) or {}
    return {"agentes": crudo.get("agentes", {}) or {}, "sin_temperatura": set(crudo.get("sin_temperatura", []) or [])}


def temperatura_de(agente: dict, modelo: str, temp_cfg: dict) -> tuple[float | None, str]:
    """Devuelve (temperatura, origen). Prioridad: modelo sin temperatura > excepción en config > valor del agente > modelo."""
    if modelo and modelo in temp_cfg["sin_temperatura"]:
        return None, "fija del modelo"
    if agente["nombre"] in temp_cfg["agentes"]:
        return float(temp_cfg["agentes"][agente["nombre"]]), "config"
    if agente["temperatura"] is not None:
        return agente["temperatura"], "agente"
    return None, "del modelo"


def validar_config(config: dict, agentes: list[dict]) -> list[str]:
    nombres = {a["nombre"] for a in agentes}
    errores = []
    for h in config.get("modelos", {}):
        for nombre in config_modelos(config, h)["agentes"]:
            if nombre not in nombres:
                errores.append(f"equipo/config.json: modelos.{h}.agentes.{nombre} no corresponde a ningún agente de equipo/agentes/")
        for nombre, valor in config_temperatura(config, h)["agentes"].items():
            if nombre not in nombres:
                errores.append(f"equipo/config.json: modelos.{h}.temperatura.agentes.{nombre} no corresponde a ningún agente")
            elif not isinstance(valor, (int, float)) or not 0 <= valor <= 2:
                errores.append(f"equipo/config.json: modelos.{h}.temperatura.agentes.{nombre} debe ser un número entre 0 y 2")
    return errores


def yaml_texto(valor: str) -> str:
    """Cadena YAML segura (JSON es YAML válido)."""
    return json.dumps(valor, ensure_ascii=False)


def toml_texto_multilinea(valor: str) -> str:
    if "'''" not in valor:
        return "'''\n" + valor + "\n'''"
    escapado = valor.replace("\\", "\\\\").replace('"""', '\\"\\"\\"')
    return '"""\n' + escapado + '\n"""'


def json_bytes(datos: dict) -> bytes:
    return (json.dumps(datos, ensure_ascii=False, indent=2) + "\n").encode()


# ---------------------------------------------------------------- generadores

def generar_claude(agentes: list[dict], config: dict) -> dict[str, bytes]:
    salida: dict[str, bytes] = {}
    modelos = config_modelos(config, "claude")

    salida["CLAUDE.md"] = (
        f"<!-- {AVISO.format(fuente='AGENTS.md')} -->\n"
        "Las instrucciones de este proyecto están en AGENTS.md (compartido con otras herramientas):\n\n"
        "@AGENTS.md\n"
    ).encode()

    herramientas_por_acceso = {
        "lectura": ["Read", "Grep", "Glob", "Bash"],
        "documentos": ["Read", "Write", "Edit", "Grep", "Glob"],
        "completo": ["Read", "Write", "Edit", "Grep", "Glob", "Bash"],
    }
    for a in agentes:
        tools = list(herramientas_por_acceso[a["acceso"]])
        if a["web"]:
            tools += ["WebSearch", "WebFetch"]
        cab = ["---", f"name: {a['nombre']}", f"description: {yaml_texto(a['descripcion'])}", f"tools: {', '.join(tools)}"]
        modelo, _ = modelo_de(a, modelos)
        if modelo:
            cab.append(f"model: {modelo}")
        if a["skills"]:
            cab.append(f"skills: {', '.join(a['skills'])}")
        cab.append("---")
        texto = "\n".join(cab) + f"\n<!-- {AVISO.format(fuente=a['fuente'])} -->\n\n{a['instrucciones']}\n"
        salida[f".claude/agents/{a['nombre']}.md"] = texto.encode()

    # Skills: copia de .agents/skills para Claude Code.
    origen = RAIZ / ".agents/skills"
    for archivo in sorted(origen.rglob("*")):
        if archivo.is_file():
            rel = archivo.relative_to(origen).as_posix()
            salida[f".claude/skills/{rel}"] = archivo.read_bytes()

    ajustes_ruta = RAIZ / "equipo/adaptadores/claude/settings.json"
    ajustes = json.loads(ajustes_ruta.read_text(encoding="utf-8")) if ajustes_ruta.exists() else {}
    if modelos["orquestador"]:
        ajustes["model"] = modelos["orquestador"]
    if ajustes:
        salida[".claude/settings.json"] = json_bytes(ajustes)
    return salida


def generar_codex(agentes: list[dict], config: dict) -> dict[str, bytes]:
    salida: dict[str, bytes] = {}
    modelos = config_modelos(config, "codex")
    esfuerzo = config.get("esfuerzo_codex", {})
    for a in agentes:
        lineas = [
            f"# {AVISO.format(fuente=a['fuente'])}",
            f"name = {json.dumps(a['nombre'])}",
            f"description = {json.dumps(a['descripcion'], ensure_ascii=False)}",
        ]
        modelo, _ = modelo_de(a, modelos)
        if modelo:
            lineas.append(f"model = {json.dumps(modelo)}")
        if esfuerzo.get(a["nivel"]):
            lineas.append(f"model_reasoning_effort = {json.dumps(esfuerzo[a['nivel']])}")
        sandbox = "read-only" if a["acceso"] == "lectura" else "workspace-write"
        lineas.append(f"sandbox_mode = {json.dumps(sandbox)}")
        lineas.append("")
        lineas.append(f"developer_instructions = {toml_texto_multilinea(a['instrucciones'])}")
        salida[f".codex/agents/{a['nombre']}.toml"] = ("\n".join(lineas) + "\n").encode()
    return salida


def generar_opencode(agentes: list[dict], config: dict) -> dict[str, bytes]:
    salida: dict[str, bytes] = {}
    modelos = config_modelos(config, "opencode")
    temp_cfg = config_temperatura(config, "opencode")
    for a in agentes:
        edit = "deny" if a["acceso"] == "lectura" else "allow"
        bash = "deny" if a["acceso"] == "documentos" else "allow"
        web = "allow" if a["web"] else "deny"
        cab = ["---", f"description: {yaml_texto(a['descripcion'])}", "mode: subagent"]
        modelo, _ = modelo_de(a, modelos)
        if modelo:
            cab.append(f"model: {modelo}")
        temperatura, _ = temperatura_de(a, modelo, temp_cfg)
        if temperatura is not None:
            cab.append(f"temperature: {temperatura:g}")
        cab += ["permission:", f"  edit: {edit}", f"  bash: {bash}", f"  webfetch: {web}", "---"]
        texto = "\n".join(cab) + f"\n<!-- {AVISO.format(fuente=a['fuente'])} -->\n\n{a['instrucciones']}\n"
        salida[f".opencode/agents/{a['nombre']}.md"] = texto.encode()

    # opencode.json: modelo del orquestador y modelo ligero, sobre la base del adaptador si existe.
    base_ruta = RAIZ / "equipo/adaptadores/opencode/opencode.json"
    base = json.loads(base_ruta.read_text(encoding="utf-8")) if base_ruta.exists() else {}
    if modelos["orquestador"] or modelos["ligero"] or base:
        datos = {"$schema": "https://opencode.ai/config.json", **base}
        if modelos["orquestador"]:
            datos["model"] = modelos["orquestador"]
        if modelos["ligero"]:
            datos["small_model"] = modelos["ligero"]
        salida["opencode.json"] = json_bytes(datos)
    return salida


GENERADORES = {"claude": generar_claude, "codex": generar_codex, "opencode": generar_opencode}


# ---------------------------------------------------------------- informe de modelos

def modelos_disponibles_opencode() -> set[str] | None:
    """Lista de modelos que reconoce la instalación local de OpenCode, o None si no está instalado."""
    if not shutil.which("opencode"):
        return None
    try:
        r = subprocess.run(["opencode", "models"], capture_output=True, text=True, timeout=30)
    except (OSError, subprocess.TimeoutExpired):
        return None
    return {l.strip() for l in r.stdout.splitlines() if "/" in l} if r.returncode == 0 else None


def informe_modelos(agentes: list[dict], config: dict, activas: list[str]) -> int:
    disponibles = modelos_disponibles_opencode() if "opencode" in activas else None
    desconocidos = 0
    for h in activas:
        m = config_modelos(config, h)
        t = config_temperatura(config, h)
        soporta = h in SOPORTA_TEMPERATURA
        print(f"\n{h}" + ("" if soporta else "   (temperatura no configurable en esta herramienta)"))
        print(f"  {'orquestador':<20} {m['orquestador'] or '(modelo de la sesión)'}")
        if h == "opencode" and m["ligero"]:
            print(f"  {'(tareas ligeras)':<20} {m['ligero']}")
        for a in agentes:
            modelo, origen = modelo_de(a, m)
            marca = ""
            if h == "opencode" and disponibles is not None and modelo and modelo not in disponibles:
                marca = "   ⚠ no aparece en `opencode models`"
                desconocidos += 1
            temp_txt = ""
            if soporta:
                temperatura, t_origen = temperatura_de(a, modelo, t)
                temp_txt = f"  temp {temperatura:g} [{t_origen}]" if temperatura is not None else f"  temp — [{t_origen}]"
            print(f"  {a['nombre']:<20} {modelo or '(modelo de la sesión)':<34} [{origen}]{temp_txt}{marca}")
    if "opencode" in activas and disponibles is None:
        print("\n(OpenCode no está instalado aquí: no se verificó que los modelos existan.)")
    return 1 if desconocidos else 0


# ---------------------------------------------------------------- escritura / verificación

def archivos_actuales() -> dict[str, bytes]:
    actuales: dict[str, bytes] = {}
    for g in GESTIONADOS.values():
        for d in g["dirs"]:
            base = RAIZ / d
            if base.exists():
                for f in base.rglob("*"):
                    if f.is_file():
                        actuales[f.relative_to(RAIZ).as_posix()] = f.read_bytes()
        for f in g["files"]:
            p = RAIZ / f
            if p.exists():
                actuales[f] = p.read_bytes()
    return actuales


def main() -> int:
    args = sys.argv[1:]
    config = json.loads((RAIZ / "equipo/config.json").read_text(encoding="utf-8"))
    activas = config.get("herramientas", [])
    desconocidas = set(activas) - set(GENERADORES)
    if desconocidas:
        print(f"Herramientas desconocidas en equipo/config.json: {sorted(desconocidas)}", file=sys.stderr)
        return 1

    try:
        agentes = [leer_agente(p) for p in sorted((RAIZ / "equipo/agentes").glob("*.md"))]
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1
    errores = validar_config(config, agentes)
    if errores:
        for e in errores:
            print(f"Error: {e}", file=sys.stderr)
        return 1

    if "--modelos" in args:
        return informe_modelos(agentes, config, activas)

    # Un opencode.json escrito a mano (MCP, proveedores…) se conserva moviéndolo al adaptador.
    if "--verificar" not in args:
        existente = RAIZ / "opencode.json"
        adaptador = RAIZ / "equipo/adaptadores/opencode/opencode.json"
        if existente.exists() and not adaptador.exists():
            datos = json.loads(existente.read_text(encoding="utf-8"))
            datos.pop("model", None)
            datos.pop("small_model", None)
            adaptador.parent.mkdir(parents=True, exist_ok=True)
            adaptador.write_bytes(json_bytes(datos))
            print("Aviso: opencode.json existente movido a equipo/adaptadores/opencode/opencode.json "
                  "(edítalo ahí; los modelos se toman de equipo/config.json).")

    esperado: dict[str, bytes] = {}
    for h in activas:
        esperado.update(GENERADORES[h](agentes, config))

    if "--verificar" in args:
        actuales = archivos_actuales()
        faltan = sorted(set(esperado) - set(actuales))
        sobran = sorted(set(actuales) - set(esperado))
        distintos = sorted(k for k in set(esperado) & set(actuales) if esperado[k] != actuales[k])
        if faltan or sobran or distintos:
            for titulo, lista in (("Faltan", faltan), ("Sobran", sobran), ("Desactualizados", distintos)):
                for k in lista:
                    print(f"  {titulo}: {k}", file=sys.stderr)
            print("Los archivos generados no están al día. Ejecuta `make sincronizar`.", file=sys.stderr)
            return 1
        print(f"OK: {len(esperado)} archivos generados al día ({', '.join(activas)}).")
        return 0

    # Limpia todo lo gestionado (también de herramientas desactivadas) y vuelve a escribir.
    for g in GESTIONADOS.values():
        for d in g["dirs"]:
            shutil.rmtree(RAIZ / d, ignore_errors=True)
        for f in g["files"]:
            (RAIZ / f).unlink(missing_ok=True)
    for rel, contenido in esperado.items():
        destino = RAIZ / rel
        destino.parent.mkdir(parents=True, exist_ok=True)
        destino.write_bytes(contenido)

    print(f"Generados {len(esperado)} archivos para: {', '.join(activas)}.")
    for h in activas:
        n = sum(1 for k in esperado if k.startswith(PREFIJOS[h]))
        print(f"  {h}: {n} archivos")
    print("Revisa qué modelo usa cada agente con: make modelos")
    return 0


if __name__ == "__main__":
    sys.exit(main())
