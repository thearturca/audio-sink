package sink

import (
	"context"
	"net"
)

type Client struct {
	udp_connection *net.UDPConn
	host           string
	port           int
	ctx            context.Context
}

func NewClient(ctx context.Context, host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
		ctx:  ctx,
	}
}

func (client *Client) Start() error {
	connection, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP(client.host), Port: client.port})

	if err != nil {
		return err
	}

	client.udp_connection = connection

	return nil
}

func (client *Client) Write(audio []byte) (n int, err error) {
	if client.udp_connection == nil {
		return 0, nil
	}

	n, err = client.udp_connection.Write(audio)

	if err != nil {
		client.ctx.Done()
	}

	return n, err
}

func (client *Client) Close() error {
	return client.udp_connection.Close()
}
