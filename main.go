package main

import (
	"flag"
	"log"
	"net"
	"strconv"
)

var (
	name     = flag.String("name", "", "name to self identify over the network with")
	grpcAddr = flag.String("grpc-addr", ":8079", "optional: address to bind for grpc (host:port or :port)")
	tcpAddr  = flag.String("tcp-addr", ":8078", "optional: address to bind for ascii tcp (host:port or :port)")
)

func main() {
	flag.Parse()

	if *name == "" {
		log.Fatal("-name must be provided")
	}

	errCh := make(chan error, 2)

	any := false
	if isEnabledAddr(*grpcAddr) {
		any = true
		go func() {
			errCh <- serveGRPC(*name, *grpcAddr)
		}()
	}
	if isEnabledAddr(*tcpAddr) {
		any = true
		go func() {
			errCh <- serveTCP(*name, *tcpAddr)
		}()
	}

	if !any {
		log.Fatal("one of -grpc-addr or -tcp-addr should be enabled")
	}

	if err := <-errCh; err != nil {
		log.Fatal(err)
	}
}

func isEnabledAddr(addr string) bool {
	return addr != "" && addr != "disabled"
}

func isValidAddr(addr string) bool {
	if !isEnabledAddr(addr) {
		return false
	}
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	v, err := strconv.Atoi(port)
	return err == nil && v > 0
}
