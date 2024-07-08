package server

import (
	"Blockem/log"
	"Blockem/model"
	"Blockem/util"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

type Server struct {
	dnsServers []*dns.Server
}

func logger() *logrus.Entry {
	return log.PrefixedLog("server")
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
) (context.Context, *model.Request) {
	ctx, logger := log.CtxWithFields(ctx, logrus.Fields{
		"req_id":    uuid.New().String(),
		"question":  util.QuestionToString(request.Question),
		"client_ip": clientIP,
	})

	logger.WithFields(logrus.Fields{
		"client_request_id": request.Id,
		"client_id":         clientID,
		"protocol":          protocol,
	}).Trace("new request")

	req := model.Request{
		ClientIP:        clientIP,
		RequestClientID: clientID,
		Protocol:        protocol,
		Req:             request,
		RequestTS:       time.Now(),
	}

	return ctx, &req
}

func newRequestFromDNS(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg) (context.Context, *model.Request) {
	var (
		clientIP net.IP
		protocol model.RequestProtocol
	)

	if w != nil {
		clientIP, protocol = resolveClientIPAndProtocol(w.RemoteAddr())
	}

	var clientID string
	if con, ok := w.(dns.ConnectionStater); ok && con.ConnectionState() != nil {
		clientID = extractClientIDFromHost(con.ConnectionState().ServerName)
	}
	return newRequest(ctx, clientIP, clientID, protocol, msg)
}

func (s *Server) OnRequest(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg) {
	ctx, request := newRequestFromDNS(ctx, w, msg)

	s.handleReq(ctx, request, w)
}

func isBlocked(domain string) bool {
	for _, blockedDomain := range blockList {
		if strings.Contains(domain, blockedDomain) {
			return true
		}
	}
	return false
}

type msgWriter interface {
	WriteMsg(msg *dns.Msg) error
}

func (s *Server) handleReq(ctx context.Context, request *model.Request, w msgWriter) {
	response, err := s.resolve(ctx, request)
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

var upstreamServer = "8.8.8.8"

func resolveClientIPAndProtocol(addr net.Addr) (ip net.IP, protocol model.RequestProtocol) {
	switch a := addr.(type) {
	case *net.UDPAddr:
		return a.IP, model.RequestProtocolUDP
	case *net.TCPAddr:
		return a.IP, model.RequestProtocolTCP
	}
	return nil, model.RequestProtocolUDP
}
