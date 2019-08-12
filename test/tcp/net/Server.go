package net

import (
	"crypto/tls"
	"log"
	"net"
)

// Server tcp server
type Server struct {
	address        string // Address to open connection: localhost:9999
	config         *tls.Config
	onListen       func(s *Server)
	onNewClient    func(c *Client)
	onClientClosed func(c *Client, err error)
	onNewMessage   func(c *Client, code int16, message string)
}

// OnNewClient Called right after server starts listening new client
func (s *Server) OnNewClient(callback func(c *Client)) {
	s.onNewClient = callback
}

// OnClientClosed Called right after connection closed
func (s *Server) OnClientClosed(callback func(c *Client, err error)) {
	s.onClientClosed = callback
}

// OnNewMessage Called when Client receives new message
func (s *Server) OnNewMessage(callback func(c *Client, code int16, message string)) {
	s.onNewMessage = callback
}

// OnListen emit on server listen
func (s *Server) OnListen(callback func(s *Server)) {
	s.onListen = callback
}

// Address get address
func (s *Server) Address() string {
	return s.address
}

// Listen starts network server
func (s *Server) Listen(address string) {
	s.address = address
	var listener net.Listener
	var err error
	if s.config == nil {
		listener, err = net.Listen("tcp", s.address)
	} else {
		listener, err = tls.Listen("tcp", s.address, s.config)
	}
	if err != nil {
		log.Fatal("Error starting TCP server.")
	} else {
		log.Println("Creating server with address", s.address)
	}
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		client := &Client{
			conn:   conn,
			Server: s,
		}
		go client.work()
	}
}
