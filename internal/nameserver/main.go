package nameserver

import (
	"net"
	"time"

	"github.com/Cycloctane/dnspass/internal/records"
	"github.com/miekg/dns"
)

const defaultTTL = 300

var UpstreamDNS = "8.8.8.8:53"
var client = &dns.Client{Timeout: 5 * time.Second}

func lookupA(q *dns.Question) ([]dns.RR, bool) {
	res := records.Lookup(q.Name, "A")
	if len(res) == 0 {
		res2 := records.Lookup(q.Name, "AAAA")
		if len(res2) != 0 {
			return nil, true
		} else {
			return nil, false
		}
	}

	ans := make([]dns.RR, len(res))
	for k, v := range res {
		ans[k] = &dns.A{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    defaultTTL,
			},
			A: net.ParseIP(v),
		}
	}
	return ans, true
}

func lookupAAAA(q *dns.Question) ([]dns.RR, bool) {
	res := records.Lookup(q.Name, "AAAA")
	if len(res) == 0 {
		res2 := records.Lookup(q.Name, "A")
		if len(res2) != 0 {
			return nil, true
		} else {
			return nil, false
		}
	}

	ans := make([]dns.RR, len(res))
	for k, v := range res {
		ans[k] = &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    defaultTTL,
			},
			AAAA: net.ParseIP(v),
		}
	}
	return ans, true
}

func forwardQuery(q *dns.Question) ([]dns.RR, error) {
	m := &dns.Msg{}
	m.SetQuestion(q.Name, q.Qtype)
	m.RecursionDesired = true

	resp, _, err := client.Exchange(m, UpstreamDNS)
	if err != nil {
		return nil, err
	}
	return resp.Answer, err
}

func handler(w dns.ResponseWriter, req *dns.Msg) {
	q := &req.Question[0]
	resp := &dns.Msg{}

	var ans []dns.RR
	var ok bool
	if q.Qtype == dns.TypeA {
		ans, ok = lookupA(q)
	} else if q.Qtype == dns.TypeAAAA {
		ans, ok = lookupAAAA(q)
	}

	if !ok {
		var err error
		ans, err = forwardQuery(q)
		if err != nil {
			resp.SetRcode(req, dns.RcodeServerFailure)
			w.WriteMsg(resp)
			return
		} else {
			ok = true
		}
	}

	if ok && len(ans) == 0 {
		resp.SetRcode(req, dns.RcodeNameError)
		w.WriteMsg(resp)
		return
	}

	resp.Answer = ans
	resp.SetReply(req)
	w.WriteMsg(resp)
}

func NewDNSServer() *dns.Server {
	server := &dns.Server{}
	server.Handler = dns.HandlerFunc(handler)
	return server
}
