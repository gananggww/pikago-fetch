// package main

// import (
// 	"fmt"
// 	"sync"
// 	"time"
// )

// func main() {
// 	var x int = 0
// 	var mu sync.Mutex
// 	for i := 0; i < 50; i++ {
// 		go func() {
// 			mu.Lock()
// 			defer mu.Unlock()
// 			x++
// 		}()
// 		// x++
// 	}

// 	time.Sleep(500 * time.Microsecond)
// 	fmt.Println("counter : ", x)
// }
