# Ensayo y conducción de los demos · Sesión 2 · 22 sep 2026

Acompaña al guion de clase. Aquí está **qué comando se teclea, en qué terminal
y qué debe salir**. Todo lo marcado ✅ se ejecutó y verificó en esta máquina
el 22 de septiembre de 2026 (macOS arm64, 10 núcleos, Go 1.26.5, Node v25.9.0,
Docker 29.6.2).

## Montaje de terminales (B0)

| Terminal | Título | Carpeta base |
|---|---|---|
| 1 | `1-GORUTINAS` | `demos-sesion2/` |
| 2 | `2-NODE` | `demos-sesion2/` |
| 3 | `3-GO` | `demos-sesion2/` |
| 4 | `4-CLUSTER` | `demos-sesion2/05-kind/` |

Puertos altos para no chocar con nada: **18090** (Go), **18092** (Node), **18094** (carrera).

Antes del primer alumno:

```bash
cd /Users/albertolopez/Repositorios/backend-go/demos-sesion2
lsof -ti:18090,18092,18094 | xargs -r kill -9   # limpia puertos de ensayos previos
go build ./...                                   # calienta la caché: la 1a vez tarda
```

---

## B1.1 · El gancho (terminal 3)

Las dos cifras que se proyectan salen de aquí. **Ya medidas**, se pueden
proyectar como texto sin ejecutar nada:

```
Node:  rapido (2.921614s)   <- espero
Go:    rapido (0.000587s)   <- no espero
```

---

## B3 · Gorutinas y canales (terminal 1) ✅ todo verificado

```bash
go run ./01-gorutinas/paso1-roto
```
```
main termino en 85.959µs
```
No imprime ninguna consulta. El número de microsegundos cambia cada corrida.

```bash
go run ./01-gorutinas/paso2-waitgroup
```
```
secuencial  : 50 consultas en 5.058392875s
concurrente : 50 consultas en 101.355917ms
```
Tarda 5 segundos en aparecer la primera línea. **Avisarlo** o se lee como cuelgue.

```bash
go run ./01-gorutinas/paso3-diezmil
```
```
10000 gorutinas en 113.846542ms
pico de gorutinas vivas : 10001
nucleos logicos usables : 10 (GOMAXPROCS)
memoria pedida al SO    : 37.3 MB
10 000 hilos del SO a 1 MB de pila habrian pedido ~10000 MB (unos 9.8 GB)
```

> ⚠️ El guion impreso dice «~10 240 MB». Es un error de aritmética:
> 10 000 × 1 MB = **10 000 MB**. La cifra correcta ya está en el código.

### Canales

```bash
go run ./02-canales/paso1-deadlock-send        # goroutine 1 [chan send]
go run ./02-canales/paso2-deadlock-receive     # 3 multas y luego [chan receive]
go run ./02-canales/paso3-close                # TOTAL $55 + abierto=false
go vet ./02-canales/_paso4-direccional/main.go # NO COMPILA a proposito
```

Salida de paso3 ✅:
```
El llano en llamas       $15
Pedro Paramo             $35
Los detectives salvajes  $5
TOTAL                    $55
tras el cierre: multa={Titulo: Pesos:0} abierto=false
```

Paso 4 ✅ — la carpeta lleva guion bajo (`_paso4-direccional`) **a propósito**:
así `go build ./...` y `go vet ./...` la ignoran y el resto del repo queda limpio.
```
vet: invalid operation: cannot receive from send-only channel chan<- Multa
```

---

## B4 · Node contra Go (terminales 2, 3 y una cuarta para curl) ✅ verificado

**Node** (terminal 2), 3 000 millones de vueltas:
```bash
node 03-node-vs-go/node/servidor.js
```
Cliente:
```bash
bash 03-node-vs-go/cliente/prueba-node.sh
```
Log del servidor — **esto es lo que se proyecta**:
```
node escuchando en 18092
 0.949s  entra /pesado
 5.068s  sale  /pesado
 5.072s  entra /rapido
 5.072s  sale  /rapido
```
`/rapido` se mandó en el segundo 1.2 y Node **ni se enteró** hasta el 5.07.

**Go** (terminal 3), 5 000 millones de vueltas (más trabajo todavía):
```bash
go run ./03-node-vs-go/go
bash 03-node-vs-go/cliente/prueba-go.sh
```
```
go escuchando en :18090
 0.453s  entra /pesado
 0.845s  entra /rapido
 1.259s  entra /rapido
 1.674s  entra /rapido
 2.093s  entra /rapido
 2.504s  sale  /pesado
```

> `go run` la primera vez compila y tarda. **Precompilar antes de clase:**
> `go build -o /tmp/srv-go ./03-node-vs-go/go` y en clase lanzar `/tmp/srv-go`.

### Carrera de datos ✅ verificado

```bash
go run -race ./04-carrera-datos/roto        # terminal 3
bash 04-carrera-datos/cliente/carga.sh      # terminal 4
```
```
contador final: 198      <- varia cada corrida, puede salir 200
```
`WARNING: DATA RACE` aparece **3 veces**. El detector es la prueba, no el número.

```bash
go run -race ./04-carrera-datos/arreglado
bash 04-carrera-datos/cliente/carga.sh
```
```
contador final: 200      <- 200 de 200, cero avisos
```

> `go run -race` tarda ~4 s en levantar. Esperar a ver «escuchando» antes del curl.

---

## B7 · La báscula de memoria ⚠️ CORREGIDO

### Lo que cambió respecto al guion impreso

El guion fijó el límite en **32Mi** a partir del RSS medido **en macOS**
(Go 10.7–11.1 MB, Node 43–51.5 MB). Esas cifras **no aplican dentro del
contenedor Linux**: ahí Node arranca su montón según el límite del cgroup.

Medido dentro del contenedor (linux/arm64) el 22-sep-2026:

| | Go | Node |
|---|---|---|
| Memoria en reposo | **4.96 MiB** | **10.56 MiB** |
| A 32Mi | Running | **Running — la demo NO ocurre** |
| A 11Mi | Running | Running |
| **A 10Mi** | **Running (49.6%)** | **OOMKilled, exit 137** |
| A 6Mi | Running | OOMKilled |

Repetido 4 veces a 10Mi: **Node OOMKilled 4/4, Go Running 4/4.**

**El manifiesto ya está corregido a `limits.memory: 10Mi`.**
No subirlo a 11Mi (Node sobrevive) ni bajarlo a 6Mi (Go pierde margen).

### Frases del cierre, con las cifras corregidas

La frase 1 del guion («caben cinco de Go en una réplica de Node») **ya no es
defendible**: la proporción real medida en contenedor es **2.1 a 1**, no 5 a 1.
Sustitución sugerida, con la misma estructura y esta sí medida:

> «En el presupuesto de memoria de una réplica de Node caben **dos** de Go.
> Y la imagen de Go pesa **13.8 MB contra 239 MB**: diecisiete veces menos.
> Eso es una línea de su reporte, con la fecha de hoy y el método escrito.»

Las frases 2 y 3 quedan igual.

### Tamaños reales ✅ (pendiente #6 del guion, ya medido)

| | Tamaño |
|---|---|
| Binario Go con `-ldflags="-s -w"` | **5.1 MB** (el guion dice 5.3) |
| Imagen `api-go:v1` (distroless) | **13.8 MB** |
| Imagen `api-node:v1` (node:25-alpine) | **239 MB** |

### Plan B de nivel 1 — solo Docker, sin Kubernetes ✅ VERIFICADO

**Esta es la versión segura y toma 20 segundos.** Si kind falla, o si no hay
tiempo, este bloque se sostiene solo:

```bash
cd 05-kind
docker run -d --name d-go   --platform linux/arm64 --memory=10m --memory-swap=10m api-go:v1
docker run -d --name d-node --platform linux/arm64 --memory=10m --memory-swap=10m api-node:v1
sleep 7
docker ps -a --filter name=d- --format 'table {{.Names}}\t{{.Status}}'
docker stats --no-stream d-go
```
Resultado verificado: `d-go` **Up**, `d-node` **Exited (137)** con `OOMKilled=true`.

### Con kind ✅ EJECUTADO Y VERIFICADO de punta a punta

`kind` 0.33.0 y `kubectl` v1.37.0 instalados. **El clúster `backend` ya existe
y las dos imágenes ya están cargadas en el nodo.** No hay que hacer nada antes
de clase salvo comprobar que sigue vivo:

```bash
kind get clusters                 # debe decir: backend
kubectl get nodes                 # backend-control-plane  Ready
```

Si el Mac se reinició, el nodo es un contenedor de Docker y vuelve solo al
arrancar Docker. Si no vuelve: `kind delete cluster --name backend` y repetir
el montaje (tarda ~1 min):

```bash
cd 05-kind
kind create cluster --config kind-config.yaml
kind load docker-image api-go:v1 api-node:v1 --name backend
```

**En clase el único comando nuevo es `kubectl apply`:**

```bash
kubectl apply -f despliegue.yaml
kubectl get pods -w
kubectl describe pod -l app=api-node | sed -n '/Last State/,/Ready/p'
kubectl scale deployment/api-go --replicas=5
```

Salida verificada de `kubectl get pods -w`:

```
api-go-...     Running     reinicios=0
api-node-...   OOMKilled   reinicios=1
api-node-...   OOMKilled   reinicios=3
```

> ⚠️ **Node parpadea en `Running` uno o dos segundos** antes de morir, y el
> contador de reinicios sube despacio. No es que la demostración falle: hay
> que esperar entre 30 y 60 segundos y narrar los reinicios mientras suben.
> A los 10 s ya hay un `OOMKilled` en pantalla.

Salida verificada del `describe` — **esto es lo que se proyecta**:

```
    Last State:     Terminated
      Reason:       OOMKilled
      Exit Code:    137
      Started:      Tue, 22 Sep 2026 13:05:53 -0700
      Finished:     Tue, 22 Sep 2026 13:05:53 -0700
    Ready:          False
```

`Started` y `Finished` **en el mismo segundo**: Node muere durante el arranque,
no bajo carga. Señalarlo con el dedo, es lo más contundente del bloque.

Y el remate, verificado: `kubectl scale deployment/api-go --replicas=5` deja
**cinco réplicas de Go en Running con cero reinicios**, mientras la única de
Node sigue sin poder arrancar — todas con el mismo límite de 10Mi.

Comprobación de que Go de verdad responde (terminal dedicada):

```bash
kubectl port-forward svc/api-go 8080:80
curl http://localhost:8080/salud      # -> ok desde go
```

`kubectl top` no funciona en kind sin `metrics-server`: **saltarse ese paso.**

### Limpieza entre ensayo y clase

```bash
kubectl delete -f despliegue.yaml     # ya hecho: la clase arranca sin pods
```

### Imágenes construidas ✅

```bash
cd 05-kind
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o api-go ./src-go/main.go
docker build --platform linux/arm64 -f Dockerfile.api-go   -t api-go:v1   .
docker build --platform linux/arm64 -f Dockerfile.api-node -t api-node:v1 .
```

> Dos trampas ya resueltas en el repo:
> 1. El archivo **no** se llama `Dockerfile.go` (cualquier `go build`/`gofmt`
>    en la carpeta lo parsearía como Go y fallaría con `illegal character U+0023 '#'`).
> 2. El fuente vive en `src-go/`, **no** en `api-go/`: si la carpeta se llama
>    igual que el binario, `COPY api-go /app/api-go` copia la carpeta y el
>    contenedor muere con `exec: "/app/api-go": is a directory`.

---

## Estado de los pendientes del ensayo

| # | Pendiente del guion | Estado |
|---|---|---|
| 1 | ¿Existe `node:25-alpine`? | ✅ **Sí existe.** No hace falta plan B |
| 2 | ¿Node muere con 32 MiB? | ❌ **NO.** Sobrevive. Muere a **10Mi** — manifiesto corregido |
| 3 | ¿Go sobrevive con el límite? | ✅ Sí, 4.96 MiB de 10Mi (49.6%). `GOMEMLIMIT` no hace falta |
| 4 | `runAsNonRoot` con usuario numérico | ✅ Resuelto: 65532 (Go) y 1000 (Node) |
| 5 | `kubectl top` en kind | ⏭️ Se salta, como indica el guion |
| — | **B7 completo con kind** | ✅ **Ejecutado de punta a punta.** Ya no es el bloque sin verificar |
| 6 | Tamaños reales de las imágenes | ✅ Medidos: 13.8 MB contra 239 MB |

## Limpieza al terminar

```bash
docker rm -f d-go d-node p-go p-node 2>/dev/null
lsof -ti:18090,18092,18094 | xargs -r kill -9
```
