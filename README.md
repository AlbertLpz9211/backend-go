# Desarrollo Web de Software de Servidor (Back-End) con Go

Universidad Nueva Galicia · DSING008 · Cuarto cuatrimestre, septiembre–diciembre 2026.

Código de las demostraciones de clase. Cada sesión vive en su propia carpeta
`demos-sesionN/` y es un módulo de Go independiente, con su propio `go.mod`.

| Sesión | Fecha | Tema | Carpeta |
| --- | --- | --- | --- |
| 2 | 22 sep | 1.2 Go frente a Node.js, Python y Java. Gorutinas y canales | [`demos-sesion2/`](demos-sesion2/) |
| 4 | 29 sep | 2.1 REST y `net/http`. La API en Kubernetes (kind) | [`demos-sesion4/`](demos-sesion4/) |

## Requisitos

- Go 1.26 o superior (`go version`).
- Docker y [kind](https://kind.sigs.k8s.io/) con `kubectl`, solo para los bloques de Kubernetes.

## macOS y Windows

- `go run`, `go build` y `go test` funcionan igual en los dos sistemas.
- Los scripts `.sh` son de **bash**. En macOS y Linux se usan tal cual; en Windows se corren desde
  **Git Bash** (viene con Git para Windows), no desde PowerShell ni CMD.
- En PowerShell, `curl` es un alias de `Invoke-WebRequest` y no acepta las mismas opciones. Si
  no tienes Git Bash, escribe `curl.exe`.
- El repositorio fija finales de línea LF (`.gitattributes`), así que los `.sh` funcionan aunque
  clones en Windows.
- Kubernetes con kind requiere Docker Desktop (en Windows, con WSL2).
- `go test -race` en Windows necesita un compilador de C (gcc). Si no lo tienes, usa `go test`.

## Cómo usar una sesión

```bash
cd demos-sesion2
go run ./01-gorutinas/paso2-waitgroup   # cada carpeta es un programa completo
```

Las instrucciones y salidas esperadas de cada sesión están en su bitácora
(por ejemplo, `demos-sesion2/ENSAYO.md`, `demos-sesion4/SALIDAS.md`).

## Ramas: una por día de clase

- `main` contiene solo las sesiones **ya impartidas**.
- Cada día de clase se prepara en su propia rama, `sesion-NN` (dos dígitos: `sesion-04`,
  `sesion-05`, …), que nace de `main` y agrega su carpeta `demos-sesionN/` y su fila en la
  tabla de arriba.
- El día de la clase, esa rama se une a `main` y se publica.

```bash
# Preparar la próxima sesión
git switch main && git pull
git switch -c sesion-05
# ... trabajo, commits ...

# El día de la clase: unir y publicar
git switch main
git merge --no-ff sesion-05
git push origin main sesion-05
```

Las correcciones posteriores a una sesión ya impartida se hacen en su rama y se vuelven a unir
a `main`.
