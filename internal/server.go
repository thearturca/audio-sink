package sink

import (
	"context"
	"fmt"
"encoding/binary"
    "math"
	"log"
	"sync"
	"net"
)

type Listener struct {
	address *net.UDPAddr
	audio   []byte
	mutex   sync.Mutex
}

func NewServer(ctx context.Context, host string, port int, bufferSize int) *Server {
	return &Server{
		host:       host,
		port:       port,
		clients:    make(map[string]*Listener),
		ctx:        ctx,
		bufferSize: bufferSize,
	}
}

type Server struct {
	udp_listener *net.UDPConn
	clients      map[string]*Listener
	host         string
	port         int
	ctx          context.Context
	bufferSize   int
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
	for _, listener := range server.clients {
        listener.mutex.Lock()

        if len(listener.audio) == 0 {
            listener.mutex.Unlock()
            continue
        }

		for i := 0; i < len(p) && i < len(listener.audio); i += 4 {
            audioFloat32 := math.Float32frombits(binary.LittleEndian.Uint32(listener.audio[i:i+4]))
            pFloat32 := math.Float32frombits(binary.LittleEndian.Uint32(p[i:i+4]))

            pFloat32 += audioFloat32

            binary.LittleEndian.PutUint32(p[i:i+4], math.Float32bits(pFloat32))
		}

        if len(listener.audio) < len(p) {
            listener.audio = nil
        } else {
            listener.audio = listener.audio[len(p):]
        }

        listener.mutex.Unlock()
	}

	return len(p), nil
}

func (server *Server) listener() {
	for {
		select {
		case <-server.ctx.Done():
			return
		default:
			buffer := make([]byte, server.bufferSize)
			n, addr, err := server.udp_listener.ReadFromUDP(buffer)

            if n == 0 {
                continue
            }

			if err != nil {
				fmt.Println(err)
				return
			}

			if listener, ok := server.clients[addr.String()]; !ok {
                server.clients[addr.String()] = &Listener{addr, buffer[:n], sync.Mutex{}}
				log.Printf("New client: %v", addr.String())
			} else {
                listener.mutex.Lock()
                listener.audio = append(listener.audio, buffer[:n]...)
                listener.mutex.Unlock()
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
