#!/usr/bin/env bash
# probar-cluster.sh — ¿qué pasa con el estado EN MEMORIA cuando hay 3 réplicas?
#
# 1) Crea un libro con POST (lo guarda UN solo pod: el que atendió).
#    El título lleva la hora, para distinguir un libro de otro.
# 2) Pide ese libro 10 veces con GET. Cada curl abre una conexión NUEVA y el
#    Service la reparte al azar entre los pods.
# 3) Imprime por línea:  código  X-Pod  [aviso]
#
# Resultados posibles de cada GET:
#   200            atendió el pod que guardó el libro.
#   404            atendió otro pod: ese libro no existe en SU memoria.
#   200 OTRO LIBRO atendió otro pod que, por su cuenta, ya había usado ese
#                  mismo id para un libro distinto (cada pod cuenta ids solo).
#
# Moraleja: REST pide servidores SIN ESTADO; el estado va en una base de datos
# compartida (unidad 3: PostgreSQL).
#
# Uso:  ./probar-cluster.sh        (BASE y VECES son opcionales)
set -u
BASE="${BASE:-http://localhost:30080}"
VECES="${VECES:-10}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# cabecera NOMBRE: valor de una cabecera de la última respuesta.
cabecera() { grep -i "^$1:" "$TMP/cab" | head -1 | cut -d' ' -f2- | tr -d '\r'; }
# titulo: el campo "titulo" del último cuerpo JSON.
titulo() { sed -n 's/.*"titulo":"\([^"]*\)".*/\1/p' "$TMP/cuerpo"; }

mio="El laberinto de la soledad ($(date +%H:%M:%S) #$$)"
codigo=$(curl -s --max-time 5 -D "$TMP/cab" -o "$TMP/cuerpo" -w '%{http_code}' \
  -X POST "$BASE/api/v1/libros" \
  -H 'Content-Type: application/json' \
  -d "{\"titulo\":\"$mio\",\"autor\":\"Octavio Paz\",\"anio\":1950}")
ruta=$(cabecera Location)
echo "POST -> $codigo  guardado en: $(cabecera X-Pod)  Location: $ruta"
echo "      $(cat "$TMP/cuerpo")"
if [ "$codigo" != "201" ] || [ -z "$ruta" ]; then
  echo "El POST falló; ¿está desplegada la API?  kubectl get pods" >&2
  exit 1
fi

echo
echo "GET $ruta  x$VECES   (código  X-Pod)"
ok=0; no=0; otro_libro=0; otro=0
for _ in $(seq "$VECES"); do
  # Un proceso curl por petición = una conexión TCP nueva cada vez.
  # --max-time 5: si un pod está muriendo, no esperamos para siempre.
  : > "$TMP/cab"; : > "$TMP/cuerpo"
  c=$(curl -s --max-time 5 -D "$TMP/cab" -o "$TMP/cuerpo" -w '%{http_code}' "$BASE$ruta")
  aviso=""
  case "$c" in
    200) if [ "$(titulo)" = "$mio" ]; then ok=$((ok+1))
         else otro_libro=$((otro_libro+1)); aviso="  <- OTRO LIBRO: $(titulo)"; fi ;;
    404) no=$((no+1)) ;;
    *)   otro=$((otro+1)) ;;   # 000 = sin respuesta
  esac
  echo "$c  $(cabecera X-Pod)$aviso"
done
echo
echo "Resumen: $ok con 200, $no con 404, $otro_libro con 200 pero OTRO libro, $otro sin respuesta (de $VECES)"
