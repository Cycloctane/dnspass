FROM golang:1.25 AS builder
WORKDIR /src

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dnspass ./cmd/dnspass

FROM scratch

COPY --from=builder /src/dnspass /dnspass
COPY <<-EOF /dnspass.toml
    [server]
    dns_listen = "0.0.0.0:53"
    dns_protocol = "tcp+udp"
    proxy_listen = "0.0.0.0:3128"
    web_listen = "0.0.0.0:8080"
EOF

EXPOSE 53/udp 53/tcp 3128 8080
ENTRYPOINT ["/dnspass"]
