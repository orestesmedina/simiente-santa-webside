#!/usr/bin/env bash
# Verifica que el entorno de desarrollo tenga todo lo necesario. Uso: make doctor
set -uo pipefail

OK=0; FALTA=0; AVISO=0
verde() { printf "  \033[32m✓\033[0m %s\n" "$1"; OK=$((OK+1)); }
rojo()  { printf "  \033[31m✗\033[0m %s\n" "$1"; FALTA=$((FALTA+1)); }
ambar() { printf "  \033[33m!\033[0m %s\n" "$1"; AVISO=$((AVISO+1)); }

version_minima() { # $1 actual, $2 mínima
  [ "$(printf '%s\n%s\n' "$2" "$1" | sort -V | head -n1)" = "$2" ]
}

requerido() { # nombre comando versión_mínima comando_versión
  local nombre=$1 cmd=$2 minima=$3 vcmd=$4
  if ! command -v "$cmd" >/dev/null 2>&1; then rojo "$nombre no está instalado"; return; fi
  local v; v=$(eval "$vcmd" 2>/dev/null | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n1)
  if [ -n "$minima" ] && [ -n "$v" ] && ! version_minima "$v" "$minima"; then
    rojo "$nombre $v (se requiere $minima o superior)"
  else
    verde "$nombre ${v:-instalado}"
  fi
}

opcional() { # nombre comando
  command -v "$2" >/dev/null 2>&1 && verde "$1" || ambar "$1 no está instalado (recomendado)"
}

echo "Herramientas base"
requerido "Git"            git     "2.30" "git --version"
requerido "Docker"         docker  "24.0" "docker --version"
requerido "Go"             go      "1.23" "go version"
requerido "Node.js"        node    "22.0" "node --version"
requerido "Python"         python3 "3.11" "python3 --version"
requerido "uv"             uv      ""     "uv --version"
requerido "jq"             jq      ""     "jq --version"
requerido "make"           make    ""     "make --version"
requerido "Spec Kit (specify)" specify "" "specify --version"

echo "Herramientas de calidad y seguridad"
opcional "golangci-lint"          golangci-lint
opcional "govulncheck"            govulncheck
opcional "golang-migrate"         migrate
opcional "gitleaks"               gitleaks

echo "Agente de código (al menos uno)"
AGENTES=0
for a in claude codex opencode; do
  if command -v "$a" >/dev/null 2>&1; then verde "$a"; AGENTES=$((AGENTES+1)); fi
done
[ "$AGENTES" -eq 0 ] && rojo "No hay ningún agente instalado (Claude Code, Codex u OpenCode)"

echo "Proyecto"
if docker info >/dev/null 2>&1; then verde "Docker está corriendo"; else rojo "Docker no está corriendo"; fi
if [ "$(git config core.hooksPath 2>/dev/null)" = ".githooks" ]; then verde "Hooks de git activos"; else rojo "Hooks de git inactivos (ejecuta: make instalar-hooks)"; fi
[ -f .env ] && verde "Archivo .env existe" || rojo "Falta .env (ejecuta: cp .env.example .env y completa valores)"
if python3 scripts/sincronizar.py --verificar >/dev/null 2>&1; then verde "Configuración de agentes al día"; else rojo "Configuración de agentes desactualizada (ejecuta: make sincronizar)"; fi
RUTA_KIT=$(bash scripts/ruta-kit.sh)
if [ -f .kit-manifest.json ]; then
  if [ ! -f "$RUTA_KIT/scripts/instalar_kit.py" ]; then rojo "Submódulo $RUTA_KIT/ sin inicializar (ejecuta: git submodule update --init)"
  elif python3 "$RUTA_KIT/scripts/instalar_kit.py" --verificar >/dev/null 2>&1; then verde "Kit instalado y al día ($(jq -r .version .kit-manifest.json 2>/dev/null), en $RUTA_KIT/)"
  else rojo "Kit desactualizado o modificado (detalle: make verificar-kit)"; fi
elif [ -f "$RUTA_KIT/scripts/instalar_kit.py" ]; then
  rojo "Submódulo $RUTA_KIT/ agregado pero el kit no está instalado (ejecuta: make -f $RUTA_KIT/Makefile instalar-kit)"
fi
[ -d .specify/templates ] && verde "Spec Kit inicializado" || rojo "Spec Kit no inicializado (ejecuta: specify init --here --integration <herramienta>)"
command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1 && verde "GitHub CLI autenticado" || ambar "GitHub CLI no autenticado (necesario para que los agentes abran PRs: gh auth login)"

echo
echo "Resultado: $OK correctos, $AVISO avisos, $FALTA problemas."
[ "$FALTA" -eq 0 ] && echo "Todo listo para trabajar." || echo "Corrige los problemas marcados con ✗ (ver docs/GUIA-INICIO.md, sección 'Problemas comunes')."
exit $([ "$FALTA" -eq 0 ] && echo 0 || echo 1)
