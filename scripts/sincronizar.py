#!/usr/bin/env python3
"""Genera la configuración específica de cada agente de código a partir de la fuente única.

Fuente única (lo que el equipo edita):
  AGENTS.md                 instrucciones del proyecto
  equipo/agentes/*.md       definición neutral de cada subagente
  equipo/perfil.json        perfil del proyecto: de ahí salen la sección «Este proyecto» de cada rol y sus skills
  equipo/comandos/*.md      comandos que la persona escribe dentro de la herramienta (/bowser-status)
  .agents/skills/           skills compartidas (estándar SKILL.md); las de tecnología llegan del catálogo del kit
  equipo/config.json        herramientas activas y modelos (por nivel, por agente y del orquestador)
  equipo/adaptadores/       archivos propios de cada herramienta (ej. permisos y hooks de Claude Code)
  equipo/principios.md      la parte de la constitución que es del kit (igual en todos los proyectos)
  equipo/reglas.md          la parte de la constitución que es del proyecto

Salida (generada, no editar a mano):
  Claude Code  -> CLAUDE.md, .claude/agents/, .claude/skills/ (skills y comandos), .claude/settings.json
  Codex        -> .codex/agents/*.toml, .agents/skills/bowser-*/ (comandos; lee AGENTS.md y .agents/skills/ directamente)
  OpenCode     -> .opencode/agents/*.md, .opencode/commands/*.md, opencode.json (si hay modelo de orquestador o adaptador)
  Spec Kit     -> .specify/memory/constitution.md (los principios y, debajo, las reglas del proyecto)

Uso:
  python3 scripts/sincronizar.py              genera los archivos
  python3 scripts/sincronizar.py --verificar  falla si lo generado no está al día (para CI)
  python3 scripts/sincronizar.py --modelos    muestra qué modelo usa cada agente en cada herramienta

Solo usa la biblioteca estándar de Python 3.9+.
"""
from __future__ import annotations

import json
import re
import shutil
import subprocess
import sys
from pathlib import Path

import perfil as perfil_proyecto

RAIZ = Path(__file__).resolve().parent.parent
AVISO = "GENERADO por scripts/sincronizar.py desde {fuente}. No editar: cambia la fuente y ejecuta `make sync`."

# Prefijo de los comandos dentro de la herramienta (/bowser-status). Si el kit cambia de nombre, se cambia solo aquí.
PREFIJO_COMANDOS = "bowser-"
# Codex solo lee skills de .agents/skills/, que es una carpeta de fuentes: sus comandos se generan ahí,
# en carpetas con el prefijo. scripts/instalar_kit.py no las copia a los proyectos (se generan en cada uno).
COMANDOS_CODEX = f".agents/skills/{PREFIJO_COMANDOS}"

# Carpetas y archivos que este script controla por completo, por herramienta.
# "dirs" admite comodines (*) en el último tramo.
GESTIONADOS = {
    "claude": {"dirs": [".claude/agents", ".claude/skills"], "files": ["CLAUDE.md", ".claude/settings.json"]},
    "codex": {"dirs": [".codex/agents", COMANDOS_CODEX + "*"], "files": []},
    "opencode": {"dirs": [".opencode/agents", ".opencode/commands"], "files": ["opencode.json"]},
}
# Spec Kit deja sus comandos dentro de carpetas que este script controla (.claude/skills/speckit-*/ y
# .opencode/commands/speckit.*.md; verificado con specify 1.0.12, 2026-10-09). Lo que empieza así no es del kit:
# no se borra, no se copia de una herramienta a otra y no cuenta al verificar.
DE_SPEC_KIT = "speckit"
PREFIJOS = {"claude": (".claude", "CLAUDE"), "codex": (".codex", COMANDOS_CODEX), "opencode": (".opencode", "opencode.json")}

# La constitución que leen Spec Kit y los agentes se genera con dos fuentes: los principios, que son del kit, y las
# reglas, que son del proyecto. No depende de la herramienta: se genera siempre.
PRINCIPIOS = "equipo/principios.md"
REGLAS = "equipo/reglas.md"
CONSTITUCION = ".specify/memory/constitution.md"

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
    # Skills fijas del rol, para los roles propios de un proyecto. Las de tecnología no van aquí: las asigna el
    # perfil a cada parte (aplicar_perfil). Solo cuentan las que el proyecto tiene.
    datos["skills"] = [s.strip() for s in datos.get("skills", "").split(",")
                       if s.strip() and (RAIZ / ".agents/skills" / s.strip() / "SKILL.md").exists()]
    datos["fuente"] = ruta.relative_to(RAIZ).as_posix()
    datos["proyecto"] = datos.get("proyecto", "")
    if datos["proyecto"] not in VISTAS_DEL_PROYECTO:
        raise ValueError(f"{ruta_txt}: proyecto debe ser 'partes' (trabaja las partes que el perfil le asigna), "
                         "'mapa' (ve todas las partes) o no estar")

    instrucciones = cuerpo.strip()
    if datos["skills"]:
        lista = ", ".join(f"`{s}`" for s in datos["skills"])
        instrucciones += f"\n\n## Skills que debes aplicar\n{lista} (en `.agents/skills/`)."
    datos["instrucciones"] = instrucciones
    return datos


# ---------------------------------------------------------------- el proyecto, según su perfil

VISTAS_DEL_PROYECTO = {"", "partes", "mapa"}


def leer_perfil() -> tuple[dict | None, list[str]]:
    """(perfil, errores). Sin perfil o con errores, el perfil es None."""
    try:
        perfil = perfil_proyecto.cargar(RAIZ)
    except perfil_proyecto.ErrorPerfil as e:
        return None, [str(e)]
    if perfil is None:
        return None, []
    errores, _ = perfil_proyecto.validar(perfil, RAIZ)
    return (None, [f"{perfil_proyecto.RUTA}: {e}" for e in errores]) if errores else (perfil, [])


def instaladas(skills: list[str]) -> list[str]:
    """De esas skills, las que el proyecto tiene: el perfil puede nombrar una que todavía no existe.
    Primero el estándar y después las convenciones del proyecto."""
    hay = [s for s in dict.fromkeys(skills) if (RAIZ / ".agents/skills" / s / "SKILL.md").exists()]
    return sorted(hay, key=lambda s: s.startswith(perfil_proyecto.CONVENCIONES))


def en_codigo(valores: list[str]) -> str:
    return ", ".join(f"`{v}`" for v in valores)


def dentro_de(parte: dict, rutas: list[str]) -> list[str]:
    """Rutas de la parte, vistas desde la raíz del proyecto."""
    carpeta = parte["carpeta"].replace("\\", "/").strip("/")
    return [r if carpeta in ("", ".") else f"{carpeta}/{r}" for r in rutas]


def aplicar_perfil(agente: dict, perfil: dict | None, con_errores: bool) -> None:
    """Agrega a las instrucciones del rol la sección «Este proyecto» y le da las skills de sus partes.

    Un rol con `proyecto: partes` ve las partes que el perfil le asigna; uno con `proyecto: mapa`, todas.
    Así el rol no nombra carpetas, tecnologías ni comandos: salen del perfil al generar.
    """
    asignado = any(agente["nombre"] == r["rol"] for p in (perfil or {}).get("partes", []) for r in perfil_proyecto.roles_de(p))
    if not agente["proyecto"] and not asignado:
        return
    lineas = ["## Este proyecto"]
    if perfil is None:
        if con_errores:
            lineas.append(f"El perfil del proyecto (`{perfil_proyecto.RUTA}`) tiene errores y no se pudo leer. "
                          "No supongas carpetas, tecnologías ni comandos: avisa al orquestador antes de trabajar.")
        else:
            lineas.append(f"El proyecto todavía no tiene perfil (`{perfil_proyecto.RUTA}`): no se sabe qué partes "
                          "tiene, con qué tecnología ni cómo se prueba. No supongas carpetas, tecnologías ni comandos: "
                          "avisa al orquestador antes de trabajar. El perfil se crea con el comando `bowser-profile`.")
        agente["instrucciones"] += "\n\n" + "\n".join(lineas)
        return

    origen = f"Sale del perfil del proyecto (`{perfil_proyecto.RUTA}`), que no editas tú."
    propias: list[str] = []
    ajenas: list[str] = []
    detalle: list[str] = []
    for parte in perfil["partes"]:
        roles = perfil_proyecto.roles_de(parte)
        mio = next((r for r in roles if r["rol"] == agente["nombre"]), None)
        otros = [r["rol"] for r in roles if r["rol"] != agente["nombre"]]
        if mio is None and agente["proyecto"] != "mapa":
            quien = f", de {en_codigo(otros)}" if otros else ""
            ajenas.append(f"`{parte['nombre']}` (`{parte['carpeta']}`{quien})")
            continue

        titulo = f"- **{parte['nombre']}**, en `{parte['carpeta']}`"
        if mio is not None and agente["proyecto"] == "mapa":
            titulo += " (la trabajas tú)"
        detalle.append(titulo + (f": {parte['descripcion']}" if parte.get("descripcion") else "."))
        if mio is not None:
            skills = instaladas(mio["skills"])
            propias += skills
            if skills:
                detalle.append(f"  - Skills que debes aplicar: {en_codigo(skills)} (en `.agents/skills/`).")
            if otros:
                detalle.append(f"  - También la trabajan: {en_codigo(otros)}.")
        else:
            if roles:
                detalle.append(f"  - La trabajan: {en_codigo([r['rol'] for r in roles])}.")
            skills = instaladas([s for r in roles for s in r["skills"]] + parte.get("skills", []))
            if skills:
                detalle.append(f"  - Sus convenciones están en las skills {en_codigo(skills)} (en `.agents/skills/`).")

        definidos, faltan = [], []
        for verbos, orden in perfil_proyecto.MAKE_DE_VERBOS:
            con_comando = [v for v in verbos if perfil_proyecto.comando_de(parte, v)]
            if con_comando:
                definidos.append(f"make {orden} PART={parte['nombre']}")
            faltan += [v for v in verbos if v not in con_comando]
        if definidos:
            detalle.append(f"  - Comandos: {en_codigo(definidos)}.")
        if faltan:
            detalle.append(f"  - El proyecto todavía no tiene cómo: {', '.join(faltan)}. No inventes un comando: dilo en tu entrega.")
        if parte.get("terceros"):
            detalle.append(f"  - Código de terceros, que no se toca: {en_codigo(dentro_de(parte, parte['terceros']))}.")
        if parte.get("inmutables"):
            detalle.append("  - No se modifican una vez versionados (se crea un archivo nuevo): "
                           f"{en_codigo(dentro_de(parte, parte['inmutables']))}.")

    if agente["proyecto"] == "mapa":
        lineas.append(f"{origen} Sus partes:")
        lineas += ["", *detalle]
    elif detalle:
        lineas.append(f"{origen} Tus partes:")
        lineas += ["", *detalle]
        if ajenas:
            lineas += ["", f"Las demás partes no son tuyas: {', '.join(ajenas)}."]
    else:
        lineas.append(f"{origen} Hoy no te asigna ninguna parte. Si te delegan una tarea, no supongas carpetas ni "
                      "tecnologías: avisa al orquestador.")
    agente["skills"] = list(dict.fromkeys(agente["skills"] + propias))
    agente["instrucciones"] += "\n\n" + "\n".join(lineas)


NOMBRE_COMANDO = re.compile(r"[a-z0-9]+(-[a-z0-9]+)*")


def leer_comando(ruta: Path) -> dict:
    """Un comando de equipo/comandos/: metadatos (nombre, descripcion, argumentos opcional) e instrucciones."""
    texto = ruta.read_text(encoding="utf-8")
    ruta_txt = ruta.relative_to(RAIZ).as_posix()
    if not texto.startswith("---\n"):
        raise ValueError(f"{ruta_txt}: falta el bloque de metadatos (---)")
    _, cabecera, cuerpo = texto.split("---\n", 2)
    datos = {}
    for linea in cabecera.strip().splitlines():
        clave, _, valor = linea.partition(":")
        datos[clave.strip()] = valor.strip()

    for campo in ("nombre", "descripcion"):
        if not datos.get(campo):
            raise ValueError(f"{ruta_txt}: falta el campo '{campo}'")
    if datos["nombre"] != ruta.stem:
        raise ValueError(f"{ruta_txt}: 'nombre' debe coincidir con el nombre del archivo")
    if not NOMBRE_COMANDO.fullmatch(datos["nombre"]):
        raise ValueError(f"{ruta_txt}: 'nombre' solo admite minúsculas, números y guiones (ej. update-kit)")
    if not cuerpo.strip():
        raise ValueError(f"{ruta_txt}: el comando no tiene instrucciones")
    if "$ARGUMENTS" in cuerpo:
        raise ValueError(f"{ruta_txt}: no escribas $ARGUMENTS; declara 'argumentos' y el generador lo agrega "
                         "en las herramientas que lo admiten")
    return {
        "corto": datos["nombre"],
        "nombre": PREFIJO_COMANDOS + datos["nombre"],
        "descripcion": datos["descripcion"],
        "argumentos": datos.get("argumentos", ""),
        "instrucciones": cuerpo.strip(),
        "fuente": ruta_txt,
    }


def cuerpo_comando(c: dict, herramienta: str) -> str:
    """Instrucciones del comando con el aviso de generado y, si los declara, sus argumentos."""
    texto = f"<!-- {AVISO.format(fuente=c['fuente'])} -->\n\n{c['instrucciones']}\n"
    if not c["argumentos"]:
        return texto
    if herramienta == "codex":      # Codex no sustituye marcadores: los argumentos se leen de la conversación.
        return texto + f"\nSi la persona escribió algo junto a `${c['nombre']}`, esos son los argumentos: {c['argumentos']}.\n"
    return texto + f"\nArgumentos que escribió la persona junto al comando ({c['argumentos']}; puede venir vacío): $ARGUMENTS\n"


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

def es_generada(constitucion: str) -> bool:
    """True si ese texto de la constitución lo escribió este script (y no una persona ni Spec Kit)."""
    return constitucion.startswith("<!-- GENERADO")


def generar_constitucion() -> dict[str, bytes]:
    """Los principios del kit y, debajo, las reglas del proyecto. Sin principios no se genera nada."""
    if not (RAIZ / PRINCIPIOS).exists():
        return {}
    partes = [(RAIZ / f).read_text(encoding="utf-8").strip() for f in (PRINCIPIOS, REGLAS) if (RAIZ / f).exists()]
    aviso = AVISO.format(fuente=f"{PRINCIPIOS} y {REGLAS}")
    return {CONSTITUCION: (f"<!-- {aviso} -->\n" + "\n\n".join(partes) + "\n").encode()}


def es_repositorio_del_kit() -> bool:
    """True en el repositorio del kit (tiene el instalador y MANTENER-KIT.md, y no es un proyecto con el kit instalado)."""
    return ((RAIZ / "scripts/instalar_kit.py").exists() and (RAIZ / "MANTENER-KIT.md").exists()
            and not (RAIZ / ".kit-manifest.json").exists())


def generar_claude(agentes: list[dict], comandos: list[dict], config: dict) -> dict[str, bytes]:
    salida: dict[str, bytes] = {}
    modelos = config_modelos(config, "claude")

    if es_repositorio_del_kit():
        # En el repositorio del kit, Claude Code mantiene el kit: no es el orquestador de un proyecto.
        salida["CLAUDE.md"] = (
            f"<!-- {AVISO.format(fuente='MANTENER-KIT.md')} -->\n"
            "**Estás en el repositorio del kit, no en un proyecto.** Tu rol es mantener y mejorar el kit. "
            "`AGENTS.md`, `equipo/orquestador.md` y las skills son el producto que se instala en los proyectos: "
            "no son instrucciones para ti. Sigue este manual:\n\n"
            "@MANTENER-KIT.md\n"
        ).encode()
    else:
        salida["CLAUDE.md"] = (
            f"<!-- {AVISO.format(fuente='AGENTS.md')} -->\n"
            "Las instrucciones de este proyecto están en AGENTS.md (compartido con otras herramientas), "
            "y tu manual de trabajo como orquestador en equipo/orquestador.md:\n\n"
            "@AGENTS.md\n\n@equipo/orquestador.md\n"
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

    # Skills: copia de .agents/skills para Claude Code (sin los comandos generados ahí para Codex).
    origen = RAIZ / ".agents/skills"
    for archivo in sorted(origen.rglob("*")):
        if archivo.is_file():
            rel = archivo.relative_to(origen).as_posix()
            if not rel.startswith((PREFIJO_COMANDOS, DE_SPEC_KIT)):
                salida[f".claude/skills/{rel}"] = archivo.read_bytes()

    # Comandos: skills que solo se cargan cuando la persona las escribe (su descripción no ocupa contexto).
    for c in comandos:
        cab = ["---", f"name: {c['nombre']}", f"description: {yaml_texto(c['descripcion'])}", "disable-model-invocation: true"]
        if c["argumentos"]:
            cab.append(f"argument-hint: {yaml_texto(c['argumentos'])}")
        cab.append("---")
        salida[f".claude/skills/{c['nombre']}/SKILL.md"] = ("\n".join(cab) + "\n" + cuerpo_comando(c, "claude")).encode()

    ajustes_ruta = RAIZ / "equipo/adaptadores/claude/settings.json"
    ajustes = json.loads(ajustes_ruta.read_text(encoding="utf-8")) if ajustes_ruta.exists() else {}
    if modelos["orquestador"]:
        ajustes["model"] = modelos["orquestador"]
    if ajustes:
        salida[".claude/settings.json"] = json_bytes(ajustes)
    return salida


def generar_codex(agentes: list[dict], comandos: list[dict], config: dict) -> dict[str, bytes]:
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

    # Comandos: skills que se invocan con $nombre y que Codex no carga por su cuenta.
    for c in comandos:
        carpeta = COMANDOS_CODEX + c["corto"]
        cab = ["---", f"name: {c['nombre']}", f"description: {yaml_texto(c['descripcion'])}", "---"]
        salida[f"{carpeta}/SKILL.md"] = ("\n".join(cab) + "\n" + cuerpo_comando(c, "codex")).encode()
        salida[f"{carpeta}/agents/openai.yaml"] = (
            f"# {AVISO.format(fuente=c['fuente'])}\npolicy:\n  allow_implicit_invocation: false\n").encode()
    return salida


def generar_opencode(agentes: list[dict], comandos: list[dict], config: dict) -> dict[str, bytes]:
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

    # Agente principal propio (reemplaza a los integrados build/plan en la interfaz de OpenCode).
    principal = config_principal_opencode(config)
    if principal["nombre"]:
        cab = ["---", f"description: {yaml_texto('Orquestador del equipo: coordina Spec Kit y los subagentes según AGENTS.md.')}",
               "mode: primary"]
        if modelos["orquestador"]:
            cab.append(f"model: {modelos['orquestador']}")
        if principal["temperatura"] is not None:
            cab.append(f"temperature: {principal['temperatura']:g}")
        cab += ["permission:", "  edit: allow", "  bash: allow", "  webfetch: allow", "  task: allow", "  skill: allow", "---"]
        cuerpo = (RAIZ / principal["fuente"]).read_text(encoding="utf-8").strip()
        texto = "\n".join(cab) + f"\n<!-- {AVISO.format(fuente=principal['fuente'])} -->\n\n{cuerpo}\n"
        salida[f".opencode/agents/{principal['nombre']}.md"] = texto.encode()

    for c in comandos:
        cab = ["---", f"description: {yaml_texto(c['descripcion'])}", "---"]
        salida[f".opencode/commands/{c['nombre']}.md"] = ("\n".join(cab) + "\n" + cuerpo_comando(c, "opencode")).encode()

    # opencode.json: modelos y agentes integrados ocultos, sobre la base del adaptador si existe.
    base_ruta = RAIZ / "equipo/adaptadores/opencode/opencode.json"
    base = json.loads(base_ruta.read_text(encoding="utf-8")) if base_ruta.exists() else {}
    ocultar = principal["ocultar"] if principal["nombre"] else []
    if modelos["orquestador"] or modelos["ligero"] or base or principal["nombre"]:
        datos = {"$schema": "https://opencode.ai/config.json", **base}
        if modelos["orquestador"]:
            datos["model"] = modelos["orquestador"]
        if modelos["ligero"]:
            datos["small_model"] = modelos["ligero"]
        if principal["nombre"]:
            datos["default_agent"] = principal["nombre"]   # el orquestador es el agente al abrir OpenCode
        if ocultar:
            agentes_cfg = dict(datos.get("agent", {}))
            for nombre in ocultar:
                agentes_cfg[nombre] = {**agentes_cfg.get(nombre, {}), "disable": True}
            datos["agent"] = agentes_cfg
        salida["opencode.json"] = json_bytes(datos)
    return salida


def config_principal_opencode(config: dict) -> dict:
    """Sección opcional modelos.opencode.agente_principal:
      "agente_principal": {
        "nombre": "orquestador",            # "" = usar los agentes integrados de OpenCode (por defecto: orquestador)
        "ocultar": [],                      # agentes integrados que no se muestran (ej. ["build", "plan"])
        "temperatura": 0.2                  # opcional
      }
    Las instrucciones salen de equipo/orquestador.md.
    """
    crudo = (config.get("modelos", {}).get("opencode", {}) or {}).get("agente_principal", {}) or {}
    temp = crudo.get("temperatura")
    return {
        "nombre": crudo.get("nombre", "orquestador"),   # por defecto activo; "" lo desactiva
        "ocultar": list(crudo.get("ocultar", [])),
        "temperatura": float(temp) if isinstance(temp, (int, float)) else None,
        "fuente": "equipo/orquestador.md",
    }


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
        etiqueta = "orquestador"
        if h == "opencode" and config_principal_opencode(config)["nombre"]:
            p = config_principal_opencode(config)
            etiqueta = f"{p['nombre']} (principal)"
        print(f"  {etiqueta:<20} {m['orquestador'] or '(modelo de la sesión)'}")
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

def carpetas_gestionadas(g: dict) -> list[Path]:
    """Las carpetas de una herramienta que existen ahora, con los comodines resueltos."""
    return [c for d in g["dirs"] for c in sorted(RAIZ.glob(d)) if c.is_dir()]


def propios(base: Path) -> list[Path]:
    """Lo que hay en una carpeta gestionada, sin lo que dejó Spec Kit."""
    return [f for f in sorted(base.iterdir()) if not f.name.startswith(DE_SPEC_KIT)]


def archivos_actuales() -> dict[str, bytes]:
    actuales: dict[str, bytes] = {}
    for g in GESTIONADOS.values():
        for base in carpetas_gestionadas(g):
            for propio in propios(base):
                for f in [propio, *propio.rglob("*")]:
                    if f.is_file():
                        actuales[f.relative_to(RAIZ).as_posix()] = f.read_bytes()
        for f in g["files"]:
            p = RAIZ / f
            if p.exists():
                actuales[f] = p.read_bytes()
    if (RAIZ / PRINCIPIOS).exists() and (RAIZ / CONSTITUCION).exists():
        actuales[CONSTITUCION] = (RAIZ / CONSTITUCION).read_bytes()
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
        comandos = [leer_comando(p) for p in sorted((RAIZ / "equipo/comandos").glob("*.md"))]
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1
    errores = validar_config(config, agentes)
    if errores:
        for e in errores:
            print(f"Error: {e}", file=sys.stderr)
        return 1

    perfil, errores = leer_perfil()
    for e in errores:
        print(f"Aviso: {e}", file=sys.stderr)
    if errores:
        print("Aviso: con el perfil así, los roles se generan sin los datos del proyecto. Corrígelo (make profile) "
              "y ejecuta `make sync`.", file=sys.stderr)
    for a in agentes:
        aplicar_perfil(a, perfil, bool(errores))

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

    esperado: dict[str, bytes] = generar_constitucion()
    for h in activas:
        esperado.update(GENERADORES[h](agentes, comandos, config))

    if "--verificar" in args:
        actuales = archivos_actuales()
        faltan = sorted(set(esperado) - set(actuales))
        sobran = sorted(set(actuales) - set(esperado))
        distintos = sorted(k for k in set(esperado) & set(actuales) if esperado[k] != actuales[k])
        if faltan or sobran or distintos:
            for titulo, lista in (("Faltan", faltan), ("Sobran", sobran), ("Desactualizados", distintos)):
                for k in lista:
                    print(f"  {titulo}: {k}", file=sys.stderr)
            print("Los archivos generados no están al día. Ejecuta `make sync`.", file=sys.stderr)
            return 1
        print(f"OK: {len(esperado)} archivos generados al día ({', '.join(activas)}).")
        return 0

    # Limpia todo lo gestionado (también de herramientas desactivadas) y vuelve a escribir.
    for g in GESTIONADOS.values():
        for d in carpetas_gestionadas(g):
            for f in propios(d):
                if f.is_dir():
                    shutil.rmtree(f, ignore_errors=True)
                else:
                    f.unlink(missing_ok=True)
            if not any(d.iterdir()):
                d.rmdir()
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
    print("Revisa qué modelo usa cada agente con: make models")
    return 0


if __name__ == "__main__":
    sys.exit(main())
