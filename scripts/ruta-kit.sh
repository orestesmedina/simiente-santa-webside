#!/usr/bin/env bash
# Imprime la ruta (relativa a la raíz del proyecto) del submódulo del kit.
# Orden: 1) la guardada en .kit-manifest.json por `make instalar-kit`
#        2) un submódulo de .gitmodules que contenga scripts/instalar_kit.py
#        3) el nombre por defecto
# Uso: RUTA_KIT=$(bash scripts/ruta-kit.sh)      (ejecutar desde la raíz del proyecto)
POR_DEFECTO=".bowser-spec-kit-ai"

if [ -f .kit-manifest.json ]; then
  ruta=$(python3 -c 'import json;print(json.load(open(".kit-manifest.json")).get("ruta_kit",""))' 2>/dev/null)
  [ -n "$ruta" ] && { echo "$ruta"; exit 0; }
fi

if [ -f .gitmodules ]; then
  for ruta in $(git config -f .gitmodules --get-regexp '^submodule\..*\.path$' 2>/dev/null | awk '{print $2}'); do
    if [ -f "$ruta/scripts/instalar_kit.py" ] || [[ "$ruta" == *bowser-spec-kit-ai* ]]; then
      echo "$ruta"; exit 0
    fi
  done
fi

echo "$POR_DEFECTO"
