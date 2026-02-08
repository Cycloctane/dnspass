package proxy

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/Cycloctane/dnspass/internal/records"
	"github.com/elazarl/goproxy"
	"github.com/miekg/dns"
)

func dial(blockDirectIP bool, network, addr string) (net.Conn, error) {
	hostname, port, _ := net.SplitHostPort(addr)
	if net.ParseIP(hostname) != nil {
		if blockDirectIP {
			log.Printf("[proxy] Blocked direct IP connection to %s", hostname)
			return nil, errors.New("No route to host")
		}
		log.Printf("[proxy] Passing through IP address %s", hostname)
		return net.Dial(network, addr)
	}

	hostname = dns.Fqdn(hostname)
	if rec, ok := records.Lookup1(hostname, records.TypeAAAA); ok {
		log.Printf("[proxy] Resolved %s to %s", hostname, rec.Value)
		return net.Dial(network, net.JoinHostPort(rec.Value, port))
	} else if rec, ok := records.Lookup1(hostname, records.TypeA); ok {
		log.Printf("[proxy] Resolved %s to %s", hostname, rec.Value)
		return net.Dial(network, net.JoinHostPort(rec.Value, port))
	} else {
		log.Printf("[proxy] Failed to find %s", hostname)
		return net.Dial(network, addr)
	}
}

func NewHTTPProxyServer(blockDirectIP bool) *http.Server {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Tr.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
		return dial(blockDirectIP, network, addr)
	}
	proxy.ConnectDial = func(network, addr string) (net.Conn, error) {
		return dial(blockDirectIP, network, addr)
	}
	// proxy.Verbose = true

	return &http.Server{Handler: proxy}
}
