// B3 paso 3 - De 50 a 10 000.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const tareas = 10000

func consultarCatalogo(n int) {
	time.Sleep(100 * time.Millisecond)
	_ = n
}

func main() {
	inicio := time.Now()

	var wg sync.WaitGroup
	var pico int
	var mu sync.Mutex

	// Vigila el numero de gorutinas vivas mientras corre el lote.
	listo := make(chan struct{})
	go func() {
		for {
			select {
			case <-listo:
				return
			default:
				// -1 porque esta misma gorutina vigilante tambien cuenta.
				if n := runtime.NumGoroutine() - 1; n > pico {
					mu.Lock()
					pico = n
					mu.Unlock()
				}
				time.Sleep(time.Millisecond)
			}
		}
	}()

	for i := 1; i <= tareas; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			consultarCatalogo(n)
		}(i)
	}
	wg.Wait()
	transcurrido := time.Since(inicio)
	close(listo)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("%d gorutinas en %v\n", tareas, transcurrido)
	fmt.Printf("pico de gorutinas vivas : %d\n", pico)
	fmt.Printf("nucleos logicos usables : %d (GOMAXPROCS)\n", runtime.GOMAXPROCS(0))
	fmt.Printf("memoria pedida al SO    : %.1f MB\n", float64(m.Sys)/(1024*1024))
	fmt.Printf("10 000 hilos del SO a 1 MB de pila habrian pedido ~%d MB (unos %.1f GB)\n",
		tareas, float64(tareas)/1024)
}
