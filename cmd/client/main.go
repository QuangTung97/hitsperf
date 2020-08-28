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
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		go func() {
			defer wg.Done()

			req := &hitsperf.IncRequest{
				Value: 5,
			}
			_, err = client.Inc(context.Background(), req)
			if err != nil {
				panic(err)
			}
		}()
	}
	wg.Wait()

	fmt.Println(time.Now().Sub(begin).Nanoseconds())
}
