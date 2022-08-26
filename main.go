package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
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

func parsePortsString(s string) ([]int32, error) {
	if s == "" {
		return nil, nil
	}
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

func Serve(
	ctx context.Context,
	wg *sync.WaitGroup,
	network string,
	enabled bool,
	portsStr string,
	handle func(ctx context.Context, port int32) error,
) {
	if !enabled {
		return
	}

	ports, err := parsePortsString(portsStr)
	if err != nil {
		log.Fatalf("Failed to parse %s ports: %s", network, err)
	}
	for i := range ports {
		wg.Add(1)
		go func(port int32) {
			defer wg.Done()
			if err := handle(ctx, port); err != nil {
				log.Fatalf("Failed to start udp server: %s", err)
			}
		}(ports[i])
	}
}

func run(ctx context.Context) {
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

	ctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}
	Serve(ctx, wg, "UDP", enableUDP, udpPortsStr, ServeUDP)
	Serve(ctx, wg, "TCP", enableTCP, tcpPortsStr, ServeTCP)
	Serve(ctx, wg, "HTTP", enableHTTP, httpPortsStr, ServeHTTP)

	whenDrop(func() {
		log.Printf("Shutting down")
		cancel()
		wg.Wait()
	})
}

func main() {
	run(context.Background())
}
