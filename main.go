// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go

package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
	"flag" // Imported 'flag' package
)


func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer) {
	defer wg.Done()
	maxRetries := 3
    for addr := range tasks {
		var success bool
		for i := range maxRetries {      
		conn, err := dialer.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			fmt.Printf("Connection to %s was successful\n", addr)
			success = true
			break
		}
		backoff := time.Duration(1<<i) * time.Second
		fmt.Printf("Attempt %d to %s failed. Waiting %v...\n", i+1,  addr, backoff)
		time.Sleep(backoff)
	    }
		if !success {
			fmt.Printf("Failed to connect to %s after %d attempts\n", addr, maxRetries)
		}
	}
}

func main() {

	var wg sync.WaitGroup
	tasks := make(chan string, 100)

	// Defines target flag
    var target = flag.String("target", "localhost", "Specify the target host")
	
	// Defines start port flag
	var startPort = flag.Int("startPort", 1, "Specify the start port")

	// Defines end port flag
	var endPort = flag.Int("endPort", 1024, "Specify the end port")

	// Reads & applies values
	flag.Parse()

	// Optional input validation added
	if *startPort > *endPort {
		fmt.Println("Error: startPort cannot be greater than endPort")
		return
	}

	dialer := net.Dialer {
		Timeout: 5 * time.Second,
	}
  
	workers := 100

    for i := 1; i <= workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, dialer)
	}

	for p := *startPort; p <= *endPort; p++ { // Dereferenced startPort and endPort
		port := strconv.Itoa(p)
        address := net.JoinHostPort(*target, port) // Dereferenced 'target'
		tasks <- address
	}
	close(tasks)
	wg.Wait()
}