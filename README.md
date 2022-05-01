# goportscan
Go Concurrency Port Scanning Examples. This was only created as a self learning tool to better understand how to use different concurrency methods in Go. 

Scanner 1. - No concurrency  
Scanner 2. - Concurrency using WaitGroups (requires mutex to prevent race condition involving results output)  
Scanner 3. - Concurrency using Channels and Work Pools (work pool size currently based on channel buffer 100)

#### Issues / Notes
- Function `addDelay()` used to slow down scanning and allow viewing on some output from scanners.
- Does not currently support commandline arguments. Manually edit values in function `main()`.
- Currently only working with TCP
- No tests
- Will need to have services running if scanning your own localhost (127.0.0.1) or try using address scanme.nmap.org.

If testing on localhost (127.0.0.1) which has python installed you can mock / simulate some open ports using SimpleHTTPServer. Run the following command through the terminal or command line:
```terminal
python -m SimpleHTTPServer 22 > /dev/null 2>&1 &
python -m SimpleHTTPServer 53 > /dev/null 2>&1 &
python -m SimpleHTTPServer 80 > /dev/null 2>&1 &
```
The above command will start the processes in the background on ports 22, 53 and 80 to simulate SSH, DNS and HTTP ports are open.  
  
To end (kill) all these SimpleHTTPServer processes use:
```terminal
kill -9 `ps -ef | grep SimpleHTTPServer | awk '{print $2}'`
```
#### Sample Output
```
[ Go Concurrency Port Scanning Examples ]

[*] Starting Non-Concurrency tcp scan of 127.0.0.1...
[+] Scan Completed.
[+] Scan Runtime: 8.212106205s
[+] Open Ports: 3 [22 53 80]
[+] Closed Ports: 77

[*] Starting Concurrency using WaitGroups tcp scan of 127.0.0.1...
[+] Scan Completed.
[+] Scan Runtime: 104.319974ms
[+] Open Ports: 3 [22 53 80]
[+] Closed Ports: 77

[*] Starting Concurrency using channels and worker pools tcp scan of 127.0.0.1...
[+] Scan Completed.
[+] Scan Runtime: 108.778132ms
[+] Open Ports: 3 [22 53 80]
[+] Closed Ports: 77
```
