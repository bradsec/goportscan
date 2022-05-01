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

[*] Starting Go PortScan at 2022-05-01 11:55:51 AEST
[+] Scan method: Non-Concurrency
[+] Target address: 127.0.0.1
[+] Port range: 21-100
[*] Scan done: 1 IP address scanned in 8.23 seconds.
[+] Scan summary: 77 closed ports (conn-refused) 3 open ports

    PORT        STATE
    tcp/22      Open
    tcp/53      Open
    tcp/80      Open

[*] Starting Go PortScan at 2022-05-01 11:55:59 AEST
[+] Scan method: Concurrency using WaitGroups
[+] Target address: 127.0.0.1
[+] Port range: 21-100
[*] Scan done: 1 IP address scanned in 0.11 seconds.
[+] Scan summary: 77 closed ports (conn-refused) 3 open ports

    PORT        STATE
    tcp/22      Open
    tcp/53      Open
    tcp/80      Open

[*] Starting Go PortScan at 2022-05-01 11:55:59 AEST
[+] Scan method: Concurrency using Channels and Worker Pools
[+] Target address: 127.0.0.1
[+] Port range: 21-100
[*] Scan done: 1 IP address scanned in 0.11 seconds.
[+] Scan summary: 77 closed ports (conn-refused) 3 open ports

    PORT        STATE
    tcp/22      Open
    tcp/53      Open
    tcp/80      Open
```
