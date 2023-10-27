package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func serveTCP(name, addr string) error {
	if !isValidAddr(addr) {
		return fmt.Errorf("-tcp-addr is invalid %q", addr)
	}
	log.Printf("TCP listening on %s", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("tcp-server: error accepting connection: %v", err)
			continue
		}
		go func(conn net.Conn) {
			if err := handleTCPConn(name, conn); err != nil {
				log.Printf("tcp-server: error handling connection: %v", err)
			}
		}(conn)
	}
}

func handleTCPConn(name string, conn net.Conn) error {
	defer conn.Close()

	br := bufio.NewReader(conn)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return fmt.Errorf("could not read command string: %w", err)
		}
		cmd := strings.ToLower(strings.TrimSpace(line))

		switch cmd {
		case "describe":
			_, err := conn.Write([]byte(name + "\n"))
			if err != nil {
				return fmt.Errorf("could not write response: %w", err)
			}
		case "quit":
			return nil
		default:
			return fmt.Errorf("unknown command string: %s", cmd)
		}
	}
}
