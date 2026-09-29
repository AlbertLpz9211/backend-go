// Canales paso 1 - Bloqueo mutuo del lado del que ENVIA.
// Etiqueta esperada: goroutine 1 [chan send]
package main

import "fmt"

type Multa struct {
	Titulo string
	Pesos  int
}

func main() {
	// Canal SIN bufer: enviar no se suelta hasta que alguien recibe.
	salida := make(chan Multa)

	// Y aqui no hay nadie mas: main se envia a si mismo.
	salida <- Multa{Titulo: "El llano en llamas", Pesos: 15}

	fmt.Println("esta linea no se imprime jamas")
}
