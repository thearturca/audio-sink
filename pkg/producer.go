package sink

import (
	"context"
	"net"
)

type Producer struct {
	udp_connection *net.UDPConn
	host           string
	port           int
	ctx            context.Context
}

func NewProducer(ctx context.Context, host string, port int) *Producer {
	return &Producer{
		host: host,
		port: port,
		ctx:  ctx,
	}
}

func (producer *Producer) Start() error {
	connection, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP(producer.host), Port: producer.port})

	if err != nil {
		return err
	}

	producer.udp_connection = connection

	return nil
}

func (producer *Producer) Write(audio []byte) (n int, err error) {
	if producer.udp_connection == nil {
		return 0, nil
	}

	n, err = producer.udp_connection.Write(audio)

	if err != nil {
		producer.ctx.Done()
	}

	return n, err
}

func (producer *Producer) Close() error {
	return producer.udp_connection.Close()
}
