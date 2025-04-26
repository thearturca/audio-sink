package sink

import (
	"context"
	"fmt"
	"log"
	"net"
)

type Listener struct {
	address *net.UDPAddr
}

func NewServer(ctx context.Context, host string, port int, bufferSize int) *Server {
	return &Server{
		host:    host,
		port:    port,
		clients: make(map[string]Listener),
		ctx:     ctx,
		audio:   make([]byte, bufferSize),
	}
}

type Server struct {
	udp_listener *net.UDPConn
	clients      map[string]Listener
	host         string
	port         int
	ctx          context.Context
	audio        []byte
}

func (server *Server) Start() error {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(server.host), Port: server.port})

	if err != nil {
		return err
	}

	server.udp_listener = listener

	go server.listener()

	return nil
}

func (server *Server) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = server.audio[i]
	}

	return len(p), nil
}

func (server *Server) listener() {
	for {
		select {
		case <-server.ctx.Done():
			return
		default:
			_, addr, err := server.udp_listener.ReadFromUDP(server.audio)

			if err != nil {
				fmt.Println(err)
				return
			}

			if _, ok := server.clients[addr.String()]; !ok {
				server.clients[addr.String()] = Listener{addr}
				log.Printf("New client: %v", addr.String())
			}

		}

	}
}

/* func (server *Server) Broadcast(audio *[]byte) {
	for key, listener := range server.clients {
		_, err := server.udp_listener.WriteToUDP(*audio, listener.address)

		if err != nil {
			fmt.Println(err)
			delete(server.clients, key)
			return
		}
	}
} */

func (server *Server) Close() error {
	if server.udp_listener == nil {
		return nil
	}

	return server.udp_listener.Close()
}
