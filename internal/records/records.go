package records

import (
	"math/rand"
	"net/netip"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

type DNSType string

const (
	TypeA    DNSType = "A"
	TypeAAAA DNSType = "AAAA"
	TypePTR  DNSType = "PTR"
	TypeTXT  DNSType = "TXT"
)

type Record struct {
	Name  string
	Type  DNSType
	Value string
	TTL   uint32
}

type Store struct {
	mu   sync.RWMutex
	data map[DNSType]map[string][]Record
}

func NewStore() *Store {
	return &Store{data: make(map[DNSType]map[string][]Record)}
}

func normalizeName(name string) string {
	return dns.Fqdn(strings.ToLower(strings.TrimSpace(name)))
}

func normalizeType(t DNSType) (DNSType, bool) {
	s := strings.ToUpper(strings.TrimSpace(string(t)))
	switch s {
	case "A":
		return TypeA, true
	case "AAAA":
		return TypeAAAA, true
	case "PTR":
		return TypePTR, true
	case "TXT":
		return TypeTXT, true
	default:
		return "", false
	}
}

func normalizeValue(t DNSType, value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	} else if t == TypeA || t == TypeAAAA {
		ip, err := netip.ParseAddr(value)
		if err != nil {
			return "", false
		} else if t == TypeA && !ip.Is4() {
			return "", false
		} else if t == TypeAAAA && !ip.Is6() {
			return "", false
		}
		return ip.String(), true
	} else {
		return value, true
	}
}

func (s *Store) SetAll(records []Record) {
	data := make(map[DNSType]map[string][]Record)
	for _, r := range records {
		t, ok := normalizeType(r.Type)
		if !ok {
			continue
		}
		name := normalizeName(r.Name)
		value, ok := normalizeValue(t, r.Value)
		if !ok {
			continue
		}
		if _, ok := data[t]; !ok {
			data[t] = make(map[string][]Record)
		}
		list := data[t][name]
		skipped := false
		for _, item := range list {
			if item.Value == value {
				skipped = true
				break
			}
		}
		if skipped {
			continue
		}
		data[t][name] = append(list, Record{Name: name, Type: t, Value: value, TTL: r.TTL})
	}

	s.mu.Lock()
	s.data = data
	s.mu.Unlock()
}

func (s *Store) Add(r Record) bool {
	t, ok := normalizeType(r.Type)
	if !ok {
		return false
	}
	name := normalizeName(r.Name)
	value, ok := normalizeValue(t, r.Value)
	if !ok {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data[t] == nil {
		s.data[t] = make(map[string][]Record)
	}
	list := s.data[t][name]
	for _, item := range list {
		if item.Value == value {
			return false
		}
	}
	s.data[t][name] = append(list, Record{Name: name, Type: t, Value: value, TTL: r.TTL})
	return true
}

func (s *Store) Delete(name string, t DNSType, value string) bool {
	t, ok := normalizeType(t)
	if !ok {
		return false
	}
	name = normalizeName(name)
	value, ok = normalizeValue(t, value)
	if !ok {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.data[t][name]
	if len(list) == 0 {
		return false
	}
	filtered := list[:0]
	removed := false
	for _, item := range list {
		if item.Value == value {
			removed = true
			continue
		}
		filtered = append(filtered, item)
	}
	if !removed {
		return false
	}
	if len(filtered) == 0 {
		delete(s.data[t], name)
	} else {
		s.data[t][name] = filtered
	}
	return true
}

func (s *Store) Get(name string, t DNSType) []Record {
	t, ok := normalizeType(t)
	if !ok {
		return nil
	}
	name = normalizeName(name)

	s.mu.RLock()
	list := s.data[t][name]
	s.mu.RUnlock()
	if len(list) == 0 {
		return nil
	}
	out := make([]Record, len(list))
	copy(out, list)
	return out
}

func (s *Store) Get1(name string, t DNSType) (Record, bool) {
	list := s.Get(name, t)
	if len(list) == 0 {
		return Record{}, false
	}
	return list[rand.Intn(len(list))], true
}

func (s *Store) List() []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Record
	for _, byName := range s.data {
		for _, list := range byName {
			out = append(out, list...)
		}
	}
	if len(out) == 0 {
		return nil
	}
	copyOut := make([]Record, len(out))
	copy(copyOut, out)
	return copyOut
}
