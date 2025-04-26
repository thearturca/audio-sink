package sink

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
)

type ProducerClient struct {
	address *net.UDPAddr
	audio   []byte
	mutex   sync.Mutex
}

func NewConsumer(ctx context.Context, host string, port int, bufferSize int) *Consumer {
	return &Consumer{
		host:       host,
		port:       port,
		clients:    make(map[string]*ProducerClient),
		ctx:        ctx,
		bufferSize: bufferSize,
	}
}

type Consumer struct {
	udp_listener *net.UDPConn
	clients      map[string]*ProducerClient
	host         string
	port         int
	ctx          context.Context
	bufferSize   int
}

func (consumer *Consumer) Start() error {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(consumer.host), Port: consumer.port})

	if err != nil {
		return err
	}

	consumer.udp_listener = listener

	go consumer.listener()

	return nil
}

func (consumer *Consumer) Read(p []byte) (n int, err error) {
	for _, client := range consumer.clients {
		client.mutex.Lock()

		if len(client.audio) == 0 {
			client.mutex.Unlock()
			continue
		}

		for i := 0; i < len(p) && i < len(client.audio); i += 4 {
			audioFloat32 := math.Float32frombits(binary.LittleEndian.Uint32(client.audio[i : i+4]))
			pFloat32 := math.Float32frombits(binary.LittleEndian.Uint32(p[i : i+4]))

			pFloat32 += audioFloat32

			binary.LittleEndian.PutUint32(p[i:i+4], math.Float32bits(pFloat32))
		}

		if len(client.audio) < len(p) {
			client.audio = nil
		} else {
			client.audio = client.audio[len(p):]
		}

		client.mutex.Unlock()
	}

	return len(p), nil
}

func (consumer *Consumer) listener() {
	for {
		select {
		case <-consumer.ctx.Done():
			return
		default:
			buffer := make([]byte, consumer.bufferSize)
			n, addr, err := consumer.udp_listener.ReadFromUDP(buffer)

			if n == 0 {
				continue
			}

			if err != nil {
				fmt.Println(err)
				return
			}

			if client, ok := consumer.clients[addr.String()]; !ok {
				consumer.clients[addr.String()] = &ProducerClient{addr, buffer[:n], sync.Mutex{}}
				log.Printf("New client: %v", addr.String())
			} else {
				client.mutex.Lock()
				client.audio = append(client.audio, buffer[:n]...)
				client.mutex.Unlock()
			}
		}

	}
}

func (consumer *Consumer) Close() error {
	if consumer.udp_listener == nil {
		return nil
	}

	return consumer.udp_listener.Close()
}
