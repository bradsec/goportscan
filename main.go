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
	TimerStart      time.Time
	TimerResult     time.Duration
	PortResult      []PortResult
}

func addDelay() {
	time.Sleep(time.Second / 10)
}

func (s ScanDetails) runTime() {
	fmt.Printf("[+] Scan Runtime: %s\n", s.TimerResult)
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
	fmt.Printf("[+] Open Ports: %d %v\n[+] Closed Ports: %d\n", openPortsCount, openPorts, closedPortsCount)
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
	fmt.Printf("\n[*] Starting %s %s scan of %s...\n", s1.Name, s1.NetworkProtocol.Type, s1.Address.Name)
	for i := scanRange.Start; i <= scanRange.End; i++ {
		scanPortSlow(i, &s1)
	}
	s1.TimerResult = time.Since(s1.TimerStart)
	fmt.Printf("\n\033[1A\033[K[+] Scan Completed.\n")
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
	fmt.Printf("\n[*] Starting %s %s scan of %s...\n", s2.Name, s2.NetworkProtocol.Type, s2.Address.Name)
	for i := scanRange.Start; i <= scanRange.End; i++ {
		wg.Add(1)
		go scanPortWg(i, &wg, &m, &s2)
	}
	wg.Wait()
	s2.TimerResult = time.Since(s2.TimerStart)
	fmt.Printf("\n\033[1A\033[K[+] Scan Completed.\n")
	return
}

func scannerThree(scanAddress Address, scanRange PortRange, scanProtocol NetworkProtocol) (s3 ScanDetails) {
	// Concurrency using channels and worker pools
	s3.TimerStart = time.Now()
	s3.Name = "Concurrency using channels and worker pools"
	s3.PortRange = scanRange
	s3.Address = scanAddress
	s3.NetworkProtocol = scanProtocol
	fmt.Printf("\n[*] Starting %s %s scan of %s...\n", s3.Name, s3.NetworkProtocol.Type, s3.Address.Name)
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
	s3.TimerResult = time.Since(s3.TimerStart)
	fmt.Printf("\n\033[1A\033[K[+] Scan Completed.\n")
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
	s1.runTime()
	s1.portSummary()

	// Initiate Scanner Two - Concurrency using WaitGroups
	s2 := scannerTwo(scanAddress, scanRange, scanProtocol)
	s2.runTime()
	s2.portSummary()

	// Initiate Scanner Three - Concurrency using Channels
	s3 := scannerThree(scanAddress, scanRange, scanProtocol)
	s3.runTime()
	s3.portSummary()

	// s1.jsonResults()
	// s2.jsonResults()
	// s3.jsonResults()
}
