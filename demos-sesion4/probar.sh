#!/usr/bin/env bash
# probar.sh — prueba con curl los pasos 1 a 5 contra la API local.
#
# Uso:
#   ./probar.sh        # corre TODAS las pruebas (pensado para paso5 o api/)
#   ./probar.sh 3      # solo las pruebas de los pasos 1..3
#   BASE=http://localhost:9090 ./probar.sh
#
# Antes, en otra terminal:   go run ./pasoN-xxx
set -u
BASE="${BASE:-http://localhost:8080}"
HASTA="${1:-5}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# pedir METODO RUTA [opciones extra de curl...]
# Imprime: método, ruta, código, cabeceras interesantes y el cuerpo.
pedir() {
  local metodo="$1" ruta="$2"; shift 2
  local codigo
  codigo=$(curl -s -X "$metodo" -D "$TMP/cab" -o "$TMP/cuerpo" -w '%{http_code}' "$@" "$BASE$ruta")
  printf '%-6s %-24s -> %s\n' "$metodo" "$ruta" "$codigo"
  # Solo las cabeceras que importan para la clase.
  grep -iE '^(location|allow|x-pod):' "$TMP/cab" | tr -d '\r' | sed 's/^/         /'
  if [ -s "$TMP/cuerpo" ]; then
    sed 's/^/         /' "$TMP/cuerpo"; echo
  else
    echo "         (sin cuerpo)"
  fi
}

JSON=(-H 'Content-Type: application/json')

echo "### Paso 1 — salud"
pedir GET /api/v1/salud

if [ "$HASTA" -ge 2 ]; then
  echo "### Paso 2 — listar"
  pedir GET /api/v1/libros
fi

if [ "$HASTA" -ge 3 ]; then
  echo "### Paso 3 — obtener"
  pedir GET /api/v1/libros/1
  pedir GET /api/v1/libros/abc
  pedir GET /api/v1/libros/99
fi

if [ "$HASTA" -ge 4 ]; then
  echo "### Paso 4 — crear"
  pedir POST /api/v1/libros "${JSON[@]}" -d '{"titulo":"Aura","autor":"Carlos Fuentes","anio":1962}'
  pedir POST /api/v1/libros "${JSON[@]}" -d '{"titulo":"Aura",'                          # JSON roto
  pedir POST /api/v1/libros "${JSON[@]}" -d '{"titulo":"Aura","editorial":"Era"}'         # campo desconocido
  pedir POST /api/v1/libros "${JSON[@]}" -d '{"titulo":"","autor":"Anónimo"}'             # título vacío
  pedir POST /api/v1/libros -H 'Content-Type: text/plain' -d 'hola'                       # tipo incorrecto
  pedir POST /api/v1/libros -H 'Content-Type: application/json; charset=utf-8' -d '{"titulo":"Balún Canán","autor":"Rosario Castellanos","anio":1957}'
fi

if [ "$HASTA" -ge 5 ]; then
  echo "### Paso 5 — eliminar y 405"
  pedir DELETE /api/v1/libros/1
  pedir DELETE /api/v1/libros/1
  pedir DELETE /api/v1/libros/abc
  pedir PUT /api/v1/libros/2 "${JSON[@]}" -d '{"titulo":"x"}'
  pedir GET /api/v1/libros
fi
