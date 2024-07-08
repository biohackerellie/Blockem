package resolver

import (
	"Blockem/model"
	"context"

	"github.com/miekg/dns"
)

var blockList = []string{
	"example.com",
	"pornhub.com",
}

type FilteringResolver struct {
	NextResolver
}

func NewFilteringResolver() *FilteringResolver {
	return &FilteringResolver{}
}

func (r *FilteringResolver) Resolve(ctx context.Context, req *model.Request) (*model.Response, error) {
	if isBlocked(req.Req.Question[0].Name) {
		return newResponse(req, dns.RcodeNameError, model.ResponseTypeBlocked, "domain is blocked"), nil
	}
	return r.NextResolver.Resolve(ctx, req)
}
