// Canales paso 2 - Ahora si hay productor, pero nunca cierra.
// Imprime las tres multas Y AUN ASI muere.
// Etiqueta esperada: goroutine 1 [chan receive]
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

func calcularMultas(salida chan Multa, prestamos []Prestamo) {
	for _, p := range prestamos {
		salida <- Multa{Titulo: p.Titulo, Pesos: p.Dias * 5}
	}
	// FALTA close(salida). Por eso el range de abajo nunca termina.
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
	// range sobre un canal NO termina cuando el canal se vacia:
	// termina cuando alguien lo CIERRA.
	for m := range salida {
		fmt.Printf("%-25s$%d\n", m.Titulo, m.Pesos)
		total += m.Pesos
	}

	fmt.Printf("%-25s$%d\n", "TOTAL", total)
	fmt.Println("corte de caja terminado") // nunca se imprime
}
