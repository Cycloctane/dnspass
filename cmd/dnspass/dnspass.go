package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Cycloctane/dnspass/internal/config"
	"github.com/Cycloctane/dnspass/internal/dnsserver"
	"github.com/Cycloctane/dnspass/internal/proxy"
	"github.com/Cycloctane/dnspass/internal/web"
	"github.com/miekg/dns"
)

const (
	defaultUpstream = "8.8.8.8:53"
	defaultConfig   = "dnspass.toml"
)

var version = "dev"

//go:embed dnspass.example.toml
var exampleConfig []byte

func main() {
	configPath := flag.String("c", defaultConfig, "path to dnspass config file")
	example := flag.Bool("example", false, "generate an example config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if *example {
		fmt.Print(string(exampleConfig))
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg.ApplyRecords()

	upstreams := cfg.DNS.Upstream
	if len(upstreams) == 0 {
		upstreams = []string{defaultUpstream}
	}
	for i, upstream := range upstreams {
		if _, _, err := net.SplitHostPort(upstream); err != nil {
			upstreams[i] = net.JoinHostPort(upstream, "53")
		}
	}
	dnsserver.Upstreams = upstreams

	if cfg.Server.DNSProtocol != "" && cfg.Server.DNSListen == "" {
		log.Fatal("dns_protocol requires dns_listen in the config.")
	}

	if cfg.Server.DNSListen == "" && cfg.Server.ProxyListen == "" {
		log.Fatal("Nothing to do. Please specify dns_listen or proxy_listen in the config.")
	}

	var dnsServers []*dns.Server
	if cfg.Server.DNSListen != "" {
		proto := cfg.Server.DNSProtocol
		if proto == "" {
			proto = "udp"
		}
		for _, netProto := range strings.Split(proto, "+") {
			if netProto == "" {
				continue
			}
			server := dnsserver.NewDNSServer()
			server.Net = netProto
			server.Addr = cfg.Server.DNSListen
			dnsServers = append(dnsServers, server)
			go func(netProto string, server *dns.Server) {
				log.Printf("Starting DNS (%s) on %s ...\n", netProto, cfg.Server.DNSListen)
				if err := server.ListenAndServe(); err != nil {
					log.Fatal(err)
				}
			}(netProto, server)
		}
	}

	var httpProxy *http.Server
	if cfg.Server.ProxyListen != "" {
		httpProxy = proxy.NewHTTPProxyServer(cfg.Server.ProxyBlockIP)
		httpProxy.Addr = cfg.Server.ProxyListen
		go func() {
			log.Printf("Starting HTTP Proxy on %s ...\n", cfg.Server.ProxyListen)
			if err := httpProxy.ListenAndServe(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	var webServer *http.Server
	if cfg.Server.WebListen != "" {
		webServer, err = web.NewServer(cfg.Server.WebUserName, cfg.Server.WebPassWord)
		if err != nil {
			log.Fatal(err)
		}
		webServer.Addr = cfg.Server.WebListen
		go func() {
			log.Printf("Starting Web UI on http://%s ...\n", cfg.Server.WebListen)
			if err := webServer.ListenAndServe(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	for _, dnsServer := range dnsServers {
		dnsServer.Shutdown()
	}
	if httpProxy != nil {
		httpProxy.Shutdown(context.Background())
	}
	if webServer != nil {
		webServer.Shutdown(context.Background())
	}
}
