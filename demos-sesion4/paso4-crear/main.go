// Paso 4 — Crear libros.
//
// Qué agrega respecto al paso 3:
//   - POST /api/v1/libros con validación en capas:
//   - 415 si el Content-Type no es application/json.
//   - 400 si el JSON está mal formado o trae un campo desconocido
//     (DisallowUnknownFields).
//   - 422 si el JSON está bien pero el título viene vacío.
//   - 201 + cabecera Location si todo salió bien.
//   - http.MaxBytesReader: techo de 1 MiB para el cuerpo.
//
// Ejecutar: go run ./paso4-crear
//
// Probar:
//   - curl -i -X POST localhost:8080/api/v1/libros -H 'Content-Type: application/json' -d '{"titulo":"Aura","autor":"Carlos Fuentes","anio":1962}'
package main

import (
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
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

// errorAPI es la forma ÚNICA de todos los errores de la API.
// "codigo" es para las máquinas (estable); "mensaje" es para las personas.
type errorAPI struct {
	Codigo  string `json:"codigo"`
	Mensaje string `json:"mensaje"`
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

// errNoEncontrado es un error "centinela": lo comparamos con errors.Is.
var errNoEncontrado = errors.New("no encontrado")

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

func (a *almacen) obtener(id int64) (Libro, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	l, ok := a.libros[id]
	if !ok {
		return Libro{}, errNoEncontrado
	}
	return l, nil
}

func (a *almacen) crear(l Libro) Libro {
	a.mu.Lock() // escritura: exclusiva, nadie más lee ni escribe mientras tanto
	defer a.mu.Unlock()
	a.sig++
	l.ID = a.sig
	a.libros[l.ID] = l
	return l
}

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
	// Semilla: dos libros para tener qué consultar desde el primer curl.
	a.crear(Libro{Titulo: "Pedro Páramo", Autor: "Juan Rulfo", Anio: 1955})
	a.crear(Libro{Titulo: "Como agua para chocolate", Autor: "Laura Esquivel", Anio: 1989})

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/salud", func(w http.ResponseWriter, r *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
	})

	// Envolvemos la lista en {"datos": ...}: así mañana podemos añadir
	// "total" o "pagina" sin romper a los clientes.
	mux.HandleFunc("GET /api/v1/libros", func(w http.ResponseWriter, r *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]any{"datos": a.listar()})
	})

	// {id} es un COMODÍN: /api/v1/libros/1, /api/v1/libros/2, ...
	// r.PathValue("id") nos da el texto que venía en esa posición.
	mux.HandleFunc("GET /api/v1/libros/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			escribirJSON(w, http.StatusBadRequest, errorAPI{"id_invalido", "el id debe ser un entero"})
			return // ¡no olvidar el return tras responder un error!
		}
		l, err := a.obtener(id)
		if errors.Is(err, errNoEncontrado) {
			escribirJSON(w, http.StatusNotFound, errorAPI{"no_encontrado", "el libro no existe"})
			return
		}
		escribirJSON(w, http.StatusOK, l)
	})

	mux.HandleFunc("POST /api/v1/libros", func(w http.ResponseWriter, r *http.Request) {
		// Capa 1 — ¿me mandas JSON? Si no, 415.
		// mime.ParseMediaType acepta también "application/json; charset=utf-8".
		if ct := r.Header.Get("Content-Type"); ct != "" {
			tipo, _, err := mime.ParseMediaType(ct)
			if err != nil || tipo != "application/json" {
				escribirJSON(w, http.StatusUnsupportedMediaType, errorAPI{"tipo_no_soportado", "se espera application/json"})
				return
			}
		}

		// Capa 2 — ¿el JSON se puede leer? Si no, 400.
		var in Libro
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)) // techo de 1 MiB
		dec.DisallowUnknownFields()                                   // una errata del cliente = 400, no silencio
		if err := dec.Decode(&in); err != nil {
			escribirJSON(w, http.StatusBadRequest, errorAPI{"json_invalido", err.Error()})
			return
		}

		// Capa 3 — ¿cumple las reglas del negocio? Si no, 422.
		if in.Titulo == "" {
			escribirJSON(w, http.StatusUnprocessableEntity, errorAPI{"validacion", "titulo es obligatorio"})
			return
		}

		// Todo bien: 201 Created + Location (dónde quedó lo que se creó).
		creado := a.crear(in)
		w.Header().Set("Location", "/api/v1/libros/"+strconv.FormatInt(creado.ID, 10))
		escribirJSON(w, http.StatusCreated, creado)
	})

	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	log.Println("escuchando en :" + puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, mux))
}
