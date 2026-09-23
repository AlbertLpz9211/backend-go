// B4 - La correccion: atomic.Int64. 200 de 200 y cero avisos de -race.
package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Add es una sola instruccion del procesador: nadie se puede meter en medio.
var visitas atomic.Int64

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /libros/{id}", func(w http.ResponseWriter, r *http.Request) {
		n := visitas.Add(1)
		fmt.Fprintf(w, "libro %s, visita numero %d\n", r.PathValue("id"), n)
	})

	mux.HandleFunc("GET /total", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%d\n", visitas.Load())
	})

	fmt.Println("arreglado escuchando en :18094")
	http.ListenAndServe(":18094", mux)
}
