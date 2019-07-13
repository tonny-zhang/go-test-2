package main

import (
	"log"
	. "test/test/tcp/net"
)

func main() {
	server := &Server{}

	server.OnListen(func(s *Server) {
		log.Printf("listen at %s\n", s.Address())
	})
	server.OnNewClient(func(c *Client) {
		log.Printf("client [%s] connect \n", c.Conn().RemoteAddr())
	})
	server.OnNewMessage(func(c *Client, code int16, message string) {
		log.Printf("get client [%s] code [%d] message [%s] \n", c.Conn().RemoteAddr(), code, message)
	})
	server.OnClientClosed(func(c *Client, e error) {
		log.Printf("client [%s] closed\n", c.Conn().RemoteAddr())
	})
	server.Listen("0.0.0.0:6006")
}
