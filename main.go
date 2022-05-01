package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"
)

type PortRange struct {
	Start int
	End   int
}

type Address struct {
	Name string
}

type NetworkProtocol struct {
	Type string
}

type PortResult struct {
	Number int
	State  string
}

type ScanDetails struct {
	Name            string
	Address         Address
	PortRange       PortRange
	NetworkProtocol NetworkProtocol
	ScanDateTime    string
	TimerStart      time.Time
	TimerResult     float64
	PortResult      []PortResult
}

func addDelay() {
	time.Sleep(time.Second / 10)
}

func getDateTime() string {
	dtNow := time.Now()
	dtString := dtNow.Format("2006-01-02 15:04:05 MST")
	return dtString
}

func (s ScanDetails) portSummary() {
	var openPortsCount int
	var openPorts []int
	var closedPortsCount int
	sort.Slice(s.PortResult, func(i, j int) bool {
		return s.PortResult[i].Number < s.PortResult[j].Number
	})
	for _, p := range s.PortResult {
		if bytes.Count([]byte(p.State), []byte("Open")) > 0 {
			openPortsCount++
			openPorts = append(openPorts, p.Number)
		} else if bytes.Count([]byte(p.State), []byte("Closed")) > 0 {
			closedPortsCount++
		}
	}
	fmt.Printf("[+] Scan summary: %d closed ports (conn-refused) %d open ports\n", closedPortsCount, openPortsCount)
	if openPortsCount > 0 {
		fmt.Printf("\n    PORT\tSTATE\n")
		for _, p := range openPorts {
			fmt.Printf("    %s/%d\t%s\n", s.NetworkProtocol.Type, p, "Open")
		}
	} else {
		fmt.Printf("[+] Host may be down or blocking all connections.\n")
	}
}

func (s ScanDetails) jsonResults() {
	sort.Slice(s.PortResult, func(i, j int) bool {
		return s.PortResult[i].Number < s.PortResult[j].Number
	})
	jsonResult, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		log.Printf("Error occurred with JSON formatting.")
	}
	fmt.Print(string(jsonResult))
}

func scanPortSlow(p int, s *ScanDetails) {
	scanString := fmt.Sprintf("%v:%v", s.Address.Name, p)
	fmt.Printf("[-] Scanning port: %d/%d\r", p, s.PortRange.End)
	addDelay()
	// conn, err := net.DialTimeout(s.NetworkProtocol.Type, scanString, (time.Second * 6))
	conn, err := net.Dial(s.NetworkProtocol.Type, scanString)
	if err != nil {
		s.PortResult = append(s.PortResult, PortResult{p, "Closed"})
		return
	} else {
		s.PortResult = append(s.PortResult, PortResult{p, "Open"})
	}
	conn.Close()
}

func scanPortWg(p int, wg *sync.WaitGroup, m *sync.Mutex, s *ScanDetails) {
	defer wg.Done()
	scanString := fmt.Sprintf("%v:%v", s.Address.Name, p)
	fmt.Printf("[-] Scanning port: %d/%d\r", p, s.PortRange.End)
	addDelay()
	conn, err := net.DialTimeout(s.NetworkProtocol.Type, scanString, (time.Second * 5))
	if err != nil {
		m.Lock()
		s.PortResult = append(s.PortResult, PortResult{p, "Closed"})
		m.Unlock()
		return
	} else {
		m.Lock()
		s.PortResult = append(s.PortResult, PortResult{p, "Open"})
		m.Unlock()
	}
	conn.Close()
}

func scanPortChan(workerId int, ports chan int, results chan PortResult, s *ScanDetails) {
	for p := range ports {
		scanString := fmt.Sprintf("%v:%v", s.Address.Name, p)
		fmt.Printf("[-] Scanning port: %d/%d\r", p, s.PortRange.End)
		addDelay()
		conn, err := net.DialTimeout(s.NetworkProtocol.Type, scanString, (time.Second * 6))
		if err != nil {
			results <- PortResult{p, "Closed"}
			continue
		}
		conn.Close()
		results <- PortResult{p, "Open"}
	}
}

func scannerOne(scanAddress Address, scanRange PortRange, scanProtocol NetworkProtocol) (s1 ScanDetails) {
	// s1 - No Concurrency
	s1.TimerStart = time.Now()
	s1.Name = "Non-Concurrency"
	s1.PortRange = scanRange
	s1.Address = scanAddress
	s1.NetworkProtocol = scanProtocol
	s1.ScanDateTime = getDateTime()

	fmt.Printf("\n[*] Starting Go PortScan at %s\n", s1.ScanDateTime)
	fmt.Printf("[+] Scan method: %s\n", s1.Name)
	fmt.Printf("[+] Target address: %s\n", s1.Address.Name)
	fmt.Printf("[+] Port range: %d-%d\n", s1.PortRange.Start, s1.PortRange.End)
	for i := scanRange.Start; i <= scanRange.End; i++ {
		scanPortSlow(i, &s1)
	}
	s1.TimerResult = time.Now().Sub(s1.TimerStart).Seconds()
	fmt.Printf("\n\033[1A\033[K[*] Scan done: 1 IP address scanned in %.2f seconds.\n", s1.TimerResult)
	return
}

func scannerTwo(scanAddress Address, scanRange PortRange, scanProtocol NetworkProtocol) (s2 ScanDetails) {
	// s2 - Concurrency using WaitGroups
	var wg sync.WaitGroup
	var m sync.Mutex
	s2.TimerStart = time.Now()
	s2.Name = "Concurrency using WaitGroups"
	s2.PortRange = scanRange
	s2.Address = scanAddress
	s2.NetworkProtocol = scanProtocol
	s2.ScanDateTime = getDateTime()

	fmt.Printf("\n[*] Starting Go PortScan at %s\n", s2.ScanDateTime)
	fmt.Printf("[+] Scan method: %s\n", s2.Name)
	fmt.Printf("[+] Target address: %s\n", s2.Address.Name)
	fmt.Printf("[+] Port range: %d-%d\n", s2.PortRange.Start, s2.PortRange.End)
	for i := scanRange.Start; i <= scanRange.End; i++ {
		wg.Add(1)
		go scanPortWg(i, &wg, &m, &s2)
	}
	wg.Wait()
	s2.TimerResult = time.Now().Sub(s2.TimerStart).Seconds()
	fmt.Printf("\n\033[1A\033[K[*] Scan done: 1 IP address scanned in %.2f seconds.\n", s2.TimerResult)
	return
}

func scannerThree(scanAddress Address, scanRange PortRange, scanProtocol NetworkProtocol) (s3 ScanDetails) {
	// Concurrency using channels and worker pools
	s3.TimerStart = time.Now()
	s3.Name = "Concurrency using Channels and Worker Pools"
	s3.PortRange = scanRange
	s3.Address = scanAddress
	s3.NetworkProtocol = scanProtocol
	s3.ScanDateTime = getDateTime()

	fmt.Printf("\n[*] Starting Go PortScan at %s\n", s3.ScanDateTime)
	fmt.Printf("[+] Scan method: %s\n", s3.Name)
	fmt.Printf("[+] Target address: %s\n", s3.Address.Name)
	fmt.Printf("[+] Port range: %d-%d\n", s3.PortRange.Start, s3.PortRange.End)
	ports := make(chan int, 100)
	results := make(chan PortResult)

	// Worker pool size based on ports channel buffer set above
	for w := 1; w < cap(ports); w++ {
		go scanPortChan(w, ports, results, &s3)
	}

	go func() {
		for i := scanRange.Start; i <= scanRange.End; i++ {
			ports <- i
		}
	}()

	for i := scanRange.Start; i <= scanRange.End; i++ {
		port := <-results
		if port.State == "Open" {
			s3.PortResult = append(s3.PortResult, PortResult{port.Number, port.State})
		} else {
			s3.PortResult = append(s3.PortResult, PortResult{port.Number, port.State})
		}
	}

	close(ports)
	close(results)
	s3.TimerResult = time.Now().Sub(s3.TimerStart).Seconds()
	fmt.Printf("\n\033[1A\033[K[*] Scan done: 1 IP address scanned in %.2f seconds.\n", s3.TimerResult)
	return
}

func main() {
	fmt.Printf("\n[ Go Concurrency Port Scanning Examples ]\n")

	// // Set Scan Settings
	scanAddress := Address{"127.0.0.1"}
	scanRange := PortRange{21, 100}
	scanProtocol := NetworkProtocol{"tcp"}

	// Initiate Scanner One - No Concurrency
	s1 := scannerOne(scanAddress, scanRange, scanProtocol)
	s1.portSummary()

	// Initiate Scanner Two - Concurrency using WaitGroups
	s2 := scannerTwo(scanAddress, scanRange, scanProtocol)
	s2.portSummary()

	// Initiate Scanner Three - Concurrency using Channels
	s3 := scannerThree(scanAddress, scanRange, scanProtocol)
	s3.portSummary()

	// s1.jsonResults()
	// s2.jsonResults()
	// s3.jsonResults()
}
