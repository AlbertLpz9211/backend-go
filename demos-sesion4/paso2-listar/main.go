// Paso 2 — Listar libros.
//
// Qué agrega respecto al paso 1:
//   - El tipo Libro (con etiquetas `json:"..."` para los nombres en el JSON).
//   - Un almacén en memoria protegido con sync.RWMutex.
//   - La función escribirJSON, que centraliza cabecera + código + cuerpo.
//   - GET /api/v1/libros  ->  {"datos":[]}
//
// Ejecutar:   go run ./paso2-listar
// Probar:     curl -i localhost:8080/api/v1/libros
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

// Libro es el RECURSO de nuestra API. Las etiquetas json dicen cómo se
// llama cada campo al convertirlo a JSON (minúsculas, sin acentos).
type Libro struct {
	ID     int64  `json:"id"`
	Titulo string `json:"titulo"`
	Autor  string `json:"autor"`
	Anio   int    `json:"anio"`
}

// almacen guarda los libros EN MEMORIA: si el proceso se reinicia, se pierden.
// (La base de datos llega en la unidad 3.)
// El mutex es obligatorio: net/http atiende cada petición en su propia
// gorutina, y un map compartido sin protección puede tumbar el programa.
type almacen struct {
	mu     sync.RWMutex
	libros map[int64]Libro
	sig    int64 // último id asignado
}

func (a *almacen) listar() []Libro {
	a.mu.RLock() // lectura: varias gorutinas pueden leer a la vez
	defer a.mu.RUnlock()
	out := make([]Libro, 0, len(a.libros)) // make, no var: JSON debe ser [] y no null
	for _, l := range a.libros {
		out = append(out, l)
	}
	return out
}

// VERSIÓN CON BUG (para mostrar en clase): cambia listar por esta y
// la respuesta pasa de {"datos":[]} a {"datos":null}.
// Un slice declarado con var vale nil, y encoding/json convierte nil en null.
//
// func (a *almacen) listar() []Libro {
// 	a.mu.RLock()
// 	defer a.mu.RUnlock()
// 	var out []Libro
// 	for _, l := range a.libros {
// 		out = append(out, l)
// 	}
// 	return out
// }

// escribirJSON: una sola función para responder JSON. Sin ella, cada handler
// repite tres líneas y alguna se olvida.
func escribirJSON(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // 1) cabeceras
	w.WriteHeader(estado)                                             // 2) código
	if v != nil {
		_ = json.NewEncoder(w).Encode(v) // 3) cuerpo
	}
}

func main() {
	a := &almacen{libros: map[int64]Libro{}}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/salud", func(w http.ResponseWriter, r *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
	})

	// Envolvemos la lista en {"datos": ...}: así mañana podemos añadir
	// "total" o "pagina" sin romper a los clientes.
	mux.HandleFunc("GET /api/v1/libros", func(w http.ResponseWriter, r *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]any{"datos": a.listar()})
	})

	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	log.Println("escuchando en :" + puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, mux))
}
