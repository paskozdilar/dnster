package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// DNSClient sends a DNS query and returns the response.
// The context is used to cancel the DNS query.
func DNSClient(ctx context.Context, server string, query *dnsmessage.Message) (*dnsmessage.Message, error) {
	// Marshal the DNS query message to bytes
	queryBytes, err := query.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack DNS query: %w", err)
	}

	// Dial UDP connection to the DNS server
	addr := &net.UDPAddr{IP: net.ParseIP(server), Port: 53}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP: %w", err)
	}
	defer conn.Close()

	// Create a channel to signal the completion of the DNS query
	done := make(chan struct{})
	var response *dnsmessage.Message

	go func() {
		defer close(done)

		// Send the DNS query
		_, err = conn.Write(queryBytes)
		if err != nil {
			return
		}

		// Set a read deadline
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// Read the DNS response
		responseBytes := make([]byte, 1024)
		n, err := conn.Read(responseBytes)
		if err != nil {
			return
		}

		// Unpack the DNS response
		var res dnsmessage.Message
		err = res.Unpack(responseBytes[:n])
		if err != nil {
			return
		}

		response = &res
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		if response == nil {
			return nil, fmt.Errorf("failed to receive DNS response")
		}
		return response, nil
	}
}
