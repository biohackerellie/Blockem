package resolver

import (
	"Blockem/model"
	"Blockem/util"
	"context"
	"fmt"
	"github.com/miekg/dns"
	"net"
	"time"
)

func newRequest(question string, rType dns.Type) *model.Request {
	return &model.Request{
		Req:      util.NewMsgWithQuestion(question, rType),
		Protocol: model.RequestProtocolUDP,
	}
}

func newRequestWithClient(question string, rType dns.Type, ip string, clientNames ...string) *model.Request {
	return &model.Request{
		ClientIP:    net.ParseIP(ip),
		ClientNames: clientNames,
		Req:         util.NewMsgWithQuestion(question, rType),
		RequestTS:   time.Time{},
		Protocol:    model.RequestProtocolUDP,
	}
}

func newResponse(request *model.Request, rcode int, rtype model.ResponseType, reason string) *model.Response {
	response := new(dns.Msg)
	response.SetReply(request.Req)
	response.Rcode = rcode

	return &model.Response{
		Res:    response,
		RType:  rtype,
		Reason: reason,
	}
}
func newRequestWithClientID(question string, rType dns.Type, ip, requestClientID string) *model.Request {
	return &model.Request{
		ClientIP:        net.ParseIP(ip),
		RequestClientID: requestClientID,
		Req:             util.NewMsgWithQuestion(question, rType),
		RequestTS:       time.Time{},
		Protocol:        model.RequestProtocolUDP,
	}
}

type Resolver interface {
	fmt.Stringer

	Type() string

	Resolve(ctx context.Context, req *model.Request) (*model.Response, error)
}

type ChainedResolver interface {
	Resolver

	Next(n Resolver)

	GetNext() Resolver
}

type NextResolver struct {
	next Resolver
}

func (r *NextResolver) Next(n Resolver) {
	r.next = n
}

func (r *NextResolver) GetNext() Resolver {
	return r.next
}
