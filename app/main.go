package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/server"
)

func main() {
	s := server.NewServer("0.0.0.0", 6379)

	var wg sync.WaitGroup
	err := s.Start(&wg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	wg.Wait()
	fmt.Println("all connections closed, shutting down...")
}
