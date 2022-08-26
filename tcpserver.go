package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
)

func HandleTCPConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF {
				log.Printf("Failed to read from TCP connection: %s", err)
			}
			return
		}
		if _, err := conn.Write(bytes); err != nil {
			log.Printf("Failed to write bytes to TCP connection: %s", err)
		}
	}
}

func ServeTCP(ctx context.Context, port int32) error {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}
	defer listener.Close()
	log.Printf("Listening TCPServer on %s", addr)

	var (
		connChan = make(chan net.Conn, 1024)
		errChan  = make(chan error)
	)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				errChan <- err
				return
			} else {
				connChan <- conn
			}
		}
	}()

	for {
		select {
		case conn := <-connChan:
			log.Printf("New TCP connection: %s", conn.RemoteAddr())
			go HandleTCPConnection(conn)
		case err := <-errChan:
			log.Printf("Failed to accept new TCP connection: %s, closing", err)
			return fmt.Errorf("accept TCP connection: %w", err)
		case <-ctx.Done():
			return nil
		}
	}
}
