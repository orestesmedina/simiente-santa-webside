#!/usr/bin/env python3
"""Perfil del proyecto: qué partes tiene, en qué carpeta está cada una y cómo se ejecuta cada verbo.

El perfil vive en equipo/perfil.json y es del proyecto (el kit no lo instala ni lo reemplaza):

  {
    "partes": [
      {
        "nombre": "api",                      minúsculas, números y guiones; único
        "carpeta": "backend",                 relativa a la raíz ("." si el proyecto es una sola parte)
        "descripcion": "API REST",            opcional
        "roles": [                            opcional: los agentes de equipo/agentes/ que la trabajan
          "dev-backend",                        solo el nombre, o
          {"rol": "dev-frontend", "skills": ["react-frontend"]}     el nombre y las skills que son solo suyas
        ],
        "skills": ["go-backend"],             opcional: skills para todos los roles de la parte (del catálogo del
                                              kit, que `make install-kit` copia, o propias en .agents/skills/)
        "terceros": ["vendor"],               opcional: carpetas con código ajeno (relativas a la parte)
        "inmutables": ["migrations/*.sql"],   opcional: archivos que no se modifican una vez versionados
        "verbos": {                           opcional: comando de cada verbo, ejecutado dentro de la carpeta
          "probar": "go test ./...",
          "cobertura": null                   null o ausente: el proyecto todavía no lo tiene
        }
      }
    ],
    "requisitos": [                           opcional: lo que la máquina necesita para trabajar en el proyecto
      "php",                                    solo el programa, o
      {"comando": "go", "nombre": "Go", "minima": "1.26", "version": "go version"},
      {"comando": "golangci-lint", "opcional": true, "instalar": "go install …"}
    ]
  }

"rol": "dev-backend" (un solo rol, el formato de la primera versión del perfil) equivale a "roles": ["dev-backend"].

Cada comando recibe dos variables de entorno: PERFIL_RAIZ (ruta absoluta de la raíz del proyecto) y
PERFIL_TERCEROS (las carpetas de terceros de la parte, una por línea). Los scripts propios del proyecto que
un verbo necesite van en tools/ (scripts/ es del kit).

"requisitos" es lo que revisa `make doctor`, además de las herramientas que el kit necesita siempre. De cada uno:
"comando" (el programa que debe estar en el PATH), "nombre" (cómo se muestra), "minima" (versión mínima),
"version" (el comando que imprime la versión; por defecto, `<comando> --version`), "opcional" (true: si falta es
un aviso y no un problema) e "instalar" (cómo se instala, para mostrarlo cuando falta).

Sin perfil, el kit no supone nada del proyecto: los comandos que ejecutan verbos piden crearlo (/bowser-profile).

Uso:
  python3 scripts/perfil.py              muestra el perfil y lo valida
  python3 scripts/perfil.py --verificar  solo valida (para hooks y CI)
  python3 scripts/perfil.py --detectar   mira el proyecto y muestra, en JSON, lo que encontró (no escribe nada)
  python3 scripts/perfil.py --requisitos los requisitos, uno por línea, para scripts/doctor.sh

Solo usa la biblioteca estándar de Python 3.9+.
"""
from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from pathlib import Path

import catalogo

RAIZ =Path(__file__).resolve().parent.parent
RUTA = "equipo/perfil.json"

# Lo único que el núcleo sabe pedir. El comando de cada uno lo declara el proyecto.
VERBOS = {
    "formato": "comprueba que el código esté formateado (rápido: corre en cada commit)",
    "revisar": "análisis estático: linters, tipos, sintaxis",
    "probar": "ejecuta las pruebas",
    "cobertura": "mide la cobertura de pruebas y falla bajo el mínimo del proyecto",
    "auditar": "busca vulnerabilidades en las dependencias",
    "generar": "regenera el código generado",
}
# Con qué comando de make se ejecuta cada verbo (el mismo para todos los proyectos).
MAKE_DE_VERBOS = (
    (("formato", "revisar"), "lint"),
    (("probar",), "test"),
    (("cobertura",), "coverage"),
    (("generar",), "generate"),
    (("auditar",), "security"),
)
CAMPOS = {"nombre", "carpeta", "descripcion", "rol", "roles", "skills", "terceros", "inmutables", "verbos"}
NOMBRE = re.compile(r"[a-z0-9]+(-[a-z0-9]+)*")
CAMPOS_REQUISITO = {"comando", "nombre", "minima", "version", "opcional", "instalar"}
PROGRAMA = re.compile(r"[A-Za-z0-9][A-Za-z0-9._+-]*")
VERSION_MINIMA = re.compile(r"\d+(\.\d+){0,3}")


class ErrorPerfil(Exception):
    """El perfil no se puede leer."""


# ---------------------------------------------------------------- lectura y validación

def cargar(raiz: Path = RAIZ) -> dict | None:
    """El perfil, o None si el proyecto no tiene."""
    ruta = raiz / RUTA
    if not ruta.exists():
        return None
    try:
        datos = json.loads(ruta.read_text(encoding="utf-8"))
    except (ValueError, OSError) as e:
        raise ErrorPerfil(f"{RUTA} no es un JSON válido: {e}")
    if not isinstance(datos, dict):
        raise ErrorPerfil(f"{RUTA} debe ser un objeto con la lista \"partes\"")
    return datos


def ruta_segura(valor) -> bool:
    """Ruta relativa que no sale del proyecto."""
    if not isinstance(valor, str) or not valor.strip():
        return False
    p = valor.replace("\\", "/")
    return (not p.startswith("/") and not re.match(r"^[A-Za-z]:", p) and ".." not in p.split("/")
            and not any(ord(c) < 32 for c in p))


def lista_de_textos(valor) -> bool:
    return isinstance(valor, list) and all(isinstance(x, str) and x.strip() for x in valor)


def una_linea(valor) -> bool:
    return isinstance(valor, str) and bool(valor.strip()) and not any(ord(c) < 32 for c in valor)


def validar_requisitos(perfil: dict) -> list[str]:
    requisitos = perfil.get("requisitos", [])
    if not isinstance(requisitos, list):
        return ["\"requisitos\" debe ser una lista de programas (ej. [\"php\", {\"comando\": \"go\", \"minima\": \"1.26\"}])"]
    errores: list[str] = []
    vistos: set[str] = set()
    for i, r in enumerate(requisitos, 1):
        if isinstance(r, str):
            r = {"comando": r}
        if not isinstance(r, dict):
            errores.append(f"el requisito {i} debe ser el nombre de un programa o un objeto con \"comando\"")
            continue
        comando = r.get("comando")
        if not isinstance(comando, str) or not PROGRAMA.fullmatch(comando):
            errores.append(f"el requisito {i}: \"comando\" debe ser el nombre de un programa, sin espacios ni rutas (ej. go, php)")
            continue
        quien = f"el requisito «{comando}»"
        if comando in vistos:
            errores.append(f"{quien} está dos veces")
        vistos.add(comando)
        for clave in r:
            if clave not in CAMPOS_REQUISITO and not clave.startswith("_"):
                errores.append(f"{quien}: campo desconocido «{clave}» (campos: {', '.join(sorted(CAMPOS_REQUISITO))})")
        if "minima" in r and not (isinstance(r["minima"], str) and VERSION_MINIMA.fullmatch(r["minima"])):
            errores.append(f"{quien}: \"minima\" debe ser una versión entre comillas (ej. \"1.26\")")
        for clave in ("nombre", "version", "instalar"):
            if clave in r and not una_linea(r[clave]):
                errores.append(f"{quien}: \"{clave}\" debe ser un texto de una sola línea")
        if not isinstance(r.get("opcional", False), bool):
            errores.append(f"{quien}: \"opcional\" debe ser true o false")
    return errores


def requisitos_de(perfil: dict) -> list[dict]:
    """Lo que la máquina necesita para trabajar en el proyecto, con todos sus campos. Para un perfil ya validado."""
    resultado = []
    for r in perfil.get("requisitos", []):
        r = {"comando": r} if isinstance(r, str) else r
        resultado.append({
            "comando": r["comando"], "nombre": r.get("nombre") or r["comando"], "minima": r.get("minima", ""),
            "version": r.get("version") or f"{r['comando']} --version", "opcional": bool(r.get("opcional")),
            "instalar": r.get("instalar", ""),
        })
    return resultado


def validar(perfil: dict, raiz: Path = RAIZ) -> tuple[list[str], list[str]]:
    """Devuelve (errores, avisos). Con errores, el perfil no se usa."""
    errores: list[str] = validar_requisitos(perfil)
    avisos: list[str] = []
    partes = perfil.get("partes")
    if not isinstance(partes, list) or not partes:
        return [f"{RUTA}: \"partes\" debe ser una lista con al menos una parte"], avisos

    nombres: set[str] = set()
    carpetas: set[str] = set()
    en_catalogo, _ = catalogo.leer(catalogo.carpeta(raiz))
    for i, parte in enumerate(partes, 1):
        if not isinstance(parte, dict):
            errores.append(f"la parte {i} debe ser un objeto")
            continue
        nombre = parte.get("nombre")
        quien = f"la parte «{nombre}»" if isinstance(nombre, str) and nombre else f"la parte {i}"
        if not isinstance(nombre, str) or not NOMBRE.fullmatch(nombre):
            errores.append(f"{quien}: \"nombre\" solo admite minúsculas, números y guiones (ej. api, panel-web)")
        elif nombre in nombres:
            errores.append(f"hay dos partes con el nombre «{nombre}»")
        else:
            nombres.add(nombre)

        for clave in parte:
            if clave not in CAMPOS and not clave.startswith("_"):
                errores.append(f"{quien}: campo desconocido «{clave}» (campos: {', '.join(sorted(CAMPOS))})")

        carpeta = parte.get("carpeta")
        if not ruta_segura(carpeta):
            errores.append(f"{quien}: \"carpeta\" debe ser una ruta relativa dentro del proyecto (\".\" para la raíz)")
            carpeta = None
        else:
            normal = os.path.normpath(carpeta).replace("\\", "/")
            if not (raiz / carpeta).is_dir():
                errores.append(f"{quien}: la carpeta «{carpeta}» no existe")
            if normal in carpetas:
                errores.append(f"{quien}: la carpeta «{carpeta}» ya pertenece a otra parte")
            carpetas.add(normal)

        skills = parte.get("skills", [])
        if not lista_de_textos(skills):
            errores.append(f"{quien}: \"skills\" debe ser una lista de nombres de skills")
            skills = []
        nombradas = list(skills)

        if "rol" in parte and "roles" in parte:
            errores.append(f"{quien}: usa \"roles\" (lista) o \"rol\" (uno solo), no los dos")
        roles = parte["roles"] if "roles" in parte else [parte["rol"]] if "rol" in parte else []
        if not isinstance(roles, list):
            errores.append(f"{quien}: \"roles\" debe ser una lista de roles (ej. [\"dev-backend\"])")
            roles = []
        vistos: set[str] = set()
        for r in roles:
            propias = []
            if isinstance(r, dict):
                sobran = sorted(k for k in r if k not in ("rol", "skills") and not k.startswith("_"))
                if sobran:
                    errores.append(f"{quien}: un rol solo admite \"rol\" y \"skills\" (sobra «{sobran[0]}»)")
                r, propias = r.get("rol"), r.get("skills", [])
            if not isinstance(r, str) or not (raiz / "equipo/agentes" / f"{r}.md").exists():
                errores.append(f"{quien}: el rol «{r}» no existe en equipo/agentes/")
                continue
            if r in vistos:
                errores.append(f"{quien}: el rol «{r}» está dos veces")
            vistos.add(r)
            if not lista_de_textos(propias):
                errores.append(f"{quien}: las \"skills\" del rol «{r}» deben ser una lista de nombres de skills")
            else:
                nombradas += propias

        for s in dict.fromkeys(nombradas):
            if (raiz / catalogo.INSTALADAS / s / "SKILL.md").exists():
                continue
            if s in en_catalogo:
                avisos.append(f"{quien}: la skill «{s}» está en el catálogo del kit pero todavía no en el proyecto "
                              "(ejecuta: make install-kit)")
            else:
                avisos.append(f"{quien}: la skill «{s}» no está en .agents/skills/ ni en el catálogo del kit "
                              "(el agente trabajará sin ella)")

        terceros = parte.get("terceros", [])
        if not lista_de_textos(terceros) or not all(ruta_segura(t) for t in terceros):
            errores.append(f"{quien}: \"terceros\" debe ser una lista de carpetas relativas a la parte")
        elif carpeta:
            for t in terceros:
                if not (raiz / carpeta / t).exists():
                    avisos.append(f"{quien}: la carpeta de terceros «{t}» no existe")

        inmutables = parte.get("inmutables", [])
        if not lista_de_textos(inmutables) or not all(ruta_segura(p) for p in inmutables):
            errores.append(f"{quien}: \"inmutables\" debe ser una lista de patrones relativos a la parte (ej. migrations/*.sql)")

        verbos = parte.get("verbos", {})
        if not isinstance(verbos, dict):
            errores.append(f"{quien}: \"verbos\" debe ser un objeto verbo → comando")
        else:
            for verbo, comando in verbos.items():
                if verbo.startswith("_"):
                    continue
                if verbo not in VERBOS:
                    errores.append(f"{quien}: verbo desconocido «{verbo}» (verbos: {', '.join(VERBOS)})")
                elif comando is not None and (not isinstance(comando, str) or not comando.strip()):
                    errores.append(f"{quien}: el verbo «{verbo}» debe ser un comando (texto) o null si el proyecto no lo tiene")
    return errores, avisos


def comando_de(parte: dict, verbo: str) -> str | None:
    comando = (parte.get("verbos") or {}).get(verbo)
    return comando if isinstance(comando, str) and comando.strip() else None


def roles_de(parte: dict) -> list[dict]:
    """Los roles que trabajan la parte, cada uno con sus skills: las de la parte más las que son solo suyas.
    Para un perfil ya validado."""
    comunes = parte.get("skills", [])
    roles = parte["roles"] if "roles" in parte else [parte["rol"]] if "rol" in parte else []
    resultado = []
    for r in roles:
        nombre, propias = (r["rol"], r.get("skills", [])) if isinstance(r, dict) else (r, [])
        skills = list(dict.fromkeys([*comunes, *propias]))
        # Primero el estándar y después lo propio del proyecto, las nombre donde las nombre el perfil.
        skills.sort(key=lambda s: s.startswith(CONVENCIONES))
        resultado.append({"rol": nombre, "skills": skills})
    return resultado


def skills_nombradas(perfil) -> list[str]:
    """Todas las skills que el perfil nombra, en las partes y en sus roles, sin repetir.
    No exige un perfil validado: lo que no tenga la forma esperada se salta."""
    nombres: list[str] = []
    partes = perfil.get("partes") if isinstance(perfil, dict) else None
    for parte in partes if isinstance(partes, list) else []:
        if not isinstance(parte, dict):
            continue
        listas = [parte.get("skills")]
        roles = parte.get("roles")
        listas += [r.get("skills") for r in (roles if isinstance(roles, list) else []) if isinstance(r, dict)]
        for lista in listas:
            nombres += [s for s in (lista if isinstance(lista, list) else []) if isinstance(s, str) and s.strip()]
    return list(dict.fromkeys(nombres))


def perfil_valido(raiz: Path = RAIZ) -> dict | None:
    """El perfil ya validado, o None si no hay. Termina con error (código 1) si está mal escrito."""
    try:
        perfil = cargar(raiz)
    except ErrorPerfil as e:
        print(f"✗ {e}", file=sys.stderr)
        sys.exit(1)
    if perfil is None:
        return None
    errores, _ = validar(perfil, raiz)
    if errores:
        for e in errores:
            print(f"✗ {RUTA}: {e}", file=sys.stderr)
        print("Corrige el perfil (o pide en el chat: /bowser-profile).", file=sys.stderr)
        sys.exit(1)
    return perfil


# ---------------------------------------------------------------- mostrar

# Prefijo de la skill con las convenciones propias de una parte (la redacta el comando bowser-conventions).
CONVENCIONES = "convenciones-de-"


def mostrar(perfil: dict) -> None:
    partes = perfil["partes"]
    print(f"Perfil del proyecto ({RUTA}): {len(partes)} parte{'s' if len(partes) != 1 else ''}\n")
    for parte in partes:
        print(f"  {parte['nombre']}  ·  carpeta {parte['carpeta']}")
        if parte.get("descripcion"):
            print(f"    {parte['descripcion']}")
        roles = roles_de(parte)
        if roles:
            print("    roles:      " + "  ·  ".join(
                r["rol"] + (f" ({', '.join(r['skills'])})" if r["skills"] else "") for r in roles))
        elif parte.get("skills"):
            print(f"    skills:     {', '.join(parte['skills'])} (ningún rol las recibe: la parte no tiene roles)")
        for verbo in VERBOS:
            print(f"    {verbo:<10} {comando_de(parte, verbo) or '— sin definir'}")
        if parte.get("terceros"):
            print(f"    terceros:   {', '.join(parte['terceros'])}")
        if parte.get("inmutables"):
            print(f"    inmutables: {', '.join(parte['inmutables'])}")
        print()
    sin_convenciones = [p["nombre"] for p in partes
                        if roles_de(p) and CONVENCIONES + p["nombre"] not in {s for r in roles_de(p) for s in r["skills"]}]
    if sin_convenciones:
        print(f"Sin convenciones propias: {', '.join(sin_convenciones)}. Las skills del catálogo dicen el estándar de cada "
              "tecnología; lo propio del proyecto se redacta con /bowser-conventions.")
    requisitos = requisitos_de(perfil)
    if requisitos:
        print("Requisitos de la máquina (los revisa make doctor): " + "  ·  ".join(
            r["nombre"] + (f" {r['minima']} o superior" if r["minima"] else "") + (" (recomendado)" if r["opcional"] else "")
            for r in requisitos))
    else:
        print("Sin requisitos: make doctor no revisa las herramientas de la tecnología del proyecto. "
              "Se declaran en \"requisitos\" (/bowser-profile).")


# ---------------------------------------------------------------- detección

IGNORAR = {".git", "node_modules", "vendor", "bower_components", ".venv", "venv", "__pycache__", ".idea", ".vscode",
           "target", ".next", ".cache", "coverage"}
TERCEROS_SEGUROS = {"node_modules", "vendor", "bower_components"}
TERCEROS_A_REVISAR = {"libraries", "lib", "libs", "third_party", "third-party", "plugins", "dist", "build"}
MARCADORES = {
    "go.mod": "Go", "package.json": "Node.js", "composer.json": "PHP con Composer", "pyproject.toml": "Python",
    "requirements.txt": "Python", "Cargo.toml": "Rust", "pom.xml": "Java (Maven)", "build.gradle": "Java o Kotlin (Gradle)",
    "build.gradle.kts": "Kotlin (Gradle)", "Gemfile": "Ruby", "artisan": "Laravel", "manage.py": "Django",
    "bower.json": "Bower (dependencias web antiguas)", "pubspec.yaml": "Dart o Flutter", "mix.exs": "Elixir",
}
HERRAMIENTAS = (
    "Dockerfile", "docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml", "phpunit.xml",
    "phpunit.xml.dist", ".php-cs-fixer.php", ".php-cs-fixer.dist.php", "phpstan.neon", "phpcs.xml", ".golangci.yml",
    ".golangci.yaml", "sqlc.yaml", "sqlc.yml", ".eslintrc", ".eslintrc.js", ".eslintrc.json", ".eslintrc.cjs",
    "eslint.config.js", "eslint.config.mjs", ".prettierrc", ".prettierrc.json", "prettier.config.js", "tsconfig.json",
    "vite.config.ts", "vite.config.js", "vitest.config.ts", "jest.config.js", "playwright.config.ts", "pytest.ini",
    "tox.ini", "ruff.toml", ".htaccess", "Makefile", "proyecto.mk",
)
LENGUAJES = {
    ".go": "Go", ".php": "PHP", ".ts": "TypeScript", ".tsx": "TypeScript", ".js": "JavaScript", ".jsx": "JavaScript",
    ".mjs": "JavaScript", ".py": "Python", ".cs": "C#", ".java": "Java", ".kt": "Kotlin", ".rb": "Ruby", ".rs": "Rust",
    ".swift": "Swift", ".dart": "Dart", ".vue": "Vue", ".svelte": "Svelte", ".sql": "SQL", ".html": "HTML",
    ".css": "CSS", ".scss": "Sass", ".less": "Less", ".sh": "Shell",
}
PAQUETES_NODE = ("react", "vue", "@angular/core", "svelte", "next", "nuxt", "express", "vite", "typescript", "vitest",
                 "jest", "eslint", "prettier", "@playwright/test")


def archivos_del_proyecto(raiz: Path) -> tuple[list[str], bool]:
    """Rutas relativas de los archivos propios del proyecto (sin el kit) y si es un repositorio de git."""
    es_git = (raiz / ".git").exists()
    rutas: list[str] = []
    if es_git:
        r = subprocess.run(["git", "ls-files", "-co", "--exclude-standard"], cwd=raiz, capture_output=True)
        if r.returncode == 0:
            rutas = [l for l in r.stdout.decode("utf-8", "replace").splitlines() if l]
    if not rutas:
        for base, carpetas, archivos in os.walk(raiz):
            carpetas[:] = [c for c in carpetas if c != ".git"]
            rel = os.path.relpath(base, raiz).replace("\\", "/")
            for a in archivos:
                rutas.append(a if rel == "." else f"{rel}/{a}")

    del_kit: set[str] = set()
    submodulo = ".bowser-spec-kit-ai"
    manifiesto = raiz / ".kit-manifest.json"
    if manifiesto.exists():
        try:
            datos = json.loads(manifiesto.read_text(encoding="utf-8"))
            del_kit = set(datos.get("archivos", {}))
            submodulo = datos.get("ruta_kit") or submodulo
        except ValueError:
            pass
    generados = (".claude/", ".codex/", ".opencode/", ".agents/", ".specify/", ".githooks/", "equipo/", "specs/", submodulo + "/")
    return sorted(r for r in rutas if r not in del_kit and not r.startswith(generados)
                  and r not in ("CLAUDE.md", "AGENTS.md", "opencode.json", ".kit-manifest.json")), es_git


def analizar_sql(ruta: Path) -> dict:
    try:
        texto = ruta.read_bytes()[:4_000_000].decode("utf-8", "replace")
    except OSError:
        return {}
    motor = "desconocido"
    if re.search(r"ENGINE\s*=\s*(InnoDB|MyISAM)|AUTO_INCREMENT|DELIMITER\s+(;;|\$\$|//)", texto, re.I):
        motor = "MySQL o MariaDB"
    elif re.search(r"\b(BIGSERIAL|SERIAL)\b|LANGUAGE\s+plpgsql|TIMESTAMPTZ|::[a-z]+", texto, re.I):
        motor = "PostgreSQL"
    elif re.search(r"\bIDENTITY\s*\(|\bNVARCHAR\b|^\s*GO\s*$", texto, re.I | re.M):
        motor = "SQL Server"
    elif re.search(r"\bAUTOINCREMENT\b|PRAGMA\s", texto, re.I):
        motor = "SQLite"
    return {
        "motor_probable": motor,
        "tablas": len(re.findall(r"CREATE\s+TABLE", texto, re.I)),
        "procedimientos_y_funciones": len(re.findall(r"CREATE\s+(?:OR\s+REPLACE\s+)?(?:DEFINER\s*=\s*\S+\s+)?(?:PROCEDURE|FUNCTION)", texto, re.I)),
        "disparadores": len(re.findall(r"CREATE\s+(?:OR\s+REPLACE\s+)?(?:DEFINER\s*=\s*\S+\s+)?TRIGGER", texto, re.I)),
    }


def leer_json(ruta: Path) -> dict:
    try:
        datos = json.loads(ruta.read_text(encoding="utf-8"))
        return datos if isinstance(datos, dict) else {}
    except (ValueError, OSError):
        return {}


def detectar(raiz: Path = RAIZ) -> dict:
    """Hechos del proyecto para que el arquitecto redacte el perfil. No decide nada ni escribe archivos."""
    rutas, es_git = archivos_del_proyecto(raiz)
    carpetas: dict[str, dict] = {}
    marcadores: list[dict] = []
    herramientas: list[str] = []
    terceros: set[str] = set()
    dudosas: dict[str, int] = {}
    sql: list[dict] = []
    workflows: list[str] = []

    for rel in rutas:
        tramos = rel.split("/")
        nombre = tramos[-1]
        ajeno = False
        for i, tramo in enumerate(tramos[:-1]):
            if tramo in TERCEROS_SEGUROS:
                terceros.add("/".join(tramos[:i + 1]))
                ajeno = True
                break
            if tramo in IGNORAR:
                ajeno = True
                break
        if ajeno:
            continue
        # Carpetas que a veces son código ajeno y a veces propio: se cuentan igual y se informan aparte.
        dudosa = None
        for i, tramo in enumerate(tramos[:-1]):
            if tramo in TERCEROS_A_REVISAR and i + 1 < len(tramos) - 1:
                dudosa = "/".join(tramos[:i + 2])
                dudosas[dudosa] = dudosas.get(dudosa, 0) + 1
                break
        if rel.startswith(".github/workflows/"):
            workflows.append(rel)
            continue

        if nombre in MARCADORES or nombre.endswith((".csproj", ".sln")):
            dato = {"archivo": rel, "indica": MARCADORES.get(nombre, ".NET")}
            if dudosa:
                # Puede ser el archivo de una biblioteca copiada dentro, no la tecnología del proyecto.
                dato["dentro_de_carpeta_a_confirmar"] = dudosa
            if nombre == "package.json":
                paquete = leer_json(raiz / rel)
                deps = {**paquete.get("dependencies", {}), **paquete.get("devDependencies", {})} \
                    if isinstance(paquete.get("dependencies", {}), dict) and isinstance(paquete.get("devDependencies", {}), dict) else {}
                dato["paquetes_conocidos"] = [p for p in PAQUETES_NODE if p in deps]
                dato["scripts"] = sorted(paquete.get("scripts", {})) if isinstance(paquete.get("scripts"), dict) else []
            if nombre == "composer.json":
                paquete = leer_json(raiz / rel)
                dato["requiere"] = sorted(paquete.get("require", {}))[:30] if isinstance(paquete.get("require"), dict) else []
                dato["scripts"] = sorted(paquete.get("scripts", {})) if isinstance(paquete.get("scripts"), dict) else []
            marcadores.append(dato)
        if nombre in HERRAMIENTAS:
            herramientas.append(rel)

        ext = os.path.splitext(nombre)[1].lower()
        if ext == ".sql" and len(sql) < 20:
            sql.append({"archivo": rel, **analizar_sql(raiz / rel)})
        lenguaje = LENGUAJES.get(ext)
        if lenguaje:
            arriba = tramos[0] if len(tramos) > 1 else "."
            dato = carpetas.setdefault(arriba, {"carpeta": arriba, "archivos": 0, "lenguajes": {}, "subcarpetas": set()})
            dato["archivos"] += 1
            dato["lenguajes"][lenguaje] = dato["lenguajes"].get(lenguaje, 0) + 1
            if len(tramos) > 2:
                dato["subcarpetas"].add(tramos[1])

    for dato in carpetas.values():
        dato["lenguajes"] = dict(sorted(dato["lenguajes"].items(), key=lambda kv: -kv[1]))
        dato["subcarpetas"] = sorted(dato["subcarpetas"])[:40]

    en_catalogo, _ = catalogo.leer(catalogo.carpeta(raiz))
    propias = {p.parent.name for p in (raiz / catalogo.INSTALADAS).glob("*/SKILL.md")
               if not p.parent.name.startswith(("bowser-", "equipo-", "perfil-"))}
    return {
        "repositorio_git": es_git,
        "perfil_actual": RUTA if (raiz / RUTA).exists() else None,
        "carpetas_con_codigo": sorted(carpetas.values(), key=lambda d: -d["archivos"]),
        "archivos_que_indican_tecnologia": marcadores,
        "archivos_de_herramientas": sorted(herramientas),
        "integracion_continua": sorted(workflows),
        "carpetas_de_terceros": sorted(terceros),
        "carpetas_a_confirmar_si_son_de_terceros": [
            {"carpeta": c, "archivos": n} for c, n in sorted(dudosas.items(), key=lambda kv: -kv[1])[:30]],
        "archivos_sql": sql,
        "roles_disponibles": sorted(p.stem for p in (raiz / "equipo/agentes").glob("*.md")),
        # Las del catálogo del kit (make install-kit copia las que el perfil nombre) y las propias del proyecto.
        "skills_disponibles": sorted(propias | set(en_catalogo)),
        "madurez_de_las_skills_del_catalogo": {n: s["madurez"] for n, s in en_catalogo.items()},
        "verbos": VERBOS,
        "nota": "Son hechos, no decisiones. Lo que no aparece aquí (cómo se prueba, cómo se levanta, qué versión se usa) "
                "se le pregunta a la persona: no se inventa.",
    }


# ---------------------------------------------------------------- principal

def main() -> int:
    args = sys.argv[1:]
    if "--detectar" in args:
        print(json.dumps(detectar(), ensure_ascii=False, indent=2))
        return 0

    try:
        perfil = cargar()
    except ErrorPerfil as e:
        print(f"✗ {e}", file=sys.stderr)
        return 1
    if "--requisitos" in args:
        # Para scripts/doctor.sh: un requisito por línea, con los campos separados por el carácter 0x1F.
        if perfil is None:
            return 0
        if validar(perfil)[0]:
            return 1
        for r in requisitos_de(perfil):
            print("\x1f".join([r["comando"], r["nombre"], r["minima"], r["version"], "1" if r["opcional"] else "", r["instalar"]]))
        return 0
    if perfil is None:
        if "--verificar" in args:
            print("Sin perfil: no hay nada que verificar.")
            return 0
        print(f"Este proyecto no tiene perfil ({RUTA}).")
        print("Sin perfil, el kit no sabe qué partes tiene el proyecto ni cómo se prueban: make test, make lint, "
              "make ci y los demás verbos no funcionan, y los agentes no reciben skills de tecnología.")
        print("Para crearlo, escribe en el chat: /bowser-profile")
        return 0

    errores, avisos = validar(perfil)
    if not errores and "--verificar" not in args:
        mostrar(perfil)
    for a in avisos:
        print(f"⚠ {a}", file=sys.stderr)
    if errores:
        for e in errores:
            print(f"✗ {RUTA}: {e}", file=sys.stderr)
        print("El perfil tiene errores: el kit no lo usará hasta que se corrijan.", file=sys.stderr)
        return 1
    print(f"OK: perfil válido ({len(perfil['partes'])} parte{'s' if len(perfil['partes']) != 1 else ''}).")
    return 0


if __name__ == "__main__":
    sys.exit(main())
