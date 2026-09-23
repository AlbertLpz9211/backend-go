// B7 - API minima en Go para la bascula de memoria.
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /salud", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok desde go")
	})
	mux.HandleFunc("GET /libros/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "libro %s\n", r.PathValue("id"))
	})
	fmt.Println("api-go escuchando en " + puerto)
	http.ListenAndServe(":"+puerto, mux)
}
