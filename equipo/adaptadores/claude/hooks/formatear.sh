#!/usr/bin/env bash
# Hook PostToolUse: formatea automáticamente el archivo que Claude acaba de editar.
# Recibe por stdin el JSON del evento; nunca bloquea (siempre sale con 0).

FILE=$(jq -r '.tool_input.file_path // empty')
[ -z "$FILE" ] || [ ! -f "$FILE" ] && exit 0

case "$FILE" in
  *.go)
    command -v gofmt >/dev/null && gofmt -w "$FILE"
    ;;
  *.ts|*.tsx|*.js|*.jsx|*.css|*.json|*.md)
    if [ -x "$CLAUDE_PROJECT_DIR/frontend/node_modules/.bin/prettier" ]; then
      "$CLAUDE_PROJECT_DIR/frontend/node_modules/.bin/prettier" --write "$FILE" >/dev/null 2>&1
    fi
    ;;
esac

exit 0
