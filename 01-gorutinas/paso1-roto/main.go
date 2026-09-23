// B3 paso 1 - Roto a proposito.
// 'go' arranca y sigue. Cuando main retorna, el proceso muere.
package main

import (
	"fmt"
	"time"
)

func consultarCatalogo(n int) {
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("consulta %d lista\n", n)
}

func main() {
	inicio := time.Now()

	// 'go' arranca la funcion y devuelve el control en el acto.
	// Nadie le dijo a main que esperara, asi que main no espera.
	for i := 1; i <= 50; i++ {
		fmt.Printf("consulta %d lanzada\n", i)
		go consultarCatalogo(i)
	}

	// Cuando main retorna el proceso muere y se lleva por delante
	// las 50 gorutinas, esten donde esten.
	fmt.Printf("main termino en %v\n", time.Since(inicio))
}
