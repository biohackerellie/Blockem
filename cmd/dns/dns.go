package main

import (
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
)

var blockList = []string{
	"example.com",
	"pornhub.com",
}

var upstreamServer = "8.8.8.8"

func isBlocked(domain string) bool {
	for _, blockedDomain := range blockList {
		if strings.Contains(domain, blockedDomain) {
			return true
		}
	}
	return false
}

func forwardRequest(r *dns.Msg) (*dns.Msg, error) {
	c := new(dns.Client)
	in, _, err := c.Exchange(r, upstreamServer)
	if err != nil {
		return nil, err
	}
	return in, nil

}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true
	log.Printf("Received query for %s\n", r.Question[0].Name)
	for _, q := range r.Question {
		if isBlocked(q.Name) {
			msg.Rcode = dns.RcodeNameError
		} else {
			resp, err := forwardRequest(r)
			if err != nil {
				log.Printf("Error forwarding request: %s\n", err)
				return

			}
			w.WriteMsg(resp)
			return
		}
	}
	w.WriteMsg(&msg)
}

func main() {
	dns.HandleFunc(".", handleDNSRequest)

	server := &dns.Server{Addr: ":53", Net: "udp"}
	log.Printf("Starting DNS server on port 53...")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n", err)
	}
}
