package audio_sink

import (
	"context"
	"fmt"
	"net"

	"github.com/gen2brain/malgo"
)

type Producer struct {
	udp_connection *net.UDPConn
	host           string
	port           int
	ctx            context.Context

	audioContext *malgo.AllocatedContext
	device       *malgo.Device
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

func (producer *Producer) InitAudio(deviceName string, deviceType string, sampleRate uint32, verbose bool) error {
	if producer.audioContext == nil {
		malgoCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
			if verbose {
				fmt.Println(message)
			}
		})

		if err != nil {
			return err
		}

		producer.audioContext = malgoCtx
	}

	if producer.device != nil {
		producer.device.Uninit()
	}

	var deviceConfig malgo.DeviceConfig

	if deviceName != "" && deviceType != "" {
		deviceInfo, malgoDeviceType, err := producer.findDevice(deviceName, deviceType)

		if err != nil {
			return err
		}

		deviceConfig = malgo.DefaultDeviceConfig(malgoDeviceType)
		deviceConfig.Capture.DeviceID = deviceInfo.ID.Pointer()
	} else if deviceType != "" {
		switch deviceType {
		case "Playback":
			deviceConfig = malgo.DefaultDeviceConfig(malgo.Loopback)
		case "Capture":
			deviceConfig = malgo.DefaultDeviceConfig(malgo.Capture)
			return fmt.Errorf("Unknown device type: %s", deviceType)
		}
	} else {
		deviceConfig = malgo.DefaultDeviceConfig(malgo.Capture)
	}

	deviceConfig.Capture.Format = malgo.FormatF32
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = sampleRate
	deviceConfig.Alsa.NoMMap = 1

	// This is the function that's used for sending more data to the device for playback.

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: producer.onSamples,
	}
	device, err := malgo.InitDevice(producer.audioContext.Context, deviceConfig, deviceCallbacks)

	if err != nil {
		return err
	}

	err = device.Start()

	if err != nil {
		return err
	}

	producer.device = device

	return nil
}

func (producer *Producer) findDevice(deviceName string, deviceType string) (malgo.DeviceInfo, malgo.DeviceType, error) {
	switch deviceType {
	case "Playback":
		devices, err := producer.audioContext.Devices(malgo.Loopback)

		if err != nil {
			return malgo.DeviceInfo{}, malgo.DeviceType(0), err
		}

		for _, device := range devices {
			if device.Name() == deviceName {
				return device, malgo.Loopback, nil
			}
		}
	case "Capture":
		devices, err := producer.audioContext.Devices(malgo.Capture)

		if err != nil {
			return malgo.DeviceInfo{}, malgo.DeviceType(0), err
		}

		for _, device := range devices {
			if device.Name() == deviceName {
				return device, malgo.Capture, nil
			}
		}
	default:
		return malgo.DeviceInfo{}, malgo.DeviceType(0), fmt.Errorf("invalid device type: %s. Device type must be Playback or Capture", deviceType)
	}

	return malgo.DeviceInfo{}, malgo.DeviceType(0), fmt.Errorf("device %s not found", deviceName)
}

func (producer *Producer) onSamples(_, pInputSamples []byte, _ uint32) {
	_, err := producer.Write(pInputSamples)

	if err != nil {
		fmt.Println(err)
	}
}

func (producer *Producer) Write(audio []byte) (n int, err error) {
	if producer.udp_connection == nil {
		return 0, fmt.Errorf("UDP connection is not initialized")
	}

	n, err = producer.udp_connection.Write(audio)

	if err != nil {
		return 0, err
	}

	return n, err
}

func (producer *Producer) Close() error {
	producer.device.Uninit()
	err := producer.audioContext.Uninit()

	if err != nil {
		return err
	}

	producer.audioContext.Free()

	err = producer.udp_connection.Close()

	if err != nil {
		return err
	}

	return nil
}
