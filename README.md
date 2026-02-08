DNSPass
=======

Simple DNS Server and HTTP proxy with configurable DNS records in webui.

Currentyly supported record types: A, AAAA, PTR, TXT

## Usage

Use `-example` flag to generate a config file template (See [dnspass.example.toml](cmd/dnspass/dnspass.example.toml)):

```bash
./dnspass -example > dnspass.toml
```

Start the server with config file:

```bash
./dnspass -c dnspass.toml
```

Config DNS records in the webui (default http://localhost:8080) to customize DNS resolving for DNS server and HTTP proxy (CONNECT hostname).

```bash
http_proxy=http://127.0.0.1:3128 curl -v -I http://example.com:8000/
```
