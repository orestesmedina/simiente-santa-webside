#!/usr/bin/env bash
# Verifica que el entorno de desarrollo tenga todo lo necesario. Uso: make doctor
set -uo pipefail

OK=0; FALTA=0; AVISO=0
DOCS="$(bash scripts/ruta-kit.sh 2>/dev/null || echo .bowser-spec-kit-ai)/docs"   # documentación del kit (en el submódulo)
verde() { printf "  \033[32m✓\033[0m %s\n" "$1"; OK=$((OK+1)); }
rojo()  { printf "  \033[31m✗\033[0m %s\n" "$1"; FALTA=$((FALTA+1)); }
ambar() { printf "  \033[33m!\033[0m %s\n" "$1"; AVISO=$((AVISO+1)); }

version_minima() { # $1 actual, $2 mínima
  [ "$(printf '%s\n%s\n' "$2" "$1" | sort -V | head -n1)" = "$2" ]
}

# Muchos instaladores dejan los programas en una carpeta bin/ de la carpeta personal (~/.local/bin y las de cada
# tecnología), que puede no estar en el PATH. Imprime la carpeta donde está el programa, o nada.
fuera_del_path() { # comando
  local f; [ -n "${HOME:-}" ] || return 0
  for f in "$HOME/bin/$1" "$HOME"/*/bin/"$1" "$HOME"/.[!.]*/bin/"$1" "$HOME"/.[!.]*/*/bin/"$1" "$HOME"/.[!.]*/*/*/bin/"$1"; do
    [ -f "$f" ] && [ -x "$f" ] && { echo "${f%/*}"; return; }
  done
}

requerido() { # nombre comando versión_mínima comando_versión [opcional: 1] [cómo se instala]
  local nombre=$1 cmd=$2 minima=$3 vcmd=$4 opcional=${5:-} instalar=${6:-}
  local falla=rojo pista=""
  [ -n "$opcional" ] && falla=ambar
  [ -n "$instalar" ] && pista=". Se instala con: $instalar"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    local carpeta; carpeta=$(fuera_del_path "$cmd")
    if [ -n "$carpeta" ]; then
      $falla "$nombre está en ${carpeta/#"$HOME"/\~}, pero esa carpeta no está en el PATH: agrega export PATH=\"${carpeta/#"$HOME"/\$HOME}:\$PATH\" al final de ~/.bashrc y abre otra terminal"
    else
      $falla "$nombre no está instalado${opcional:+ (recomendado)}$pista"
    fi
    return
  fi
  # La versión es el primer número después del nombre del programa; si ninguna línea lo nombra, el primero que
  # imprima. Hay programas que muestran antes otra versión (govulncheck -version empieza por la de Go).
  local salida v; salida=$(eval "$vcmd" </dev/null 2>/dev/null)
  v=$(printf '%s\n' "$salida" | awk -v c="$cmd" 'BEGIN { c = tolower(c) }
        { i = index(tolower($0), c); if (i) { print substr($0, i + length(c)); exit } }' \
      | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n1)
  [ -n "$v" ] || v=$(printf '%s\n' "$salida" | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n1)
  if [ -n "$minima" ] && [ -n "$v" ] && ! version_minima "$v" "$minima"; then
    $falla "$nombre $v (se requiere $minima o superior)$pista"
  else
    verde "$nombre ${v:-instalado}"
  fi
}

opcional() { # nombre comando
  command -v "$2" >/dev/null 2>&1 && verde "$1" || ambar "$1 no está instalado (recomendado)"
}

# Lo que el proyecto necesita lo dice su perfil ("requisitos"). Sin perfil, no se sabe: solo las herramientas base.
SEP=$'\x1f'
PERFIL=sin; REQUISITOS=""
if [ -f equipo/perfil.json ]; then
  if REQUISITOS=$(python3 scripts/perfil.py --requisitos 2>/dev/null); then PERFIL=bien; else PERFIL=mal; REQUISITOS=""; fi
fi
DEL_PROYECTO=""
while IFS=$SEP read -r cmd _; do
  [ -n "$cmd" ] && DEL_PROYECTO="$DEL_PROYECTO $cmd"
done <<< "$REQUISITOS"
# Docker hace falta si el proyecto tiene entorno local (make up) o si su perfil lo pide.
COMPOSE=""
for f in compose.yaml compose.yml docker-compose.yaml docker-compose.yml; do [ -f "$f" ] && COMPOSE=$f && break; done
CON_DOCKER=$COMPOSE
case " $DEL_PROYECTO " in *" docker "*) CON_DOCKER=perfil ;; esac

echo "Herramientas base"
requerido "Git"            git     "2.30" "git --version"
requerido "Python"         python3 "3.11" "python3 --version"
requerido "uv"             uv      ""     "uv --version"
requerido "jq"             jq      ""     "jq --version"
requerido "make"           make    ""     "make --version"
requerido "Spec Kit (specify)" specify "" "specify --version"
[ -n "$COMPOSE" ] && [ "$CON_DOCKER" != perfil ] && requerido "Docker" docker "24.0" "docker --version"
opcional "gitleaks"               gitleaks

case $PERFIL in
  sin)
    echo "Herramientas del proyecto"
    ambar "El proyecto no tiene perfil: no se sabe qué tecnología usa ni qué herramientas necesita (en el chat: /bowser-profile)"
    ;;
  bien)
    echo "Herramientas del proyecto (según su perfil)"
    if [ -z "$REQUISITOS" ]; then
      ambar "El perfil no declara requisitos: no se revisan las herramientas de su tecnología (en el chat: /bowser-profile)"
    fi
    while IFS=$SEP read -r cmd nombre minima vcmd opc instalar; do
      [ -n "$cmd" ] && requerido "$nombre" "$cmd" "$minima" "$vcmd" "$opc" "$instalar"
    done <<< "$REQUISITOS"
    ;;
  mal)
    echo "Herramientas del proyecto"
    rojo "El perfil tiene errores: no se sabe qué necesita el proyecto (detalle: make profile)"
    ;;
esac

echo "Agente de código (al menos uno)"
AGENTES=0
for a in claude codex opencode; do
  ruta=$(command -v "$a" 2>/dev/null) || continue
  if [[ "$ruta" == /mnt/* ]]; then
    rojo "$a es la versión de Windows ($ruta): instálalo dentro de Ubuntu (ver $DOCS/windows-wsl.md)"
  else
    verde "$a ($ruta)"; AGENTES=$((AGENTES+1))
  fi
done
[ "$AGENTES" -eq 0 ] && rojo "No hay ningún agente instalado en este sistema (Claude Code, Codex u OpenCode)"

# En WSL, Ubuntu también ve los programas de Windows. Si una herramienta de desarrollo resuelve a /mnt/...,
# se está usando la de Windows por error.
if grep -qi microsoft /proc/version 2>/dev/null; then
  echo "WSL"
  DE_WINDOWS=""
  for c in git python3 make jq uv specify ${CON_DOCKER:+docker} $DEL_PROYECTO; do
    ruta=$(command -v "$c" 2>/dev/null) || continue
    [[ "$ruta" == /mnt/* ]] && DE_WINDOWS="$DE_WINDOWS $c"
  done
  if [ -n "$DE_WINDOWS" ]; then
    rojo "Se están usando versiones de Windows de:$DE_WINDOWS. Instálalas en Ubuntu (ver $DOCS/windows-wsl.md)"
  else
    verde "Las herramientas de desarrollo son las de Ubuntu"
  fi
  case "$PWD" in
    /mnt/*) ambar "El proyecto está en el disco de Windows ($PWD). Muévelo a Ubuntu (ej. ~/proyectos/) para evitar lentitud y problemas de permisos" ;;
    *) verde "El proyecto está dentro de Ubuntu" ;;
  esac
fi

echo "Proyecto"
if [ -n "$CON_DOCKER" ]; then
  if docker info >/dev/null 2>&1; then verde "Docker está corriendo"; else rojo "Docker no está corriendo"; fi
fi
if [ "$(git config core.hooksPath 2>/dev/null)" != ".githooks" ]; then rojo "Hooks de git inactivos (ejecuta: make install-hooks)"
elif [ ! -x .githooks/pre-commit ] || [ ! -x .githooks/commit-msg ]; then rojo "Hooks de git sin permiso de ejecución: git los ignora (ejecuta: make install-hooks)"
else verde "Hooks de git activos"; fi
# .env solo hace falta si el proyecto declara alguna variable en .env.example.
if [ -f .env ]; then verde "Archivo .env existe"
elif grep -qE '^[[:space:]]*[A-Za-z_][A-Za-z0-9_]*=' .env.example 2>/dev/null; then rojo "Falta .env (ejecuta: cp .env.example .env y completa valores)"
fi
if python3 scripts/sincronizar.py --verificar >/dev/null 2>&1; then verde "Configuración de agentes al día"; else rojo "Configuración de agentes desactualizada (ejecuta: make sync)"; fi
RUTA_KIT=$(bash scripts/ruta-kit.sh)
if [ -f .kit-manifest.json ]; then
  if [ ! -f "$RUTA_KIT/scripts/instalar_kit.py" ]; then rojo "Submódulo $RUTA_KIT/ sin inicializar (ejecuta: git submodule update --init)"
  elif python3 "$RUTA_KIT/scripts/instalar_kit.py" --verificar >/dev/null 2>&1; then verde "Kit instalado y al día ($(jq -r .version .kit-manifest.json 2>/dev/null), en $RUTA_KIT/)"
  else rojo "Kit desactualizado o modificado (detalle: make check-kit)"; fi
elif [ -f "$RUTA_KIT/scripts/instalar_kit.py" ]; then
  rojo "Submódulo $RUTA_KIT/ agregado pero el kit no está instalado (ejecuta: make -f $RUTA_KIT/Makefile install-kit)"
fi
[ -d .specify/templates ] && verde "Spec Kit inicializado" || rojo "Spec Kit no inicializado (ejecuta: specify init --here --integration <herramienta>)"
command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1 && verde "GitHub CLI autenticado" || ambar "GitHub CLI no autenticado (necesario para que los agentes abran PRs: gh auth login)"

echo
echo "Resultado: $OK correctos, $AVISO avisos, $FALTA problemas."
[ "$FALTA" -eq 0 ] && echo "Todo listo para trabajar." || echo "Corrige los problemas marcados con ✗ (ver $DOCS/problemas-comunes.md)."
exit $([ "$FALTA" -eq 0 ] && echo 0 || echo 1)
