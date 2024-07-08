package util

import (
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/miekg/dns"
)

var (
	LogPrivacy   atomic.Bool
	alphanumeric = regexp.MustCompile("[^a-zA-Z0-9]+")
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

func QuestionToString(questions []dns.Question) string {
	result := make([]string, len(questions))
	for i, question := range questions {
		result[i] = fmt.Sprintf("%s (%s)", dns.TypeToString[question.Qtype], question.Name)
	}
	return Obfuscate(strings.Join(result, ", "))
}

func Obfuscate(in string) string {
	if LogPrivacy.Load() {
		return alphanumeric.ReplaceAllString(in, "*")
	}
	return in
}

func AnswerToString(answer []dns.RR) string {
	answers := make([]string, len(answer))

	for i, record := range answer {
		switch v := record.(type) {
		case *dns.A:
			answers[i] = fmt.Sprintf("A (%s)", v.A)
		case *dns.AAAA:
			answers[i] = fmt.Sprintf("AAAA (%s)", v.AAAA)
		case *dns.CNAME:
			answers[i] = fmt.Sprintf("CNAME (%s)", v.Target)
		case *dns.PTR:
			answers[i] = fmt.Sprintf("PTR (%s)", v.Ptr)
		default:
			answers[i] = fmt.Sprint(record.String())
		}
	}

	return Obfuscate(strings.Join(answers, ", "))
}
