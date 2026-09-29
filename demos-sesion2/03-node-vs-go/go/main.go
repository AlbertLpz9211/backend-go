// B4 - Go: la misma carga, con MAS trabajo todavia (5 000 millones
// de vueltas contra 3 000 millones de Node).
// Observese que la palabra 'go' no aparece en ninguna parte.
package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	puerto  = ":18090"
	vueltas = 5_000_000_000
)

var inicio = time.Now()

func t() string {
	return fmt.Sprintf("%6.3f", time.Since(inicio).Seconds())
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /pesado", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%ss  entra /pesado\n", t())
		suma := 0
		for i := 0; i < vueltas; i++ {
			suma += i
		}
		fmt.Printf("%ss  sale  /pesado\n", t())
		fmt.Fprintf(w, "pesado listo, suma=%d\n", suma)
	})

	mux.HandleFunc("GET /rapido", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%ss  entra /rapido\n", t())
		fmt.Fprintln(w, "rapido")
	})

	fmt.Printf("go escuchando en %s\n", puerto)
	// net/http lanza UNA GORUTINA POR PETICION. Nosotros no escribimos 'go'.
	http.ListenAndServe(puerto, mux)
}
