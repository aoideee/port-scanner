// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go

package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"strings" // For splitting  CSV inputs (like -targets and -ports)
	"encoding/json" // For optional JSON output
)

// ScanResult defines the structure used to store results (used in JSON output)
type ScanResult struct {
	Target string `json:"target"`             // Target hostname or IP
	Port int `json:"port"`                    // Open port number
	Banner string `json:"banner,omitempty"` // Optional banner string (if it's present)
}

// worker is a goroutine function that attempts to connect to provided address,
// performs banner grabbing, and saves results thread-safely.
func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer, openPorts *int32, results *[]ScanResult, resultsMutex *sync.Mutex) {
	defer wg.Done()

	maxRetries := 3

    for addr := range tasks {
		var success bool

		// Retry connection up to maxRetries with exponential backoff
		for i := range maxRetries {      
		conn, err := dialer.Dial("tcp", addr)
		if err == nil {
			defer conn.Close()
			fmt.Printf("Connection to %s was successful\n", addr)
			success = true
			atomic.AddInt32(openPorts, 1) // Safely increment open port counter

			// Attempt to grab the banner (initial message from server)
			buffer := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // Avoid hanging forever
			n, err := conn.Read(buffer)
			if err == nil && n > 0 {
				fmt.Printf("Banner from %s: %s\n", addr, string(buffer[:n]))
			} else {
				fmt.Printf("No banner received from %s\n", addr)
			}

			banner := ""
			if err == nil && n > 0 {
				banner = string(buffer[:n])
				fmt.Printf("Banner from %s: %s\n", addr, banner)
			} else {
				fmt.Printf("No banner received from %s\n", addr)
			}

			// Split host:port and store result safely
			host, portStr, _ := net.SplitHostPort(addr)
			port, _ := strconv.Atoi(portStr)

			// Safely appends to shared results slice using a mutex
			resultsMutex.Lock()
			*results = append(*results, ScanResult{
				Target: host,
				Port: port,
				Banner: banner,
			})
			resultsMutex.Unlock()

			break // Exits loop after success
		}

		// Wait with exponential backoff before retrying
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
	// Counter for open ports, shared across goroutines
	var openPorts int32 = 0

	// Records the start time
	start := time.Now()

	var wg sync.WaitGroup
	tasks := make(chan string, 100) // Buffered channel to hold addresses to scan

	// Command-line flags for user-configurable options
    var targets = flag.String("targets", "localhost", "Comma-separated list of targets")
	var startPort = flag.Int("startPort", 1, "Specify the start port")
	var endPort = flag.Int("endPort", 1024, "Specify the end port")
	var workers = flag.Int("workers", 100, "Specify the amount of workers")
	var timeout = flag.Int("timeout", 5, "Timeout in seconds for each connection")
	var jsonOutput = flag.Bool("json", false, "Output results in JSON format")
	var specificPorts = flag.String("ports", "", "Comma-separated list of specific ports to scan (e.g. 22,8,443)")

	// Parse command-line input
	flag.Parse()

	// Split target list
	targetList := strings.Split(*targets, ",")

	// Parse specific ports if provided
	var portList []int
	if *specificPorts != "" {
		portStrs := strings.Split(*specificPorts, ",")
		for _, ps := range portStrs {
			port, err := strconv.Atoi(strings.TrimSpace(ps))
			if err != nil {
				fmt.Printf("Invalid port: %s\n", ps)
				return
			}
			portList = append(portList, port)
		}
	}

	// Validate port range
	if *startPort > *endPort {
		fmt.Println("Error: startPort cannot be greater than endPort")
		return
	}

	// Set up the network dialer with specified timeout
	dialer := net.Dialer {
		Timeout: time.Duration(*timeout) * time.Second, // Dereferenced 'timeout'
	}

	// Slice to store scan results from all workers
	var results []ScanResult

	// Mutex to protect concurrent writes to results slice
	var resultsMutex sync.Mutex

	// Launch worker goroutines and pass pointers to results and mutex
    for i := 1; i <= *workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, dialer, &openPorts, &results, &resultsMutex) // Updated referenced openPorts
	}

	// Distribute tasks to workers
	for _, tgt := range targetList {
		if len(portList) > 0 {

			// If user specified specific ports
			for _, p := range portList {
				fmt.Printf("Scanning %s: %d\n", tgt, p)
				port := strconv.Itoa(p)
				address := net.JoinHostPort(tgt, port)
				tasks <- address
			}
		} else {
			// Scan full port range
			for p := *startPort; p <= *endPort; p++ { // Dereferenced startPort and endPort
			fmt.Printf("Scanning %s: %d/%d\n", tgt, p, *endPort) // Progress indicator
			port := strconv.Itoa(p)
			address := net.JoinHostPort(tgt, port) // Dereferenced 'target'
			tasks <- address
			}
		}
		
	}
	
	close(tasks) // Closes the task channel to signal no more work
	wg.Wait() // Wait for all workers to finish

	//Records the end time
	duration := time.Since(start)

	// Output scan results in JSON if flag is set
	if *jsonOutput {
		fmt.Println("\n--- JSON Output ---")
		jsonData, err := json.MarshalIndent(results, "", "  ") // Pretty-print with indent
		if err != nil {
			fmt.Println("Error encoding JSON:", err)
		} else {
			fmt.Println(string(jsonData))
		}
	} else {
		// Default text summary output
		fmt.Printf("\n--- Scan Summary ---\n")
		fmt.Printf("Open ports: %d\n", openPorts)
		fmt.Printf("Total ports scanned: %d\n", *endPort - *startPort + 1)
		fmt.Printf("Time taken: %s\n", duration)
	}
	
}