// Canales paso 4 - NO COMPILA A PROPOSITO.
// Demuestra que los canales direccionales los impone el compilador.
// Vive en una carpeta con guion bajo para que ./... no lo intente compilar.
package main

import "fmt"

type Multa struct {
	Titulo string
	Pesos  int
}

func calcularMultas(salida chan<- Multa) {
	// salida es SOLO DE ENVIO: recibir de el es un error de compilacion.
	m := <-salida
	fmt.Println(m)
}

func consumir(entrada <-chan Multa) {
	for m := range entrada {
		fmt.Println(m)
	}
	// entrada es SOLO DE RECEPCION: cerrarlo es un error de compilacion.
	close(entrada)
}

func main() {
	ch := make(chan Multa)
	go calcularMultas(ch)
	consumir(ch)
}
