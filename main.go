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
	"sync/atomic" // For counting safely from goroutines
)

// worker attempts to connect to each address sent over the 'tasks' channel.
// It retries a few times with exponential backoff and counts successful connections using atomic increment.
func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer, openPorts *int32) {
	defer wg.Done()

	maxRetries := 3

    for addr := range tasks {
		var success bool

		// Retry up to maxRetries with exponential backoff
		for i := range maxRetries {      
		conn, err := dialer.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			fmt.Printf("Connection to %s was successful\n", addr)
			success = true
			atomic.AddInt32(openPorts, 1) // Safely increments shared openPorts counter
			break // Exits the retry loop once successful
		}

		// Wait with exponential backoff before retrying
		backoff := time.Duration(1<<i) * time.Second
		fmt.Printf("Attempt %d to %s failed. Waiting %v...\n", i+1,  addr, backoff)
		time.Sleep(backoff)
	    }

		// Logs failure if all retries failed
		if !success {
			fmt.Printf("Failed to connect to %s after %d attempts\n", addr, maxRetries)
		}
	}
}

func main() {
	// Counter for open ports, shared across goroutines
	var openPorts int32 = 0

	// Records the start time
	start := time.Now()

	var wg sync.WaitGroup
	tasks := make(chan string, 100) // Buffered channel to hold addresses to scan

	// Command-line flags for user-configurable options
    var target = flag.String("target", "localhost", "Specify the target host")
	var startPort = flag.Int("startPort", 1, "Specify the start port")
	var endPort = flag.Int("endPort", 1024, "Specify the end port")
	var workers = flag.Int("workers", 100, "Specify the amount of workers")
	var timeout = flag.Int("timeout", 5, "Timeout in seconds for each connection")

	// Parse command-line input
	flag.Parse()

	// Validate port range
	if *startPort > *endPort {
		fmt.Println("Error: startPort cannot be greater than endPort")
		return
	}

	// Set up the network dialer with specified timeout
	dialer := net.Dialer {
		Timeout: time.Duration(*timeout) * time.Second,
	}

	// Launch worker goroutine
    for i := 1; i <= *workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, dialer, &openPorts)
	}

	// Send target:port combinations to the tasks channel
	for p := *startPort; p <= *endPort; p++ {
		port := strconv.Itoa(p)
        address := net.JoinHostPort(*target, port)
		tasks <- address
	}

	close(tasks) // Closes the task channel to signal no more work
	wg.Wait() // Wait for all workers to finish

	//Records the end time
	duration := time.Since(start)

	// Scan Summary
	fmt.Printf("\n--- Scan Summary ---\n")
	fmt.Printf("Open ports: %d\n", openPorts)
	fmt.Printf("Total ports scanned: %d\n", *endPort - *startPort + 1)
	fmt.Printf("Time taken: %s\n", duration)
}