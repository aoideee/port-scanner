// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go

package main

import (
	"flag" // Imported 'flag' package
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic" // For counting safely from goroutines
	"time"
)


func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer, openPorts *int32) {
	defer wg.Done()
	maxRetries := 3
    for addr := range tasks {
		var success bool
		for i := range maxRetries {      
		conn, err := dialer.Dial("tcp", addr)
		if err == nil {
			defer conn.Close()
			fmt.Printf("Connection to %s was successful\n", addr)
			success = true
			atomic.AddInt32(openPorts, 1) // Increments counter

			// Read banner
			buffer := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, err := conn.Read(buffer)
			if err == nil && n > 0 {
				fmt.Printf("Banner from %s: %s\n", addr, string(buffer[:n]))
			} else {
				fmt.Printf("No banner received from %s\n", addr)
			}
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
	// Declared and initialized counter
	var openPorts int32 = 0

	// Records the start time
	start := time.Now()

	var wg sync.WaitGroup
	tasks := make(chan string, 100)

	// Defines target flag
    var target = flag.String("target", "localhost", "Specify the target host")
	
	// Defines start port flag
	var startPort = flag.Int("startPort", 1, "Specify the start port")

	// Defines end port flag
	var endPort = flag.Int("endPort", 1024, "Specify the end port")

	// Defines workers flag
	var workers = flag.Int("workers", 100, "Specify the amount of workers")

	// Defines timeout flag
	var timeout = flag.Int("timeout", 5, "Timeout in seconds for each connection")


	// Reads & applies values
	flag.Parse()

	// Optional input validation added
	if *startPort > *endPort {
		fmt.Println("Error: startPort cannot be greater than endPort")
		return
	}

	// Updated Dialer
	dialer := net.Dialer {
		Timeout: time.Duration(*timeout) * time.Second, // Dereferenced 'timeout'
	}

    for i := 1; i <= *workers; i++ { // Dereferenced workers
		wg.Add(1)
		go worker(&wg, tasks, dialer, &openPorts) // Referenced openPorts
	}

	for p := *startPort; p <= *endPort; p++ { // Dereferenced startPort and endPort
		fmt.Printf("Scanning port %d/%d\n", p, *endPort) // Progress indicator
		port := strconv.Itoa(p)
        address := net.JoinHostPort(*target, port) // Dereferenced 'target'
		tasks <- address
	}
	close(tasks)
	wg.Wait()

	//Records the end time
	duration := time.Since(start)

	// Scan Summary
	fmt.Printf("\n--- Scan Summary ---\n")
	fmt.Printf("Open ports: %d\n", openPorts)
	fmt.Printf("Total ports scanned: %d\n", *endPort - *startPort + 1)
	fmt.Printf("Time taken: %s\n", duration)
}