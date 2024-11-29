package main

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
)

type SipServer struct {
	ip      string
	port    int
	devices map[string]*Device
	mu      sync.Mutex
}

type Device struct {
	ID        string
	IP        string
	Port      int
	LastHeart time.Time
}

func NewSipServer(ip string, port int) *SipServer {
	return &SipServer{
		ip:      ip,
		port:    port,
		devices: make(map[string]*Device),
	}
}

func (s *SipServer) Start() {
	addr := &net.UDPAddr{IP: net.ParseIP(s.ip), Port: s.port}
	server, err := sipgo.NewServer(sipgo.ServerConfig{
		UserAgent: "SIP Server",
		Addr:      addr,
	})
	if err != nil {
		log.Fatalf("Failed to create SIP server: %v", err)
	}

	server.OnRequest(sip.REGISTER, s.handleRegister)
	server.OnRequest(sip.MESSAGE, s.handleMessage)
	server.OnRequest(sip.INVITE, s.handleInvite)
	server.OnRequest(sip.BYE, s.handleBye)

	log.Printf("Listening on %s:%d", s.ip, s.port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start SIP server: %v", err)
	}
}

func (s *SipServer) handleRegister(req *sip.Request, tx sip.ServerTransaction) {
	from, ok := req.From()
	if !ok {
		log.Printf("Failed to get 'From' header")
		return
	}

	fromAddr := from.Address.String()
	log.Printf("Received REGISTER request from %s", fromAddr)

	s.mu.Lock()
	defer s.mu.Unlock()

	deviceID := fromAddr[4:22] // Extract device ID from SIP URI
	s.devices[deviceID] = &Device{
		ID:        deviceID,
		IP:        req.Source().IP.String(),
		Port:      req.Source().Port,
		LastHeart: time.Now(),
	}

	resp := sip.NewResponseFromRequest(req, 200, "OK", "")
	tx.Respond(resp)
}

func (s *SipServer) handleMessage(req *sip.Request, tx sip.ServerTransaction) {
	from, ok := req.From()
	if !ok {
		log.Printf("Failed to get 'From' header")
		return
	}

	fromAddr := from.Address.String()
	log.Printf("Received MESSAGE request from %s", fromAddr)

	body := req.Body()
	if body != "" {
		log.Printf("Message body: %s", body)
	}

	resp := sip.NewResponseFromRequest(req, 200, "OK", "")
	tx.Respond(resp)
}

func (s *SipServer) handleInvite(req *sip.Request, tx sip.ServerTransaction) {
	from, ok := req.From()
	if !ok {
		log.Printf("Failed to get 'From' header")
		return
	}

	fromAddr := from.Address.String()
	log.Printf("Received INVITE request from %s", fromAddr)

	resp := sip.NewResponseFromRequest(req, 200, "OK", "")
	tx.Respond(resp)
}

func (s *SipServer) handleBye(req *sip.Request, tx sip.ServerTransaction) {
	from, ok := req.From()
	if !ok {
		log.Printf("Failed to get 'From' header")
		return
	}

	fromAddr := from.Address.String()
	log.Printf("Received BYE request from %s", fromAddr)

	resp := sip.NewResponseFromRequest(req, 200, "OK", "")
	tx.Respond(resp)
}

func main() {
	server := NewSipServer("0.0.0.0", 5060)
	server.Start()
}
