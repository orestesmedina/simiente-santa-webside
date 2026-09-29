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

# 2. La constitución solo la cambia un humano
case "$FILE" in
  */.specify/memory/constitution.md)
    echo "Bloqueado: la constitución solo se modifica con aprobación de la dirección técnica." >&2
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

# 4. Migraciones ya versionadas en git no se editan (se crea una nueva)
if [[ "$FILE" == */backend/migrations/*.sql ]] && [ -f "$FILE" ]; then
  if git -C "$RAIZ" ls-files --error-unmatch "$FILE" >/dev/null 2>&1; then
    echo "Bloqueado: '$FILE' ya está versionada. Crea una migración nueva en lugar de editarla." >&2
    exit 2
  fi
fi

exit 0
