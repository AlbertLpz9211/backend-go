// Paso 1 — Hola, servidor.
//
// Qué agrega: lo mínimo para tener una API viva.
//   - Un ServeMux (el "enrutador" de la biblioteca estándar).
//   - Una sola ruta: GET /api/v1/salud  ->  {"estado":"ok"}
//   - El puerto se lee de la variable de entorno PORT (por omisión 8080).
//
// Ejecutar:   go run ./paso1-hola
// Probar:     curl -i localhost:8080/api/v1/salud
package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// El mux decide qué función atiende cada petición según MÉTODO + RUTA.
	mux := http.NewServeMux()

	// Desde Go 1.22 el patrón puede incluir el método: "GET /ruta".
	mux.HandleFunc("GET /api/v1/salud", func(w http.ResponseWriter, r *http.Request) {
		// Orden obligatorio: 1) cabeceras, 2) código, 3) cuerpo.
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"estado":"ok"}` + "\n"))
	})

	// Leemos el puerto de una variable de entorno: así el mismo programa
	// sirve en tu laptop, en Docker y en Kubernetes sin recompilar.
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}

	log.Println("escuchando en :" + puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, mux))
}
