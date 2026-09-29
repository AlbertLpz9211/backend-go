# Desarrollo Web de Software de Servidor (Back-End) con Go

Universidad Nueva Galicia · DSING008 · Cuarto cuatrimestre, septiembre–diciembre 2026.

Código de las demostraciones de clase. Cada sesión vive en su propia carpeta
`demos-sesionN/` y es un módulo de Go independiente, con su propio `go.mod`.

| Sesión | Fecha | Tema | Carpeta |
| --- | --- | --- | --- |
| 2 | 22 sep | 1.2 Go frente a Node.js, Python y Java. Gorutinas y canales | [`demos-sesion2/`](demos-sesion2/) |

## Requisitos

- Go 1.26 o superior (`go version`).
- Docker y [kind](https://kind.sigs.k8s.io/) con `kubectl`, solo para los bloques de Kubernetes.

## Cómo usar una sesión

```bash
cd demos-sesion2
go run ./01-gorutinas/paso2-waitgroup   # cada carpeta es un programa completo
```

Las instrucciones y salidas esperadas de cada sesión están en su bitácora
(por ejemplo, `demos-sesion2/ENSAYO.md`).

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
