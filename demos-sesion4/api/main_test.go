package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// nuevoManejador arma la API completa (con middlewares) sobre un almacén
// NUEVO con la semilla (libros 1 y 2). Cada caso de la tabla empieza limpio.
func nuevoManejador() http.Handler {
	return conPod("pod-de-prueba", rutas(nuevoAlmacen()))
}

// TestCodigos es una "prueba de tabla": cada fila es un caso con la petición
// y lo que esperamos. Añadir un caso = añadir una fila.
func TestCodigos(t *testing.T) {
	casos := []struct {
		nombre      string
		metodo      string
		ruta        string
		contentType string
		cuerpo      string
		codigo      int    // código HTTP esperado
		contiene    string // texto que debe aparecer en el cuerpo ("" = no revisar)
		cabecera    string // cabecera que debe venir ("" = no revisar)
	}{
		{"salud", "GET", "/api/v1/salud", "", "", 200, `"estado":"ok"`, ""},
		{"listar", "GET", "/api/v1/libros", "", "", 200, `"datos":[`, ""},
		{"obtener existente", "GET", "/api/v1/libros/1", "", "", 200, "Pedro Páramo", ""},
		{"obtener id no numérico", "GET", "/api/v1/libros/abc", "", "", 400, "id_invalido", ""},
		{"obtener inexistente", "GET", "/api/v1/libros/99", "", "", 404, "no_encontrado", ""},
		{"crear", "POST", "/api/v1/libros", "application/json", `{"titulo":"Aura","autor":"Carlos Fuentes","anio":1962}`, 201, `"id":3`, "Location"},
		{"crear con charset", "POST", "/api/v1/libros", "application/json; charset=utf-8", `{"titulo":"Aura"}`, 201, `"id":3`, "Location"},
		{"crear JSON roto", "POST", "/api/v1/libros", "application/json", `{"titulo":`, 400, "json_invalido", ""},
		{"crear campo desconocido", "POST", "/api/v1/libros", "application/json", `{"titulo":"Aura","editorial":"Era"}`, 400, "json_invalido", ""},
		{"crear título vacío", "POST", "/api/v1/libros", "application/json", `{"titulo":""}`, 422, "validacion", ""},
		{"crear tipo incorrecto", "POST", "/api/v1/libros", "text/plain", `hola`, 415, "tipo_no_soportado", ""},
		{"eliminar existente", "DELETE", "/api/v1/libros/1", "", "", 204, "", ""},
		{"eliminar inexistente", "DELETE", "/api/v1/libros/99", "", "", 404, "no_encontrado", ""},
		{"método no permitido", "PUT", "/api/v1/libros/1", "application/json", `{}`, 405, "", "Allow"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			req := httptest.NewRequest(c.metodo, c.ruta, strings.NewReader(c.cuerpo))
			if c.contentType != "" {
				req.Header.Set("Content-Type", c.contentType)
			}
			rec := httptest.NewRecorder() // un ResponseWriter "de mentira" que guarda todo

			nuevoManejador().ServeHTTP(rec, req)

			if rec.Code != c.codigo {
				t.Fatalf("código = %d, se esperaba %d (cuerpo: %s)", rec.Code, c.codigo, rec.Body)
			}
			if c.contiene != "" && !strings.Contains(rec.Body.String(), c.contiene) {
				t.Errorf("el cuerpo %q no contiene %q", rec.Body, c.contiene)
			}
			if c.cabecera != "" && rec.Header().Get(c.cabecera) == "" {
				t.Errorf("falta la cabecera %s", c.cabecera)
			}
			if rec.Header().Get("X-Pod") != "pod-de-prueba" {
				t.Errorf("X-Pod = %q, se esperaba pod-de-prueba", rec.Header().Get("X-Pod"))
			}
		})
	}
}

// TestCrearConcurrente lanza 100 POST a la vez. Con `go test -race` detecta
// si alguien quita el mutex del almacén.
func TestCrearConcurrente(t *testing.T) {
	h := nuevoManejador()
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cuerpo := fmt.Sprintf(`{"titulo":"Libro %d"}`, i)
			req := httptest.NewRequest("POST", "/api/v1/libros", strings.NewReader(cuerpo))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusCreated {
				t.Errorf("código = %d", rec.Code)
			}
		}()
	}
	wg.Wait()

	req := httptest.NewRequest("GET", "/api/v1/libros/102", nil) // 2 de semilla + 100
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("el libro 102 debería existir; código = %d", rec.Code)
	}
}
