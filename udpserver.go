package main

import (
	"context"
	"fmt"
	"log"
	"net"
)

type UDPDatagram struct {
	RemoteAddr *net.UDPAddr
	Data       []byte
}

func ServeUDP(ctx context.Context, port int32) error {
	addr := net.UDPAddr{Port: int(port), IP: net.ParseIP("0.0.0.0")}
	listener, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer listener.Close()
	log.Printf("Listening UDPServer on %s", addr.String())

	var (
		dataChan = make(chan UDPDatagram, 1024)
		errChan  = make(chan error)
	)

	go func() {
		for {
			b := make([]byte, 4096)
			_, remote, err := listener.ReadFromUDP(b)
			if err != nil {
				errChan <- err
				return
			} else {
				dataChan <- UDPDatagram{Data: b, RemoteAddr: remote}
			}
		}
	}()

	for {
		select {
		case data := <-dataChan:
			if _, err := listener.WriteToUDP(data.Data, data.RemoteAddr); err != nil {
				log.Printf("Failed to write bytes to UDP destination: %s", err)
			}
		case err := <-errChan:
			log.Printf("Failed to read from UDP: %s, closing", err)
			return fmt.Errorf("read from UDP: %w", err)
		case <-ctx.Done():
			return nil
		}

	}
}
