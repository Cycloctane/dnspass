package dnsserver

import (
	"log"
	"net"
	"time"

	"github.com/Cycloctane/dnspass/internal/records"
	"github.com/miekg/dns"
)

const defaultTTL = 300

var Upstreams = []string{"8.8.8.8:53"}
var client = &dns.Client{Timeout: 5 * time.Second}

func lookupA(q *dns.Question) ([]dns.RR, bool) {
	res := records.Lookup(q.Name, records.TypeA)
	if len(res) == 0 {
		if res2 := records.Lookup(q.Name, records.TypeAAAA); len(res2) != 0 {
			return nil, true
		}
		return nil, false
	}

	ans := make([]dns.RR, len(res))
	for k, v := range res {
		ans[k] = &dns.A{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    recordTTL(v.TTL),
			},
			A: net.ParseIP(v.Value),
		}
	}
	return ans, true
}

func lookupAAAA(q *dns.Question) ([]dns.RR, bool) {
	res := records.Lookup(q.Name, records.TypeAAAA)
	if len(res) == 0 {
		if res2 := records.Lookup(q.Name, records.TypeA); len(res2) != 0 {
			return nil, true
		}
		return nil, false
	}

	ans := make([]dns.RR, len(res))
	for k, v := range res {
		ans[k] = &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    recordTTL(v.TTL),
			},
			AAAA: net.ParseIP(v.Value),
		}
	}
	return ans, true
}

func lookupTXT(q *dns.Question) ([]dns.RR, bool) {
	res := records.Lookup(q.Name, records.TypeTXT)
	if len(res) == 0 {
		return nil, false
	}
	ans := make([]dns.RR, len(res))
	for k, v := range res {
		ans[k] = &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    recordTTL(v.TTL),
			},
			Txt: []string{v.Value},
		}
	}
	return ans, true
}

func lookupPTR(q *dns.Question) ([]dns.RR, bool) {
	res := records.Lookup(q.Name, records.TypePTR)
	if len(res) == 0 {
		return nil, false
	}
	ans := make([]dns.RR, len(res))
	for k, v := range res {
		ans[k] = &dns.PTR{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    recordTTL(v.TTL),
			},
			Ptr: v.Value,
		}
	}
	return ans, true
}

func recordTTL(ttl uint32) uint32 {
	if ttl == 0 {
		return defaultTTL
	}
	return ttl
}

func forwardQuery(req *dns.Msg) (*dns.Msg, error) {
	var lastErr error
	for _, upstream := range Upstreams {
		resp, _, err := client.Exchange(req, upstream)
		if err == nil {
			resp.Id = req.Id
			return resp, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func customAnswer(q *dns.Question) ([]dns.RR, bool) {
	switch q.Qtype {
	case dns.TypeA:
		return lookupA(q)
	case dns.TypeAAAA:
		return lookupAAAA(q)
	case dns.TypePTR:
		return lookupPTR(q)
	case dns.TypeTXT:
		return lookupTXT(q)
	default:
		return nil, false
	}
}

func handler(w dns.ResponseWriter, req *dns.Msg) {
	q := &req.Question[0]
	resp := &dns.Msg{}

	if ans, ok := customAnswer(q); ok {
		if len(ans) == 0 {
			resp.SetRcode(req, dns.RcodeNameError)
			w.WriteMsg(resp)
			return
		}
		log.Printf("[dns] Resolved %s - %s", dns.Type(q.Qtype).String(), q.Name)
		resp.Answer = ans
		resp.SetReply(req)
		w.WriteMsg(resp)
		return
	}

	forwardReq := &dns.Msg{}
	forwardReq.SetQuestion(q.Name, q.Qtype)
	forwardReq.Id = req.Id
	forwardReq.RecursionDesired = true
	upstreamResp, err := forwardQuery(forwardReq)
	if err != nil {
		resp.SetRcode(req, dns.RcodeServerFailure)
		w.WriteMsg(resp)
		return
	}

	w.WriteMsg(upstreamResp)
}

func NewDNSServer() *dns.Server {
	server := &dns.Server{}
	server.Handler = dns.HandlerFunc(handler)
	return server
}
