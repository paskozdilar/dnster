package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Parse command line arguments
	var conf, addr string
	var cachesize int
	flag.StringVar(&conf, "conf", "dnster.conf", "file containing upstream DNS servers to use")
	flag.StringVar(&addr, "addr", "127.0.0.53:53", "address and port to listen on")
	flag.IntVar(&cachesize, "cachesize", 1024, "LRU cache size, 0 means disabled")
	flag.Parse()

	// Read the resolv.conf file
	upstreamServers := []string{}
	{
		data, err := os.ReadFile(conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading '%s': %v\n", conf, err)
			os.Exit(1)
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line := strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if net.ParseIP(line) == nil {
				fmt.Fprintf(os.Stderr, "Invalid IP address in resolv.conf: %s\n", line)
				os.Exit(1)
			}
			upstreamServers = append(upstreamServers, line)
		}
	}

	if len(upstreamServers) == 0 {
		fmt.Fprint(os.Stderr, "List of upstream DNS servers empty\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Upstream DNS servers: %v\n", strings.Join(upstreamServers, ", "))

	// Start the DNS server
	if err := DNSServer(addr, upstreamServers, cachesize); err != nil {
		fmt.Fprintf(os.Stderr, "DNS server error: %v\n", err)
		os.Exit(1)
	}
}
