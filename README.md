# goportscan
Go Concurrency Port Scanning Examples. This was only created as a self learning tool to better understand how to use different concurrency methods in Go. 

Scanner 1. - No concurrency  
Scanner 2. - Concurrency using WaitGroups (requires mutex to prevent race condition involving results output)  
Scanner 3. - Concurrency using Channels and Work Pools (work pool size currently based on channel buffer 100)

#### Issues / Notes
- No tests
- Function `addDelay()` used to slow down scanning and allow viewing on some output from scanners.
- Not currently supporting commandline arguments. Need to manually edit values in function `main()`.
- Will need to have services running if scanner localhost or try using scanme.nmap.org
- Not currently working with UDP protocol, will just response all ports are open.

#### Sample Output
```
[ Go Concurrency Port Scanning Examples ]

[*] Starting Non-Concurrency tcp scan of 127.0.0.1...
[+] Scan Completed.
[+] Scan Runtime: 8.212106205s
[+] Open Ports: 2 [53 80]
[+] Closed Ports: 78

[*] Starting Concurrency using WaitGroups tcp scan of 127.0.0.1...
[+] Scan Completed.
[+] Scan Runtime: 104.319974ms
[+] Open Ports: 2 [53 80]
[+] Closed Ports: 78

[*] Starting Concurrency using channels and worker pools tcp scan of 127.0.0.1...
[+] Scan Completed.
[+] Scan Runtime: 108.778132ms
[+] Open Ports: 2 [53 80]
[+] Closed Ports: 78
```
