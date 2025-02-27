package records

import (
	"slices"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

type records struct {
	mu      sync.RWMutex
	records map[string][]string
}

func (r *records) get(host string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.records[strings.ToLower(host)]
}

func (r *records) add(host, addr string) {
	host = dns.Fqdn(strings.ToLower(host))

	r.mu.Lock()
	defer r.mu.Unlock()
	existing := r.records[host]
	if slices.Contains(existing, addr) {
		return
	}
	r.records[host] = append(existing, addr)
}
