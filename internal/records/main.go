package records

import (
	"math/rand"
	"net/netip"
)

var recordsA = &records{records: make(map[string][]string)}
var recordsAAAA = &records{records: make(map[string][]string)}

func Lookup(host, qtype string) []string {
	var addrs []string
	if qtype == "AAAA" {
		addrs = recordsAAAA.get(host)
	} else if qtype == "A" {
		addrs = recordsA.get(host)
	}
	return addrs
}

func Lookup1(host, qtype string) string {
	addrs := Lookup(host, qtype)
	length := len(addrs)
	if length == 0 {
		return ""
	} else if length == 1 {
		return addrs[0]
	} else {
		return addrs[rand.Intn(length)]
	}
}

func Add(host, addr string) {
	if ip, err := netip.ParseAddr(addr); err != nil {
		return
	} else if ip.Is6() {
		recordsAAAA.add(host, addr)
	} else if ip.Is4() {
		recordsA.add(host, addr)
	}
}
