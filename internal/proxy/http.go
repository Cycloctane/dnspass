package proxy

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/Cycloctane/dnspass/internal/records"
	"github.com/elazarl/goproxy"
	"github.com/miekg/dns"
)

func dial(network, addr string) (net.Conn, error) {
	hostname, port, _ := net.SplitHostPort(addr)
	if net.ParseIP(hostname) != nil {
		log.Printf("[proxy] Passing through IP address %s", hostname)
		return net.Dial(network, addr)
	}

	hostname = dns.Fqdn(hostname)
	if newAddr := records.Lookup1(hostname, "AAAA"); newAddr != "" {
		log.Printf("[proxy] Resolved %s to %s", hostname, newAddr)
		return net.Dial(network, net.JoinHostPort(newAddr, port))
	} else if newAddr := records.Lookup1(hostname, "A"); newAddr != "" {
		log.Printf("[proxy] Resolved %s to %s", hostname, newAddr)
		return net.Dial(network, net.JoinHostPort(newAddr, port))
	} else {
		log.Printf("[proxy] Failed to find %s", hostname)
		return net.Dial(network, addr)
	}
}

func NewHTTPProxyServer() *http.Server {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Tr.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
		return dial(network, addr)
	}
	proxy.ConnectDial = dial
	// proxy.Verbose = true

	return &http.Server{Handler: proxy}
}
