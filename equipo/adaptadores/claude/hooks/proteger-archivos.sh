#!/usr/bin/env bash
# Hook PreToolUse: impide que los agentes editen archivos sensibles.
# Salir con código 2 bloquea la acción y le muestra el motivo a Claude.

FILE=$(jq -r '.tool_input.file_path // empty')
[ -z "$FILE" ] && exit 0
RAIZ="${CLAUDE_PROJECT_DIR:-$PWD}"
REL="${FILE#"$RAIZ"/}"

# 1. Secretos (.env.example sí se puede editar: documenta variables sin valores reales)
case "$FILE" in
  *.env.example) ;;
  *.env|*/.env.*|*/secrets/*|*.pem|*.key)
    echo "Bloqueado: '$FILE' puede contener secretos. Usa .env.example para documentar variables." >&2
    exit 2
    ;;
esac

# 2. La constitución solo la cambia un humano: ni sus dos fuentes (principios y reglas del proyecto) ni el
#    archivo que se genera con ellas
case "$FILE" in
  */equipo/principios.md|*/equipo/reglas.md)
    echo "Bloqueado: la constitución (principios y reglas del proyecto) solo se modifica con aprobación de la dirección técnica. Propón el texto para que una persona lo aplique." >&2
    exit 2
    ;;
  */.specify/memory/constitution.md)
    echo "Bloqueado: la constitución se genera con equipo/principios.md y equipo/reglas.md, que solo se modifican con aprobación de la dirección técnica." >&2
    exit 2
    ;;
esac

# 3. El submódulo del kit y los archivos que vienen de él no se editan en el proyecto
RUTA_KIT=$(cd "$RAIZ" && bash scripts/ruta-kit.sh 2>/dev/null || echo .bowser-spec-kit-ai)
case "$REL" in
  "$RUTA_KIT"/*)
    echo "Bloqueado: '$RUTA_KIT/' es el submódulo del kit compartido. Los cambios se hacen en el repositorio del kit." >&2
    exit 2
    ;;
esac
if [ -f "$RAIZ/.kit-manifest.json" ] && jq -e --arg p "$REL" '.archivos | has($p)' "$RAIZ/.kit-manifest.json" >/dev/null 2>&1; then
  echo "Bloqueado: '$REL' viene del kit compartido y se reemplaza al actualizarlo. Propón el cambio en el repositorio del kit, o si es propio de este proyecto, un humano puede agregarlo a \"kit.excluir\" en equipo/config.json." >&2
  exit 2
fi

# 4. Los archivos que el perfil del proyecto declara inmutables (ej. migraciones) no se editan una vez
#    versionados en git: se crea uno nuevo. Un archivo que todavía no existe no puede estar versionado.
if [ -f "$FILE" ] && [ -f "$RAIZ/scripts/verbos.py" ] && command -v python3 >/dev/null 2>&1; then
  MOTIVO=$(cd "$RAIZ" && python3 scripts/verbos.py --inmutable "$REL" 2>&1 >/dev/null)
  if [ $? -eq 1 ] && [ -n "$MOTIVO" ]; then
    echo "Bloqueado: $MOTIVO" >&2
    exit 2
  fi
fi

# 5. costos.json lo escribe solo `make costs` (los agentes no lo editan a mano)
case "$REL" in
  specs/*costos.json)
    echo "Bloqueado: '$REL' lo genera 'make costs'; no se edita a mano." >&2
    exit 2
    ;;
esac

exit 0
