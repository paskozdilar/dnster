package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// DNSServer listens for DNS queries and forwards them to multiple upstream servers.
func DNSServer(addr string, upstreamServers []string, cachesize int) error {
	// Resolve listen address
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("resolve udp addr: %w", err)
	}

	// Create a UDP connection
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	fmt.Fprintf(os.Stderr, "DNS server listening on: %v\n", addr)

	buffer := make([]byte, 1024)

	for {
		// Read a packet
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}
		if n < 12 {
			continue
		}

		// Parse the DNS query
		var query dnsmessage.Message
		err = query.Unpack(buffer[:n])
		if err != nil {
			if DEBUG {
				log.Println("query unpack error:", err)
			}
			continue
		}

		go func() {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			// Channel to receive responses
			responseChan := make(chan *dnsmessage.Message, len(upstreamServers))
			errorChan := make(chan error, len(upstreamServers))

			// Send the query to multiple upstream servers in parallel
			for _, server := range upstreamServers {
				go func(server string) {
					resp, err := DNSClient(ctx, server, &query)
					if err != nil {
						if DEBUG {
							log.Println("query", query.ID, "to", server, "error:", err)
						}
						errorChan <- err
						return
					}
					responseChan <- resp
				}(server)
			}

			// Wait for the first successful response or context cancellation
			var response *dnsmessage.Message
			errorCount := 0

		INNER_LOOP:
			for errorCount < len(upstreamServers) {
				select {
				case response = <-responseChan:
					if len(response.Answers) == 0 {
						errorCount += 1
						if errorCount == len(upstreamServers) {
							cancel()
							break INNER_LOOP
						}
						continue INNER_LOOP
					}
					cancel()
					break INNER_LOOP
				case <-errorChan:
					errorCount += 1
					if errorCount == len(upstreamServers) {
						cancel()
						return
					}
					continue INNER_LOOP
				case <-ctx.Done():
					return
				}
			}

			if errorCount == len(upstreamServers) {
				if DEBUG {
					log.Println("query", query.ID, "all error responses")
				}
				response = &dnsmessage.Message{
					Header: dnsmessage.Header{
						Response:      true,
						Authoritative: true,
						RCode:         dnsmessage.RCodeServerFailure,
						ID:            query.ID,
					},
					Questions: query.Questions,
				}
			} else {
				if DEBUG {
					log.Println("query", query.ID, "got response:", query.Answers)
				}
			}

			// Marshal the response to bytes
			responseBytes, err := response.Pack()
			if err != nil {
				return
			}

			// Send the response back to the client
			_, err = conn.WriteToUDP(responseBytes, clientAddr)
			if err != nil {
				return
			}
		}()
	}
}
