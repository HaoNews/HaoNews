package host

import (
	"net"
	"strconv"
	"testing"
)

func TestResolveListenAddrIncrementsWhenPortIsOccupied(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	next, err := resolveListenAddr(addr)
	if err != nil {
		t.Fatalf("resolveListenAddr() error = %v", err)
	}
	if next == addr {
		t.Fatalf("resolveListenAddr() = %q, want incremented port", next)
	}
}

func TestReserveListenAddrAllocatesEphemeralPort(t *testing.T) {
	t.Parallel()

	listener, next, err := reserveListenAddr("127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserveListenAddr() error = %v", err)
	}
	defer listener.Close()

	host, portText, err := net.SplitHostPort(next)
	if err != nil {
		t.Fatalf("SplitHostPort() error = %v", err)
	}
	if host != "127.0.0.1" {
		t.Fatalf("host = %q, want 127.0.0.1", host)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatalf("Atoi() error = %v", err)
	}
	if port <= 0 {
		t.Fatalf("port = %d, want > 0", port)
	}
}
