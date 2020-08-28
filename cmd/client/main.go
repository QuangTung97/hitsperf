package main

import (
	"context"
	"fmt"
	"hitsperf"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:6000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := hitsperf.NewIncreaseServiceClient(conn)

	begin := time.Now()

	var wg sync.WaitGroup
	wg.Add(1000)

	var mut sync.Mutex
	var d time.Duration = 0

	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()

			begin := time.Now()
			req := &hitsperf.IncRequest{
				Value: 5,
			}
			_, err = client.Inc(context.Background(), req)
			if err != nil {
				panic(err)
			}
			end := time.Now()
			mut.Lock()
			d += end.Sub(begin)
			mut.Unlock()
		}()
	}
	wg.Wait()

	fmt.Println(time.Now().Sub(begin).Nanoseconds())
	fmt.Println("AVG request time: ", d.Nanoseconds()/1000.0)
}
