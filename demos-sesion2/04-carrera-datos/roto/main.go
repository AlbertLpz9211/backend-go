// B4 - El precio de esa concurrencia gratis: una carrera de datos.
// Ejecutar SIEMPRE con: go run -race ./04-carrera-datos/roto
package main

import (
	"fmt"
	"net/http"
)

// Estado compartido entre TODAS las peticiones. Nadie escribio 'go'
// en este archivo, y aun asi este contador lo tocan varias gorutinas.
var visitas int

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /libros/{id}", func(w http.ResponseWriter, r *http.Request) {
		visitas++ // no es atomico: leer, sumar y escribir
		fmt.Fprintf(w, "libro %s, visita numero %d\n", r.PathValue("id"), visitas)
	})

	mux.HandleFunc("GET /total", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%d\n", visitas)
	})

	fmt.Println("roto escuchando en :18094")
	http.ListenAndServe(":18094", mux)
}
