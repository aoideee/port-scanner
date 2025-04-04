## 🚀 Part B: Enriched Scanner

This version adds new features for multi-host scanning, banner grabbing, JSON output, and specific port selection.

### ✅ Features
- Everything from Part A
- Scan multiple hosts (`-targets=a.com,b.com`)
- Banner grabbing from open ports
- JSON output with `-json`
- Scan only specific ports with `-ports=22,80,443`

### 🔧 How to Build and Run (Part B)

```bash
cd part-b    # or switch to the part-b branch
go run main.go -targets=scanme.nmap.org,example.com -ports=22,80 -timeout=2 -workers=50 -json
```

### Sample Output (Part B - JSON)
```json
[
  {
    "target": "scanme.nmap.org",
    "port": 22,
    "banner": "SSH-2.0-OpenSSH_7.9p1 Debian-10+deb10u2\r\n"
  },
  {
    "target": "example.com",
    "port": 80,
    "banner": "HTTP/1.1 200 OK\r\nServer: Apache\r\n..."
  }
]
```

### Sample Output (Part B - Summary)
```bash
Connection to scanme.nmap.org:22 was successful
No banner received from scanme.nmap.org:22

--- Scan Summary ---
Open ports: 1
Total ports scanned: 2
Time taken: 1.998s
```
