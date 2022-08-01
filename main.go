package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func whenDrop(hooks ...func()) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	for _, hook := range hooks {
		hook()
	}
}

func NewTCPServer(port int32) error {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}
	defer listener.Close()
	log.Printf("Listening TCPServer on %s", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept tcp: %w", err)
		}

		go func(conn net.Conn) {
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
		}(conn)
	}
}

func NewUDPServer(port int32) error {
	addr := net.UDPAddr{Port: int(port), IP: net.ParseIP("0.0.0.0")}
	listener, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer listener.Close()
	log.Printf("Listening UDPServer on %s", addr.String())
	for {
		b := make([]byte, 4096)
		_, remote, err := listener.ReadFromUDP(b)
		if err != nil {
			return fmt.Errorf("read from udp: %w", err)
		}
		if _, err := listener.WriteToUDP(b, remote); err != nil {
			log.Printf("Failed to write bytes to UDP connection: %s", err)
		}
	}
}

func NewHTTPServer(port int32) error {
	mux := http.NewServeMux()
	mux.Handle("/echo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read body: %s", err)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unknown error"))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))

	mux.Handle("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))

	addr := fmt.Sprintf("0.0.0.0:%d", port)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Printf("Listening HTTPServer on %s", addr)
	return server.ListenAndServe()
}

func parsePortsString(s string) ([]int32, error) {
	m := make(map[int32]struct{})
	for _, portStr := range strings.Split(s, ",") {
		port, err := strconv.ParseInt(portStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parse port: %w", err)
		}
		m[int32(port)] = struct{}{}
	}

	var ports []int32
	for port := range m {
		ports = append(ports, port)
	}

	return ports, nil
}

func main() {
	var (
		enableUDP    bool
		enableTCP    bool
		enableHTTP   bool
		udpPortsStr  string
		tcpPortsStr  string
		httpPortsStr string
	)

	flag.BoolVar(&enableUDP, "enable-udp", false, "enable udp server")
	flag.BoolVar(&enableTCP, "enable-tcp", false, "enable tcp server")
	flag.BoolVar(&enableHTTP, "enable-http", false, "enable http server")
	flag.StringVar(&udpPortsStr, "udp-ports", "", "udp ports")
	flag.StringVar(&tcpPortsStr, "tcp-ports", "", "tcp ports")
	flag.StringVar(&httpPortsStr, "http-ports", "", "http ports")
	flag.Parse()

	var (
		udpPorts  []int32
		tcpPorts  []int32
		httpPorts []int32
		err       error
	)

	if enableUDP {
		udpPorts, err = parsePortsString(udpPortsStr)
		if err != nil {
			log.Fatalf("Failed to parse udp ports: %s", err)
		}

		for i := range udpPorts {
			go func(port int32) {
				if err := NewUDPServer(port); err != nil {
					log.Fatalf("Failed to start udp server: %s", err)
				}
			}(udpPorts[i])
		}
	}

	if enableTCP {
		tcpPorts, err = parsePortsString(tcpPortsStr)
		if err != nil {
			log.Fatalf("Failed to parse tcp ports: %s", err)
		}

		for i := range tcpPorts {
			go func(port int32) {
				if err := NewTCPServer(port); err != nil {
					log.Fatalf("Failed to start tcp server: %s", err)
				}
			}(tcpPorts[i])
		}
	}

	if enableHTTP {
		httpPorts, err = parsePortsString(httpPortsStr)
		if err != nil {
			log.Fatalf("Failed to parse http ports: %s", err)
		}

		for i := range httpPorts {
			go func(port int32) {
				if err := NewHTTPServer(port); err != nil {
					log.Fatalf("Failed to start http server: %s", err)
				}
			}(httpPorts[i])
		}
	}

	nServers := 0
	if enableUDP {
		nServers += len(udpPorts)
	}
	if enableTCP {
		nServers += len(tcpPorts)
	}
	if enableHTTP {
		nServers += len(httpPorts)
	}

	if nServers == 0 {
		log.Fatal("No server enabled")
	}

	whenDrop(func() {
		log.Printf("Shutting down")
	})
}
