package util

import (
	"fmt"

	"github.com/miekg/dns"
)

func NewMsgWithAnswer(domain string, ttl uint, dnsType dns.Type, address string) (*dns.Msg, error) {
	rr, err := dns.NewRR(fmt.Sprintf("%s\t%d\tIN\t%s\t%s", domain, ttl, dnsType, address))
	if err != nil {
		return nil, err
	}
	msg := new(dns.Msg)
	msg.Answer = []dns.RR{rr}
	return msg, nil
}

func NewMsgWithQuestion(question string, qType dns.Type) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(question), uint16(qType))

	return msg
}
