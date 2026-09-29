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

## Cómo usar una sesión

```bash
cd demos-sesion4
go run ./paso1-hola        # cada carpeta es un programa completo
```

Las instrucciones y salidas esperadas de cada sesión están en su bitácora
(`demos-sesion2/ENSAYO.md`, `demos-sesion4/SALIDAS.md`).
