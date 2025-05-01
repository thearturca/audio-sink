package audio_sink

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"sync"

	"github.com/gen2brain/malgo"
)

type ProducerClient struct {
	address *net.UDPAddr
	audio   []byte
	mutex   sync.Mutex
}

func NewConsumer(ctx context.Context, host string, port int) *Consumer {
	return &Consumer{
		host:    host,
		port:    port,
		clients: make(map[string]*ProducerClient),
		ctx:     ctx,
	}
}

type Consumer struct {
	udp_listener *net.UDPConn
	clients      map[string]*ProducerClient
	host         string
	port         int
	ctx          context.Context
	bufferSize   int

	audioContext *malgo.AllocatedContext
	device       *malgo.Device
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

func (consumer *Consumer) InitAudio(deviceName string, sampleRate uint32, verbose bool) error {
	consumer.bufferSize = int(sampleRate * 2)

	if consumer.audioContext == nil {
		malgoCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
			if verbose {
				fmt.Println(message)
			}
		})

		if err != nil {
			return err
		}

		consumer.audioContext = malgoCtx
	}

	if consumer.device != nil {
		consumer.device.Uninit()
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)

	if deviceName != "" {
		deviceInfo, err := consumer.findDeviceId(deviceName)

		if err != nil {
			return err
		}

		deviceConfig.Playback.DeviceID = deviceInfo.ID.Pointer()
	}

	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = sampleRate
	deviceConfig.Alsa.NoMMap = 1

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: consumer.onSamples,
	}
	device, err := malgo.InitDevice(
		consumer.audioContext.Context,
		deviceConfig,
		deviceCallbacks,
	)

	if err != nil {
		return err
	}
	err = device.Start()

	if err != nil {
		return err
	}

	consumer.device = device

	return nil
}

func (consumer *Consumer) findDeviceId(deviceName string) (malgo.DeviceInfo, error) {
	devices, err := consumer.audioContext.Devices(malgo.Playback)

	if err != nil {
		return malgo.DeviceInfo{}, err
	}

	for _, device := range devices {
		if device.Name() == deviceName {
			return device, nil
		}
	}

	return malgo.DeviceInfo{}, fmt.Errorf("device %s not found", deviceName)
}

func (consumer *Consumer) onSamples(pOutputSample, _ []byte, _ uint32) {
	_, err := io.ReadFull(consumer, pOutputSample)

	if err != nil {
		fmt.Println(err)
	}
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
	consumer.device.Uninit()
	err := consumer.audioContext.Uninit()

	if err != nil {
		return err
	}
	consumer.audioContext.Free()

	if consumer.udp_listener == nil {
		return nil
	}

	err = consumer.udp_listener.Close()

	if err != nil {
		return err
	}

	return nil
}
