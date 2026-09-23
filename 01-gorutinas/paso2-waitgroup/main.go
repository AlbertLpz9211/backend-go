// B3 paso 2 - El WaitGroup, con la version secuencial para comparar.
package main

import (
	"fmt"
	"sync"
	"time"
)

const tareas = 50

func consultarCatalogo(n int) {
	time.Sleep(100 * time.Millisecond)
	_ = n
}

func main() {
	// --- Secuencial: cada consulta espera a que termine la anterior ---
	inicio := time.Now()
	for i := 1; i <= tareas; i++ {
		consultarCatalogo(i)
	}
	fmt.Printf("secuencial  : %d consultas en %v\n", tareas, time.Since(inicio))

	// --- Concurrente: las 50 esperas ocurren al mismo tiempo ---
	// El WaitGroup es un contador con sala de espera: Add suma, Done
	// resta y Wait bloquea hasta que el contador llega a cero.
	inicio = time.Now()
	var wg sync.WaitGroup
	for i := 1; i <= tareas; i++ {
		wg.Add(1) // SIEMPRE antes del 'go': dentro hay carrera con Wait
		go func(n int) {
			defer wg.Done() // defer lo ejecuta aunque la gorutina entre en panico
			consultarCatalogo(n)
		}(i)
	}
	wg.Wait()
	fmt.Printf("concurrente : %d consultas en %v\n", tareas, time.Since(inicio))
}
