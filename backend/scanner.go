package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type DiscoveredDevice struct {
	IP        string `json:"ip"`
	Hostname  string `json:"hostname"`
	MAC       string `json:"mac"`
	Vendor    string `json:"vendor"`
	OpenPorts []int  `json:"open_ports"`
}

var ouiLookup = map[string]string{
	"00:01:42": "Cisco",
	"00:0C:29": "VMware",
	"00:11:32": "Synology",
	"00:1E:06": "WIBO",
	"00:24:14": "Cisco",
	"00:25:9C": "Cisco",
	"00:50:56": "VMware",
	"04:D4:C4": "Raspberry Pi",
	"08:00:27": "VirtualBox",
	"10:27:F5": "Ubiquiti",
	"18:E8:29": "Ubiquiti",
	"28:EE:52": "Ubiquiti",
	"3C:3B:4D": "Apple",
	"44:D9:E7": "Ubiquiti",
	"48:E7:29": "Ubiquiti",
	"50:E5:49": "Giga-Byte",
	"54:E1:AD": "Cisco",
	"5C:F9:DD": "Dell",
	"60:22:32": "Ubiquiti",
	"68:D7:9A": "Ubiquiti",
	"70:A7:41": "Ubiquiti",
	"74:83:C2": "Ubiquiti",
	"80:2A:A8": "Ubiquiti",
	"A4:2B:B0": "Apple",
	"B8:27:EB": "Raspberry Pi",
	"D4:CA:6D": "Raspberry Pi",
	"DC:A6:32": "Raspberry Pi",
	"E0:63:DA": "Ubiquiti",
	"E4:5F:01": "Raspberry Pi",
	"F0:9F:C2": "Ubiquiti",
	"FC:EC:DA": "Ubiquiti",
}

func getVendorFromMAC(mac string) string {
	if len(mac) >= 8 {
		oui := strings.ToUpper(mac[:8])
		if vendor, ok := ouiLookup[oui]; ok {
			return vendor
		}
	}
	return "Unknown"
}

func getLocalSubnet() (*net.IPNet, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// Avoid returning Docker/bridge networks if possible, prefer 192.168, 10.x, 172.16.x
				if ipnet.IP.IsPrivate() {
					return ipnet, nil
				}
			}
		}
	}
	// Fallback to the first non-loopback IPv4 if no private IP is found
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet, nil
			}
		}
	}
	return nil, fmt.Errorf("no suitable local subnet found")
}

func getArpTable() map[string]string {
	arpMap := make(map[string]string)
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return arpMap
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Skip header
	scanner.Scan()
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 4 {
			ip := fields[0]
			mac := fields[3]
			if mac != "00:00:00:00:00:00" {
				arpMap[ip] = mac
			}
		}
	}
	return arpMap
}

func generateIPs(network *net.IPNet) []string {
	var ips []string

	// Ensure we only use the network address part for iteration
	ip := network.IP.Mask(network.Mask)

	for network.Contains(ip) {
		// Skip network and broadcast addresses (heuristic for /24)
		if ip[len(ip)-1] != 0 && ip[len(ip)-1] != 255 {
			// create a copy to append
			ipCopy := make(net.IP, len(ip))
			copy(ipCopy, ip)
			ips = append(ips, ipCopy.String())
		}
		inc(ip)
	}
	return ips
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func scanPort(ip string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		conn.Close()
		return true
	}
	return false
}

func handleNetworkScan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	subnet, err := getLocalSubnet()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ips := generateIPs(subnet)
	portsToScan := []int{80, 443, 502, 8080}
	timeout := 500 * time.Millisecond // Use a short timeout for the sweep

	var wg sync.WaitGroup
	var mu sync.Mutex
	devices := make(map[string]*DiscoveredDevice)

	// Scan in chunks to avoid overwhelming the network
	maxConcurrent := 100
	sem := make(chan struct{}, maxConcurrent)

	for _, ip := range ips {
		for _, port := range portsToScan {
			wg.Add(1)
			go func(ip string, port int) {
				defer wg.Done()
				sem <- struct{}{}

				isOpen := scanPort(ip, port, timeout)

				<-sem
				if isOpen {
					mu.Lock()
					if _, exists := devices[ip]; !exists {
						devices[ip] = &DiscoveredDevice{
							IP:        ip,
							OpenPorts: []int{},
						}
					}
					devices[ip].OpenPorts = append(devices[ip].OpenPorts, port)
					mu.Unlock()
				}
			}(ip, port)
		}
	}

	wg.Wait()

	// Gather ARP info
	arpTable := getArpTable()

	// Fill in hostnames and MACs
	var results []DiscoveredDevice
	for ip, dev := range devices {
		if mac, ok := arpTable[ip]; ok {
			dev.MAC = mac
			dev.Vendor = getVendorFromMAC(mac)
		} else {
			dev.MAC = "Unknown"
			dev.Vendor = "Unknown"
		}

		names, err := net.LookupAddr(ip)
		if err == nil && len(names) > 0 {
			dev.Hostname = strings.TrimSuffix(names[0], ".")
		} else {
			dev.Hostname = "Unknown"
		}
		results = append(results, *dev)
	}

	json.NewEncoder(w).Encode(results)
}
