# Go TCP Port Scanner

This project is a command-line TCP port scanner written in Go. It demonstrates the use of concurrency with goroutines, synchronization with WaitGroups and mutexes, and customizable flags for scanning flexibility.

## 📌 Part A: Basic Scanner

This version scans a range of ports on a single target using multiple goroutines. It includes retry logic and a summary report of open ports.

### ✅ Features
- Custom target (`-target`)
- Port range (`-startPort`, `-endPort`)
- Configurable number of workers (`-workers`)
- Timeout setting (`-timeout`)
- Open port counter and summary

### 🔧 How to Build and Run (Part A)

```bash
cd part-a    # or the branch/folder containing Part A
go run main.go -target=scanme.nmap.org -startPort=20 -endPort=80 -workers=100 -timeout=3

## Sample Output (Part A)
`
Connection to scanme.nmap.org:22 was successful
Failed to connect to scanme.nmap.org:23 after 3 attempts
...

--- Scan Summary ---
Open ports: 1
Total ports scanned: 61
Time taken: 3.123456s
`
