package server

import (
	"Blockem/util"
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
)

type Server struct {
	dnsServers []*dns.Server
}

func getServerAddress(addr string) string {
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf(":%s", addr)
	}
	return addr
}

type NewServerFunc func(address string) (*dns.Server, error)

func NewServer(ctx context.Context) (server *Server, err error) {
	dnsServers, err := createServers()
	if err != nil {
		return nil, fmt.Errorf("failed to create servers: %w", err)
	}
	server = &Server{
		dnsServers: dnsServers,
	}
	return server, err
}

func createServers() ([]*dns.Server, error) {
	address := "53" // todo make configurable
	var dnsServers []*dns.Server
	tcpServer, err := createTCPServer(getServerAddress(address))
	if err != nil {
		return nil, err
	}
	dnsServers = append(dnsServers, tcpServer)
	udpServer, err := createUDPServer(getServerAddress(address))
	if err != nil {
		return nil, err

	}
	dnsServers = append(dnsServers, udpServer)
	return dnsServers, nil
}

func createTCPServer(address string) (*dns.Server, error) {
	return &dns.Server{
		Addr:    address, // tcp port
		Net:     "tcp",
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			fmt.Printf("TCP server started on %s\n", address)
		},
	}, nil
}

func createUDPServer(address string) (*dns.Server, error) {
	return &dns.Server{
		Addr:    address,
		Net:     "udp",
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			fmt.Printf("UDP server started on %s\n", address)
		},
	}, nil
}

func forwardRequest(r *dns.Msg) (*dns.Msg, error) {
	c := new(dns.Client)
	in, _, err := c.Exchange(r, upstreamServer)
	if err != nil {
		return nil, err
	}
	return in, nil

}

func (s *Server) registerDNSHandler(ctx context.Context) {
	for _, server := range s.dnsServers {
		handler := server.Handler.(*dns.ServeMux)
		handler.HandleFunc(".", func(w dns.ResponseWriter, m *dns.Msg) {
			s.OnRequest(ctx, w, m)
		})
	}
}

func extractClientIDFromHost(hostName string) string {
	const clientIDPrefix = "id-"
	if strings.HasPrefix(hostName, clientIDPrefix) && strings.Contains(hostName, ".") {
		return hostName[len(clientIDPrefix):strings.Index(hostName, ".")]
	}
	return ""
}


func newRequest(
	ctx context.Context,
	clientIP net.IP, clientID string,
	protocol model.RequestProtocol, request *dns.Msg,
) (context.Context ) {
  ctx := context.WithValue(ctx, {
	"req_id": uuid.New().String(),
	"question": util.QuestionToString(request.Question),
	})	
}


func (s *Server) OnRequest(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg) {
	ctx, request := newRequestFromDNS(ctx, w, msg)

	s.handleReq
}

func isBlocked(domain string) bool {
	for _, blockedDomain := range blockList {
		if strings.Contains(domain, blockedDomain) {
			return true
		}
	}
	return false
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
				fmt.Errorf("Error forwarding request: %s\n", err)

			}
			w.WriteMsg(resp)
			return
		}
	}
	w.WriteMsg(&msg)
}

var blockList = []string{
	"example.com",
	"pornhub.com",
}

var upstreamServer = "8.8.8.8"
