#!/usr/bin/env bash
# montar_laboratorio.sh — deja el laboratorio de la sesión 4 listo DESDE CERO.
#
# Qué hace (en orden), y se puede correr las veces que quieras (idempotente):
#   1. Comprueba herramientas (go, docker, kind, kubectl, curl) y que Docker responda.
#   2. Calidad del código: go vet, gofmt -l, go test -race.
#   3. Pre-descarga imágenes base (golang, distroless, nodo de kind).
#   4. Crea el clúster kind "s04-rest" si no existe (NO toca otros clústeres).
#   5. Construye la imagen api-libros:v1 y la carga en el clúster (kind load).
#   6. kubectl apply, espera el rollout y hace una prueba de humo.
#
# Es el script del DOCENTE (lo corre antes de la clase). Un alumno también puede
# usarlo si quiere reproducir el bloque de Kubernetes en su máquina.
#
# Sistemas:
#   macOS y Linux:  ./montar_laboratorio.sh
#   Windows:        desde Git Bash (NO PowerShell ni CMD), con Docker Desktop
#                   encendido (backend WSL2):  ./montar_laboratorio.sh
#
# Uso:  ./montar_laboratorio.sh
set -euo pipefail
cd "$(dirname "$0")"

CLUSTER="s04-rest"
CTX="kind-$CLUSTER"
IMAGEN="api-libros:v1"
IMG_GO="golang:1.26-alpine"
IMG_DISTROLESS="gcr.io/distroless/static-debian12:nonroot"
# La imagen de nodo se lee de k8s/kind-config.yaml para no repetirla.
IMG_NODO="$(awk '/image: kindest\/node/{print $2}' k8s/kind-config.yaml)"

k() { kubectl --context "$CTX" "$@"; }   # siempre contra NUESTRO clúster
paso() { printf '\n==> %s\n' "$*"; }
cronometro() { printf '    (%ss)\n' "$(( SECONDS - T0 ))"; }

paso "1. Herramientas"
for h in go docker kind kubectl curl; do
  command -v "$h" >/dev/null || { echo "Falta '$h' en el PATH" >&2; exit 1; }
done
if ! docker info >/dev/null 2>&1; then
  echo "Docker no responde. Abre Docker Desktop y espera a que diga 'Engine running':" >&2
  echo "  macOS:   open -a Docker" >&2
  echo "  Windows: menú Inicio > Docker Desktop" >&2
  exit 1
fi
echo "    $(go version)"
echo "    docker $(docker version --format '{{.Server.Version}}')"
echo "    $(kind version)"
echo "    kubectl $(kubectl version --client -o json | awk -F'"' '/gitVersion/{print $4; exit}')"

paso "2. Calidad del código"
T0=$SECONDS
go vet ./...
sin_formato="$(gofmt -l .)"
if [ -n "$sin_formato" ]; then echo "gofmt: archivos sin formato:"; echo "$sin_formato"; exit 1; fi
# -race necesita cgo en Windows y Linux (en Windows, además, gcc). Se prueba
# con una compilación vacía: si no se puede, se corren las pruebas sin -race.
if go test -race -run '^$' ./api >/dev/null 2>&1; then
  go test -race ./...
else
  echo "    aviso: este sistema no admite -race (falta cgo/gcc); pruebas sin -race"
  go test ./...
fi
cronometro

paso "3. Imágenes base"
T0=$SECONDS
for img in "$IMG_GO" "$IMG_DISTROLESS" "$IMG_NODO"; do
  if docker image inspect "$img" >/dev/null 2>&1; then
    echo "    ya está: $img"
  else
    echo "    descargando: $img"
    docker pull -q "$img" >/dev/null
  fi
done
cronometro

paso "4. Clúster kind '$CLUSTER'"
T0=$SECONDS
if kind get clusters 2>/dev/null | grep -qx "$CLUSTER"; then
  echo "    ya existe"
  # Si Docker Desktop se reinició, el contenedor-nodo puede estar detenido.
  if [ "$(docker inspect -f '{{.State.Running}}' "$CLUSTER-control-plane")" != "true" ]; then
    echo "    el nodo estaba detenido; arrancándolo"
    docker start "$CLUSTER-control-plane" >/dev/null
  fi
else
  # ¿Alguien escucha ya en 30080? curl sale con 7 si NADIE responde (libre).
  # Se usa curl y no lsof porque lsof no existe en Windows.
  rc=0; curl -s -o /dev/null --max-time 2 http://localhost:30080/ || rc=$?
  if [ "$rc" -ne 7 ]; then
    echo "El puerto 30080 ya está ocupado (curl respondió con código $rc)." >&2
    echo "Cierra lo que lo use, o bórralo con ./desmontar.sh si es un clúster viejo." >&2
    exit 1
  fi
  kind create cluster --config k8s/kind-config.yaml
fi
# El nodo tarda unos segundos en pasar a Ready (se está instalando la red).
k wait --for=condition=Ready node --all --timeout=120s
cronometro

paso "5. Imagen $IMAGEN"
T0=$SECONDS
docker build -q -f api/Dockerfile -t "$IMAGEN" . >/dev/null
docker images "$IMAGEN"
kind load docker-image "$IMAGEN" --name "$CLUSTER"
cronometro

paso "6. Despliegue"
T0=$SECONDS
existia=no
k get deployment api-libros >/dev/null 2>&1 && existia=si
k apply -f k8s/api.yaml
if [ "$existia" = si ]; then
  # Misma etiqueta (v1) con código nuevo: hay que recrear los pods para que
  # tomen la imagen recién cargada. Además así arrancan con memoria limpia.
  k rollout restart deployment/api-libros
fi
k rollout status deployment/api-libros --timeout=120s
k get pods -l app=api-libros -o wide
cronometro

paso "7. Prueba de humo: curl localhost:30080/api/v1/salud"
for _ in $(seq 30); do
  if curl -sf localhost:30080/api/v1/salud >/dev/null; then break; fi
  sleep 1
done
curl -s -i localhost:30080/api/v1/salud | tr -d '\r' | grep -E '^HTTP|^X-Pod|estado'

echo
echo "Laboratorio listo. Siguiente:  ./probar-cluster.sh"
