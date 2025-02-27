DNSPass
===

Simple DNS and HTTP proxy. Custom DNS resolving for debugging without messing up /etc/hosts everytime.

## Usage

Write custom hostname/address pairs to txt.

```
example.com/172.17.0.2
...
```

```bash
./dnspass -c resolves.txt -l 127.0.0.1:3128 --dns 127.0.0.1:8053 &
http_proxy=http://127.0.0.1:3128 curl -v -I http://example.com:8000
```
