package dns

import (
	"github.com/miekg/dns"
	"log"
	"net"
)



func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true
	log.Printf("Received query for %s\n", r.Question[0].Name)
	for _, q := range r.Question {
		switch q.Qtype {
		case dns.TypeA:
			rr := &dns.A{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
				A: net.ParseIP("127.0.0.1")
			}
			msg.Answer = append(msg.Answer, rr)
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
