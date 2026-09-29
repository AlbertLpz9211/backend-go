// api — versión final para el contenedor (paso 5 + lo necesario para Kubernetes).
//
// Qué agrega respecto al paso 5:
//   - Cabecera X-Pod en TODAS las respuestas, con el nombre de la máquina
//     (os.Hostname). En Kubernetes el hostname de un contenedor es el nombre
//     del pod, así vemos QUÉ réplica nos atendió.
//   - Un middleware de registro: una línea JSON (log/slog) por petición.
//   - http.Server con timeouts, igual que en la knowledge base.
//   - Apagado ordenado: al recibir SIGTERM (lo que envía Kubernetes antes de
//     matar un pod) termina las peticiones en curso con srv.Shutdown.
//   - Las rutas se arman en la función rutas() para poder probarlas con
//     httptest sin levantar un servidor real (ver main_test.go).
//
// Ejecutar: go run ./api            (o: PORT=9090 go run ./api)
// Probar:   ./probar.sh
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// Libro es el recurso de la API.
type Libro struct {
	ID     int64  `json:"id"`
	Titulo string `json:"titulo"`
	Autor  string `json:"autor"`
	Anio   int    `json:"anio"`
}

// errorAPI es la forma única de todos los errores.
type errorAPI struct {
	Codigo  string `json:"codigo"`
	Mensaje string `json:"mensaje"`
}

// almacen en memoria. OJO: cada réplica (pod) tiene el SUYO y no lo comparte.
type almacen struct {
	mu     sync.RWMutex
	libros map[int64]Libro
	sig    int64
}

var errNoEncontrado = errors.New("no encontrado")

// nuevoAlmacen crea el almacén con dos libros de semilla.
func nuevoAlmacen() *almacen {
	a := &almacen{libros: map[int64]Libro{}}
	a.crear(Libro{Titulo: "Pedro Páramo", Autor: "Juan Rulfo", Anio: 1955})
	a.crear(Libro{Titulo: "Como agua para chocolate", Autor: "Laura Esquivel", Anio: 1989})
	return a
}

func (a *almacen) listar() []Libro {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]Libro, 0, len(a.libros)) // make, no var: JSON debe ser [] y no null
	for _, l := range a.libros {
		out = append(out, l)
	}
	return out
}

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
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sig++
	l.ID = a.sig
	a.libros[l.ID] = l
	return l
}

func (a *almacen) eliminar(id int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.libros[id]; !ok {
		return errNoEncontrado
	}
	delete(a.libros, id)
	return nil
}

func escribirJSON(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// rutas registra todos los endpoints y devuelve el mux.
func rutas(a *almacen) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/salud", func(w http.ResponseWriter, r *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
	})

	mux.HandleFunc("GET /api/v1/libros", func(w http.ResponseWriter, r *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]any{"datos": a.listar()})
	})

	mux.HandleFunc("GET /api/v1/libros/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			escribirJSON(w, http.StatusBadRequest, errorAPI{"id_invalido", "el id debe ser un entero"})
			return
		}
		l, err := a.obtener(id)
		if errors.Is(err, errNoEncontrado) {
			escribirJSON(w, http.StatusNotFound, errorAPI{"no_encontrado", "el libro no existe"})
			return
		}
		escribirJSON(w, http.StatusOK, l)
	})

	mux.HandleFunc("POST /api/v1/libros", func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "" {
			tipo, _, err := mime.ParseMediaType(ct)
			if err != nil || tipo != "application/json" {
				escribirJSON(w, http.StatusUnsupportedMediaType, errorAPI{"tipo_no_soportado", "se espera application/json"})
				return
			}
		}
		var in Libro
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)) // techo de 1 MiB
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			escribirJSON(w, http.StatusBadRequest, errorAPI{"json_invalido", err.Error()})
			return
		}
		if in.Titulo == "" {
			escribirJSON(w, http.StatusUnprocessableEntity, errorAPI{"validacion", "titulo es obligatorio"})
			return
		}
		creado := a.crear(in)
		w.Header().Set("Location", "/api/v1/libros/"+strconv.FormatInt(creado.ID, 10))
		escribirJSON(w, http.StatusCreated, creado)
	})

	mux.HandleFunc("DELETE /api/v1/libros/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			escribirJSON(w, http.StatusBadRequest, errorAPI{"id_invalido", "el id debe ser un entero"})
			return
		}
		if err := a.eliminar(id); errors.Is(err, errNoEncontrado) {
			escribirJSON(w, http.StatusNotFound, errorAPI{"no_encontrado", "el libro no existe"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}

// ---------------------------------------------------------------------------
// MIDDLEWARE, explicado para principiantes.
//
// Un middleware es una función que RECIBE un handler y DEVUELVE otro handler
// que hace "algo más" antes y/o después de llamar al original. Es como poner
// una capa alrededor de la cebolla:
//
//	petición -> conRegistro -> conPod -> mux -> tu handler
//
// Todos tienen la misma forma:  func(siguiente http.Handler) http.Handler
// ---------------------------------------------------------------------------

// conPod añade la cabecera X-Pod a TODAS las respuestas (incluso a los 404 y
// 405 que genera el propio mux). La ponemos ANTES de llamar a siguiente,
// porque una vez que el handler escribe el código, las cabeceras ya se fueron.
func conPod(pod string, siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Pod", pod)
		siguiente.ServeHTTP(w, r)
	})
}

// grabadora "espía" el código de estado que escribe el handler, para poder
// registrarlo. Envuelve al ResponseWriter original y solo intercepta WriteHeader.
type grabadora struct {
	http.ResponseWriter
	estado int
}

func (g *grabadora) WriteHeader(estado int) {
	g.estado = estado
	g.ResponseWriter.WriteHeader(estado)
}

// conRegistro escribe una línea JSON por petición: método, ruta, estado, duración.
func conRegistro(log *slog.Logger, siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		g := &grabadora{ResponseWriter: w, estado: http.StatusOK} // 200 si nadie llama a WriteHeader
		siguiente.ServeHTTP(g, r)
		log.Info("peticion",
			slog.String("metodo", r.Method),
			slog.String("ruta", r.URL.Path),
			slog.Int("estado", g.estado),
			slog.String("duracion", time.Since(inicio).String()), // p. ej. "215µs"
		)
	})
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// En Kubernetes, el hostname del contenedor = nombre del pod.
	pod, err := os.Hostname()
	if err != nil {
		pod = "desconocido"
	}

	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}

	a := nuevoAlmacen()
	manejador := conRegistro(log, conPod(pod, rutas(a)))

	// http.Server con timeouts: sin ellos, un cliente lento puede mantener
	// conexiones abiertas para siempre.
	srv := &http.Server{
		Addr:              ":" + puerto,
		Handler:           manejador,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	// Arrancamos el servidor en otra gorutina para que main pueda quedarse
	// esperando la señal de apagado.
	go func() {
		log.Info("escuchando", slog.String("addr", srv.Addr), slog.String("pod", pod))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("servidor", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Apagado ordenado. Kubernetes manda SIGTERM antes de borrar un pod
	// (Ctrl+C en la terminal manda os.Interrupt). En vez de morir de golpe,
	// dejamos de aceptar conexiones y terminamos las que están en curso.
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()
	<-ctx.Done()
	log.Info("apagando")
	ctxApagado, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	if err := srv.Shutdown(ctxApagado); err != nil {
		log.Error("apagado", slog.Any("error", err))
	}
}
