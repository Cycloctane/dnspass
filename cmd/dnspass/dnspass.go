package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cycloctane/dnspass/internal/config"
	dnsserver "github.com/Cycloctane/dnspass/internal/dnsserver"
	"github.com/Cycloctane/dnspass/internal/proxy"
	"github.com/miekg/dns"
)

const defaultUpstream = "8.8.8.8:53"

var (
	proxyListenAddr = flag.String("l", ":8080", "HTTP proxy listening address:port")
	dnsListenAddr   = flag.String("dns", "", "UDP DNS server listening address:port")
	upstreamDNS     = flag.String("upstream", defaultUpstream, "Upstream Nameserver address")
	nameFile        = flag.String("c", "", "Path to a file containing hostname/addresse pairs")
)

func init() {
	flag.Parse()
	if _, _, err := net.SplitHostPort(*upstreamDNS); err != nil {
		dnsserver.UpstreamDNS = net.JoinHostPort(*upstreamDNS, "53")
	} else {
		dnsserver.UpstreamDNS = *upstreamDNS
	}
	if *nameFile != "" {
		if err := config.ParseFile(*nameFile); err != nil {
			log.Fatalf("Failed to parse records file: %v\n", err)
		}
	}
}

func main() {
	httpProxy := proxy.NewHTTPProxyServer()
	httpProxy.Addr = *proxyListenAddr

	go func() {
		log.Printf("Starting HTTP Proxy on %s\n", *proxyListenAddr)
		if err := httpProxy.ListenAndServe(); err != nil {
			log.Printf("HTTP proxy failed: %v\n", err)
			return
		}
	}()

	var dnsServer *dns.Server
	if *dnsListenAddr != "" {
		dnsServer = dnsserver.NewDNSServer()
		dnsServer.Net = "udp"
		dnsServer.Addr = *dnsListenAddr
		go func() {
			log.Printf("Starting DNS server on %s\n", *dnsListenAddr)
			if err := dnsServer.ListenAndServe(); err != nil {
				log.Printf("DNS server failed: %v\n", err)
				return
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down HTTP Proxy...\n")
	httpProxy.Shutdown(context.Background())
	if dnsServer != nil {
		log.Printf("Shutting down DNS...\n")
		dnsServer.Shutdown()
	}
}
