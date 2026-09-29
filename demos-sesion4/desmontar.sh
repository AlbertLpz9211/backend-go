#!/usr/bin/env bash
# desmontar.sh — borra SOLO el clúster kind "s04-rest".
# No toca el clúster "backend" de la sesión 2 ni las imágenes de Docker.
set -euo pipefail
CLUSTER="s04-rest"
if kind get clusters 2>/dev/null | grep -qx "$CLUSTER"; then
  kind delete cluster --name "$CLUSTER"
else
  echo "El clúster $CLUSTER no existe; nada que borrar."
fi
echo "Clústeres que quedan:"; kind get clusters
