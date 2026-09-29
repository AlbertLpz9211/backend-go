// Canales paso 3 - Una sola linea de diferencia: defer close(salida).
package main

import "fmt"

type Prestamo struct {
	Titulo string
	Dias   int
}

type Multa struct {
	Titulo string
	Pesos  int
}

func calcularMultas(salida chan<- Multa, prestamos []Prestamo) {
	// Cierra SIEMPRE quien envia, y una sola vez. Cerrar dos veces
	// es panico; enviar a un canal cerrado tambien.
	defer close(salida)

	for _, p := range prestamos {
		salida <- Multa{Titulo: p.Titulo, Pesos: p.Dias * 5}
	}
}

func main() {
	prestamos := []Prestamo{
		{"El llano en llamas", 3},
		{"Pedro Paramo", 7},
		{"Los detectives salvajes", 1},
	}

	salida := make(chan Multa)
	go calcularMultas(salida, prestamos)

	total := 0
	for m := range salida {
		fmt.Printf("%-25s$%d\n", m.Titulo, m.Pesos)
		total += m.Pesos
	}
	fmt.Printf("%-25s$%d\n", "TOTAL", total)

	// Cerrar no es tirar la basura, pero el canal ya quedo vacio:
	// recibir de un canal cerrado y vacio devuelve el valor cero.
	multa, abierto := <-salida
	fmt.Printf("tras el cierre: multa={Titulo:%s Pesos:%d} abierto=%v\n",
		multa.Titulo, multa.Pesos, abierto)
}
