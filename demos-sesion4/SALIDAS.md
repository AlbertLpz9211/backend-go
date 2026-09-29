# SALIDAS — bitácora de verificación, Sesión 4 (martes 29 sep 2026)

Tema 2.1 REST + `net/http`, más el bloque de Kubernetes en Docker (kind).
Todo lo que aparece aquí se ejecutó el 29-sep-2026 (10:30–10:50, hora local) en la Mac del docente.
Las salidas están copiadas de la terminal y recortadas solo donde se indica `...`.

## 0. Versiones exactas

| Pieza | Versión |
| --- | --- |
| Go (Mac) | `go version go1.26.5 darwin/arm64` |
| Go dentro de `golang:1.26-alpine` | `go version go1.26.8 linux/arm64` |
| Docker (cliente y daemon) | 29.6.2 |
| kind | `kind v0.33.0 go1.27.0 darwin/arm64` |
| kubectl (cliente) | v1.37.0 (Kustomize v5.8.1) |
| Kubernetes del nodo | Server Version v1.37.0 |
| Imagen de nodo | `kindest/node:v1.37.0@sha256:a1ed56cfb0e7b93589bdf97c8cd566405a265939e3620fc4f5de89adff580ae5` |
| Runtime del nodo | containerd://2.3.4, Debian GNU/Linux 13 (trixie), kernel 6.12.76-linuxkit (arm64) |
| kube-proxy | modo `iptables` |

## 1. Estructura

```
demos-sesion4/
├── go.mod                 module demos-sesion4 / go 1.26 (solo biblioteca estándar)
├── .gitignore  .dockerignore
├── paso1-hola/main.go     GET /api/v1/salud
├── paso2-listar/main.go   + Libro, almacen (RWMutex), escribirJSON, GET /api/v1/libros
├── paso3-obtener/main.go  + errorAPI, errNoEncontrado, GET /api/v1/libros/{id}, semilla de 2 libros
├── paso4-crear/main.go    + POST (415 / 400 / 422 / 201 + Location, MaxBytesReader)
├── paso5-eliminar/main.go + DELETE (204 / 404 / 400); 405 automático del mux
├── api/
│   ├── main.go            paso5 + X-Pod + registro slog JSON + http.Server con timeouts + apagado ordenado
│   ├── main_test.go       prueba de tabla httptest (14 casos) + prueba concurrente (100 POST)
│   └── Dockerfile         golang:1.26-alpine -> distroless/static-debian12:nonroot, USER 65532:65532
├── k8s/kind-config.yaml   clúster s04-rest, extraPortMappings 30080 -> 127.0.0.1:30080
├── k8s/api.yaml           Deployment api-libros (3 réplicas) + Service NodePort 30080
├── probar.sh              curls de los pasos 1-5 (./probar.sh N limita a los pasos 1..N)
├── probar-cluster.sh      POST + 10 GET contra localhost:30080 (código  X-Pod)
├── montar_laboratorio.sh  todo desde cero, idempotente
└── desmontar.sh           borra SOLO el clúster s04-rest
```

Cada `pasoN` es un `main.go` completo que compila y corre solo; el comentario de cabecera dice qué agrega.
Nombres idénticos a la KB: `Libro`, `errorAPI`, `almacen`, `escribirJSON`, `errNoEncontrado`, rutas
`/api/v1/libros`, códigos `id_invalido`, `no_encontrado`, `json_invalido`, `validacion`, `tipo_no_soportado`.
Semilla (desde el paso 3): `1 Pedro Páramo / Juan Rulfo / 1955` y `2 Como agua para chocolate / Laura Esquivel / 1989`.

## 2. Calidad del código

```
$ go vet ./...
(sin salida)
$ gofmt -l .
(sin salida)
$ go test -race -count=1 ./...
ok  	demos-sesion4/api	1.508s
?   	demos-sesion4/paso1-hola	[no test files]
...
?   	demos-sesion4/paso5-eliminar	[no test files]
```

`go test -race -v ./api` (recorte):

```
--- PASS: TestCodigos (0.00s)
    --- PASS: TestCodigos/salud (0.00s)
    --- PASS: TestCodigos/listar (0.00s)
    --- PASS: TestCodigos/obtener_existente (0.00s)
    --- PASS: TestCodigos/obtener_id_no_numérico (0.00s)
    --- PASS: TestCodigos/obtener_inexistente (0.00s)
    --- PASS: TestCodigos/crear (0.00s)
    --- PASS: TestCodigos/crear_con_charset (0.00s)
    --- PASS: TestCodigos/crear_JSON_roto (0.00s)
    --- PASS: TestCodigos/crear_campo_desconocido (0.00s)
    --- PASS: TestCodigos/crear_título_vacío (0.00s)
    --- PASS: TestCodigos/crear_tipo_incorrecto (0.00s)
    --- PASS: TestCodigos/eliminar_existente (0.00s)
    --- PASS: TestCodigos/eliminar_inexistente (0.00s)
    --- PASS: TestCodigos/método_no_permitido (0.00s)
=== RUN   TestCrearConcurrente
--- PASS: TestCrearConcurrente (0.00s)
PASS
```

Cubre 200/201/204/400/404/405/415/422 y comprueba `X-Pod` en todas, `Location` en 201 y `Allow` en 405.

**Demo "quitar el mutex" (verificada en una copia aparte, no en el repo):** comentando `a.mu.Lock()` y
`defer a.mu.Unlock()` de `crear`/`eliminar` en `api/main.go`, `go test -race -run TestCrearConcurrente`
imprime varias veces `WARNING: DATA RACE` y termina en `FAIL`. Tiempo: menos de 1 s.

## 3. Pasos de live coding (local, puerto 8080)

Procedimiento usado para cada paso: `go build` del paso, arrancarlo en segundo plano, `./probar.sh N`,
matar con `lsof -ti:8080 | xargs kill`. En clase basta con `go run ./pasoN-xxx` en una terminal y
`./probar.sh N` en otra.

### Paso 1 — `go run ./paso1-hola`, `./probar.sh 1`
```
### Paso 1 — salud
GET    /api/v1/salud            -> 200
         {"estado":"ok"}
```
Log del servidor: `2026/09/29 10:32:40 escuchando en :8080`

### Paso 2 — `./probar.sh 2`
```
GET    /api/v1/libros           -> 200
         {"datos":[]}
```
**Bug del `null` verificado:** cambiando `out := make([]Libro, 0, len(a.libros))` por `var out []Libro`
(la versión comentada en `paso2-listar/main.go`) la respuesta es exactamente:
```
{"datos":null}
```

### Paso 3 — `./probar.sh 3`
```
GET    /api/v1/libros           -> 200
         {"datos":[{"id":1,"titulo":"Pedro Páramo","autor":"Juan Rulfo","anio":1955},{"id":2,"titulo":"Como agua para chocolate","autor":"Laura Esquivel","anio":1989}]}
GET    /api/v1/libros/1         -> 200
         {"id":1,"titulo":"Pedro Páramo","autor":"Juan Rulfo","anio":1955}
GET    /api/v1/libros/abc       -> 400
         {"codigo":"id_invalido","mensaje":"el id debe ser un entero"}
GET    /api/v1/libros/99        -> 404
         {"codigo":"no_encontrado","mensaje":"el libro no existe"}
```

### Paso 4 — `./probar.sh 4`
```
POST   /api/v1/libros           -> 201
         Location: /api/v1/libros/3
         {"id":3,"titulo":"Aura","autor":"Carlos Fuentes","anio":1962}
POST   /api/v1/libros           -> 400          (cuerpo '{"titulo":"Aura",')
         {"codigo":"json_invalido","mensaje":"unexpected EOF"}
POST   /api/v1/libros           -> 400          (campo "editorial")
         {"codigo":"json_invalido","mensaje":"json: unknown field \"editorial\""}
POST   /api/v1/libros           -> 422          (titulo "")
         {"codigo":"validacion","mensaje":"titulo es obligatorio"}
POST   /api/v1/libros           -> 415          (Content-Type: text/plain)
         {"codigo":"tipo_no_soportado","mensaje":"se espera application/json"}
POST   /api/v1/libros           -> 201          (Content-Type: application/json; charset=utf-8)
         Location: /api/v1/libros/4
         {"id":4,"titulo":"Balún Canán","autor":"Rosario Castellanos","anio":1957}
```

### Paso 5 — `./probar.sh` (todo)
```
DELETE /api/v1/libros/1         -> 204
         (sin cuerpo)
DELETE /api/v1/libros/1         -> 404
         {"codigo":"no_encontrado","mensaje":"el libro no existe"}
DELETE /api/v1/libros/abc       -> 400
         {"codigo":"id_invalido","mensaje":"el id debe ser un entero"}
PUT    /api/v1/libros/2         -> 405
         Allow: DELETE, GET, HEAD
         Method Not Allowed
GET    /api/v1/libros           -> 200
         {"datos":[{"id":2,...},{"id":3,...},{"id":4,...}]}
```
El 405 del ServeMux **sí incluye `Allow`** (`DELETE, GET, HEAD`: HEAD aparece solo porque registrar
GET habilita HEAD). Respuesta completa con `curl -i -X PUT localhost:8080/api/v1/libros/2`:
```
HTTP/1.1 405 Method Not Allowed
Allow: DELETE, GET, HEAD
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
X-Pod: MacBook-Pro-de-Alberto.local
...
Method Not Allowed
```
(La línea `X-Pod` es de `api/`; en `paso5` no existe.) Una ruta que no existe (`/no/existe`) da
`404 page not found` en texto plano, también con `X-Pod`.

### `api/` local — `go run ./api`, `./probar.sh`
Los 16 resultados son idénticos al paso 5 (200, 200, 200, 400, 404, 201, 400, 400, 422, 415, 201, 204,
404, 400, 405, 200) y **las 16 respuestas traen `X-Pod`** (incluidas la 405 y la 204). Log slog, una línea
por petición:
```
{"time":"...","level":"INFO","msg":"escuchando","addr":":8080","pod":"MacBook-Pro-de-Alberto.local"}
{"time":"...","level":"INFO","msg":"peticion","metodo":"GET","ruta":"/api/v1/salud","estado":200,"duracion":"281.792µs"}
{"time":"...","level":"INFO","msg":"apagando"}          <- tras kill -TERM (apagado ordenado)
```

## 4. Imagen de contenedor

Comando exacto (el contexto es la RAÍZ del módulo, porque ahí está `go.mod`):
```
cd demos-sesion4
docker build -f api/Dockerfile -t api-libros:v1 .
```
`.dockerignore` deja pasar solo `go.mod` y `api/` (sin `*_test.go`).

| Medida | Valor |
| --- | --- |
| `docker pull golang:1.26-alpine` (primera vez) | 18.6 s |
| `docker pull gcr.io/distroless/static-debian12:nonroot` | 0.7 s |
| Imagen de nodo kind | ya estaba en la Mac (sesión 2, 1.3 GB); no se midió la descarga |
| `docker build` (imagen base ya descargada) | 6.1 s |
| `docker build --no-cache` | 5.9 s |
| Imagen `api-libros:v1` | **3.13 MB** CONTENT SIZE (3 130 522 B, `docker image inspect .Size`); 14.4 MB "DISK USAGE" en `docker images` |
| Capa del binario (`COPY /out/api /api`) | 5.78 MB (descomprimido) |
| Binario linux/arm64 con `-ldflags "-s -w"` | 5 767 330 B |
| Binario linux/arm64 sin `-s -w` | 8 348 000 B |
| Binario macOS (`go build ./api`) | 8 590 226 B |
| Imagen base distroless | 721 kB content / 6.18 MB disk |
| golang:1.26-alpine (solo la etapa de compilación) | 68.9 MB content / 357 MB disk |
| Usuario de la imagen | `65532:65532`, arch arm64 |

Prueba local del contenedor (`docker run -d --rm -p 8081:8080 api-libros:v1`):
```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Pod: f4a68b4299f6            <- hostname del contenedor = su id corto
{"estado":"ok"}
```
distroless no tiene shell: `docker run --rm --entrypoint sh api-libros:v1 -c id` falla con
`exec: "sh": executable file not found in $PATH`.

## 5. Clúster kind `s04-rest`

```
$ time kind create cluster --config k8s/kind-config.yaml
Creating cluster "s04-rest" ...
 ✓ Ensuring node image (kindest/node:v1.37.0) 🖼️
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-s04-rest"
... 13.091 total
```
Justo después el nodo seguía `NotReady` (16 s de edad); por eso `montar_laboratorio.sh` hace
`kubectl wait --for=condition=Ready node --all`.

```
$ time kind load docker-image api-libros:v1 --name s04-rest
Image: "api-libros:v1" with ID "sha256:c263e2..." not yet present on node "s04-rest-control-plane", loading...
... 1.269 total
$ docker exec s04-rest-control-plane crictl images | grep api-libros
docker.io/library/api-libros   v1   1f6eb41120c4b   3.11MB
$ kubectl apply -f k8s/api.yaml
deployment.apps/api-libros created
service/api-libros created
$ kubectl rollout status deployment/api-libros
... deployment "api-libros" successfully rolled out        (primer despliegue: ~11 s incluyendo que el nodo pasara a Ready)
```
`kind load` funcionó sin problemas con el almacén de imágenes containerd de Docker 29 (manifest list + attestation).

Tiempos de `rollout restart` de 3 réplicas: 4–8 s sin `preStop`; 18–19 s con `preStop sleep 5` (cada pod
viejo espera 5 s antes de apagarse y el rolling update va de uno en uno).

Memoria real de los contenedores `api` (`crictl stats` dentro del nodo): **2.8–4.9 MB** cada uno; límite
del manifiesto 32Mi. En los logs del pod aparecen las sondas: una línea `GET /api/v1/salud` cada 5 s
(readiness) y cada 10 s (liveness).

Estado final dejado para la clase:
```
$ kubectl get deploy,pods,svc
deployment.apps/api-libros   3/3     3            3
pod/api-libros-...   1/1   Running   (3 pods)
service/api-libros   NodePort   10.96.20.202   <none>   80:30080/TCP
$ docker ps -a
s04-rest-control-plane  kindest/node:v1.37.0  Up    127.0.0.1:30080->30080/tcp, 127.0.0.1:55346->6443/tcp
backend-control-plane   kindest/node:v1.37.0  Exited (137)      <- sin tocar, como estaba
```

## 6. `montar_laboratorio.sh`

Corrido 4 veces con el clúster ya existente (camino idempotente). Salida de la última (recortada):
```
==> 1. Herramientas
    go version go1.26.5 darwin/arm64
    docker 29.6.2
    kind v0.33.0 go1.27.0 darwin/arm64
    kubectl v1.37.0
==> 2. Calidad del código
ok  	demos-sesion4/api	(cached)
    (0s)
==> 3. Imágenes base
    ya está: golang:1.26-alpine
    ya está: gcr.io/distroless/static-debian12:nonroot
    ya está: kindest/node:v1.37.0@sha256:a1ed56cf...
==> 4. Clúster kind 's04-rest'
    ya existe
node/s04-rest-control-plane condition met
==> 5. Imagen api-libros:v1
api-libros:v1   8bb27f278d6a       14.4MB         3.13MB
Image: "api-libros:v1" with ID "sha256:8bb27f..." not yet present on node "s04-rest-control-plane", loading...
    (1s)
==> 6. Despliegue
deployment.apps/api-libros configured
service/api-libros unchanged
deployment.apps/api-libros restarted
deployment "api-libros" successfully rolled out
    (18s)
==> 7. Prueba de humo: curl localhost:30080/api/v1/salud
HTTP/1.1 200 OK
X-Pod: api-libros-59497bb5f9-vmgnq
{"estado":"ok"}
Laboratorio listo. Siguiente:  ./probar-cluster.sh
```
Duración total medida: 14 s, 22 s y 21 s (sin `preStop` / con `preStop`). Una corrida canalizada a `grep`
reportó 1 min 22 s; no se reprodujo en tres intentos posteriores (21–22 s) y no se identificó la causa.

**Camino "desde cero" (clúster inexistente):** los comandos del script (`kind create cluster --config`,
`kubectl wait node`, `docker build`, `kind load`, `kubectl apply`, `rollout status`) se ejecutaron uno a uno
con éxito (secciones 4 y 5), pero el script completo **no se pudo correr sobre un clúster borrado**:
el permiso para ejecutar `kind delete cluster` / `docker rmi` fue denegado en esta sesión, así que
`desmontar.sh` **no se ejecutó** (solo `bash -n`). Conviene que el docente corra una vez
`./desmontar.sh && ./montar_laboratorio.sh` antes de la clase.

## 7. `probar-cluster.sh` — estado en memoria con 3 réplicas

**El NodePort sí reparte**: cada invocación de `curl` abre una conexión TCP nueva y kube-proxy (iptables)
elige un pod al azar. 12 peticiones de prueba: 2 / 6 / 4 entre los tres pods. No hizo falta `Connection: close`
ni lanzar las peticiones desde dentro del clúster, siempre que sea **un proceso curl por petición**.

Corridas oficiales, justo después de `kubectl rollout restart` (pods con memoria limpia):
```
######## corrida 1
POST -> 201  guardado en: api-libros-5c47885b74-f9bnf  Location: /api/v1/libros/3
      {"id":3,"titulo":"El laberinto de la soledad (10:47:42)","autor":"Octavio Paz","anio":1950}

GET /api/v1/libros/3  x10   (código  X-Pod)
200  api-libros-5c47885b74-f9bnf
200  api-libros-5c47885b74-f9bnf
200  api-libros-5c47885b74-f9bnf
404  api-libros-5c47885b74-zkww5
200  api-libros-5c47885b74-f9bnf
404  api-libros-5c47885b74-s2ctl
404  api-libros-5c47885b74-zkww5
404  api-libros-5c47885b74-s2ctl
404  api-libros-5c47885b74-zkww5
404  api-libros-5c47885b74-zkww5

Resumen: 4 con 200, 6 con 404, 0 con 200 pero OTRO libro, 0 sin respuesta (de 10)
######## corrida 2   (guardado en f9bnf, /api/v1/libros/4)
Resumen: 3 con 200, 7 con 404, 0 con 200 pero OTRO libro, 0 sin respuesta (de 10)
######## corrida 3   (guardado en f9bnf, /api/v1/libros/5)
Resumen: 5 con 200, 5 con 404, 0 con 200 pero OTRO libro, 0 sin respuesta (de 10)
```

| Corrida | 200 | 404 | 200 con OTRO libro | sin respuesta |
| --- | --- | --- | --- | --- |
| 1 | 4 | 6 | 0 | 0 |
| 2 | 3 | 7 | 0 | 0 |
| 3 | 5 | 5 | 0 | 0 |

Series anteriores (versión previa del script, que no detectaba "otro libro"), empezando con pods limpios:
3/7, 4/6, 1/9 (200/404) y 4/6, 5/5, 10/0. El 10/0 de la segunda serie NO significa que el estado se
compartiera: los tres pods ya tenían cada uno su propio id 3 (colisión, ver abajo).

**Colisión de ids (efecto extra, verificado):** si se repite el script sin reiniciar, cada pod numera
por su cuenta, así que el id 3 puede existir en dos pods con libros distintos. Corridas 5 y 6 de la misma serie:
```
POST -> 201  guardado en: api-libros-5c47885b74-zkww5  Location: /api/v1/libros/3
200  api-libros-5c47885b74-f9bnf  <- OTRO LIBRO: El laberinto de la soledad (10:47:42)
Resumen: 2 con 200, 7 con 404, 1 con 200 pero OTRO libro, 0 sin respuesta (de 10)

POST -> 201  guardado en: api-libros-5c47885b74-zkww5  Location: /api/v1/libros/4
Resumen: 4 con 200, 3 con 404, 3 con 200 pero OTRO libro, 0 sin respuesta (de 10)
```
Por eso el título lleva hora y PID (`El laberinto de la soledad (HH:MM:SS #PID)`) y el script compara el
título devuelto. Para números "limpios" en clase: `kubectl rollout restart deployment/api-libros` antes.

## 8. Demo de autocuración (`kubectl delete pod`)

Desde pods limpios; se borra justo el pod que guardó el libro:
```
POST -> 201  guardado en: api-libros-6cff467c99-rh44t  Location: /api/v1/libros/3
Resumen: 4 con 200, 6 con 404, 0 con 200 pero OTRO libro, 0 sin respuesta (de 10)
$ kubectl get pods
api-libros-6cff467c99-k9pbc   1/1     Running   0          8s
api-libros-6cff467c99-n729n   1/1     Running   0          9s
api-libros-6cff467c99-rh44t   1/1     Running   0          15s
$ kubectl delete pod api-libros-6cff467c99-rh44t
pod "api-libros-6cff467c99-rh44t" deleted from default namespace
(delete: 6s)
$ kubectl get pods
api-libros-6cff467c99-k9pbc   1/1     Running   0          14s
api-libros-6cff467c99-n26h6   1/1     Running   0          6s      <- creado solo por el Deployment
api-libros-6cff467c99-n729n   1/1     Running   0          15s
$ GET /api/v1/libros/3 x10   (agrupado con sort | uniq -c)
   5 404  api-libros-6cff467c99-k9pbc
   2 404  api-libros-6cff467c99-n26h6
   3 404  api-libros-6cff467c99-n729n
$ curl -i localhost:30080/api/v1/libros   (hasta que respondió el pod nuevo)
X-Pod: api-libros-6cff467c99-n26h6
{"datos":[{"id":1,"titulo":"Pedro Páramo",...},{"id":2,"titulo":"Como agua para chocolate",...}]}
```
El libro se perdió para siempre (10/10 404) y el pod nuevo solo tiene la semilla.
`kubectl delete pod` tarda ~6 s (5 s de `preStop` + apagado); antes de añadir `preStop` tardaba ~1 s.
Log del pod borrado (`kubectl logs -f`): `{"level":"INFO","msg":"apagando"}` → el SIGTERM llega y se atiende.

## 9. Demo de escalar a 1 réplica

```
$ kubectl scale deployment/api-libros --replicas=1
deployment.apps/api-libros scaled
api-libros-6cff467c99-k9pbc   1/1     Running       0          22s
api-libros-6cff467c99-n26h6   1/1     Terminating   0          14s
api-libros-6cff467c99-n729n   1/1     Terminating   0          23s
(quedó 1 pod a los 7s)
$ ./probar-cluster.sh
POST -> 201  guardado en: api-libros-6cff467c99-k9pbc  Location: /api/v1/libros/3
200  api-libros-6cff467c99-k9pbc     (x10)
Resumen: 10 con 200, 0 con 404, 0 con 200 pero OTRO libro, 0 sin respuesta (de 10)
```
Se restauró con `kubectl scale --replicas=3` + `rollout restart` (o simplemente `./montar_laboratorio.sh`,
que reaplica `replicas: 3`).

## 10. Trampas encontradas (y cómo se resolvieron)

1. **El 415 de la KB rechaza `application/json; charset=utf-8`.** El código literal de la KB
   (`ct != "application/json"`) se compiló y probó: con ese Content-Type devuelve 415, con
   `application/json` a secas devuelve 201. En los pasos 4, 5 y `api/` se usa `mime.ParseMediaType`,
   que acepta ambos (verificado: 201 con charset). Único punto donde el código difiere de la KB.
2. **El orden de `GET /api/v1/libros` cambia entre peticiones** (se recorre un map). Verificado: en 6
   peticiones seguidas, 5 salieron `1,2` y 1 salió `2,1`. Se dejó como en la KB; avisar a los alumnos.
3. **El 405 (y el 404 de ruta inexistente) del ServeMux es texto plano**, no el `errorAPI` JSON:
   `Method Not Allowed` / `404 page not found`, `Content-Type: text/plain`. Sí trae `Allow: DELETE, GET, HEAD`.
4. **gofmt reescribe comentarios de cabecera con líneas sangradas** (las convierte en bloque de código y
   mete una línea `//` vacía). Se reescribieron como listas `//   - ...`.
5. **`kind create cluster` cambia el contexto actual de kubectl a `kind-s04-rest`.** Los scripts usan
   siempre `--context kind-s04-rest`. El contexto `kind-backend` sigue en el kubeconfig; el actual quedó en
   `kind-s04-rest`.
6. **El nodo recién creado está `NotReady`** unos segundos (se instala la CNI). Solución:
   `kubectl wait --for=condition=Ready node --all` antes de desplegar.
7. **Keep-alive fija el pod.** `curl URL URL URL ...` (una sola invocación con 10 URLs) reutiliza la
   conexión: 10/10 al mismo pod. Con `-H 'Connection: close'` o un `curl` por petición se reparte
   (medido: 2/4/4). `probar-cluster.sh` usa un proceso curl por petición. Lo mismo pasará con el navegador.
8. **Petición colgada al escalar/borrar pods (antes del arreglo).** Con la primera versión (el proceso moría
   al instante con SIGTERM), justo después de `kubectl scale --replicas=1` un GET devolvió `000` y el
   script quedó colgado más de un minuto: la conexión fue a un pod que ya había muerto pero seguía en las
   reglas del Service. Arreglo: (a) `api/main.go` atiende SIGTERM con `signal.NotifyContext` +
   `srv.Shutdown`; (b) `lifecycle.preStop.sleep.seconds: 5` en `k8s/api.yaml` (acción nativa de
   Kubernetes, no necesita shell, que distroless no tiene); (c) `--max-time 5` en los curl del script.
   Después del arreglo: 0 "sin respuesta" en todas las pruebas.
9. **Justo después de escalar a 1, los pods en `Terminating` pueden seguir atendiendo unos segundos.**
   En 2 de 4 pruebas inmediatas salieron 404 (4 y 5 de 10). Esperando a que `kubectl get pods` muestre
   un solo pod (~7 s) el resultado fue 10/10 200 las dos veces que se midió.
10. **Estado acumulado entre corridas** → colisión de ids ("200 pero OTRO libro"). Ver sección 7.
    Si se quiere el resultado didáctico limpio: `kubectl rollout restart deployment/api-libros`.
11. **Misma etiqueta `v1` con código nuevo**: `kubectl apply` no recrea los pods si el manifiesto no
    cambió. `montar_laboratorio.sh` hace `rollout restart` cuando el Deployment ya existía.
12. **`go mod init` escribe `go 1.26.5`** (la versión de la Mac). Se dejó `go 1.26` para que cualquier
    `golang:1.26-alpine` compile sin intentar descargar otra toolchain (la imagen trae 1.26.8).
13. **zsh no parte variables en palabras:** `K="kubectl --context kind-s04-rest"; $K get pods` falla con
    `command not found`. Usar una función: `k(){ kubectl --context kind-s04-rest "$@"; }` (o bash).
14. **Docker 29 muestra dos tamaños** (`DISK USAGE` 14.4 MB y `CONTENT SIZE` 3.13 MB). El que se compara
    con la imagen Node de la sesión 2 debe ser el mismo criterio en ambas; `docker image inspect -f '{{.Size}}'`
    da 3 130 522 B.
15. **Las sondas ensucian los logs**: una línea de `/api/v1/salud` cada 5 s y cada 10 s por pod. Para
    mirar solo tráfico real: `kubectl logs <pod> | grep -v salud`.
16. **No verificado por falta de permiso**: `desmontar.sh` y `montar_laboratorio.sh` sobre un clúster
    inexistente (ver sección 6). La rama "puerto 30080 ocupado" del script tampoco se ejercitó.
