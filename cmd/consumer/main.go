package main

import (
	sink "audio-sink/internal"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gen2brain/malgo"
)

const sampleRate = 48000
const seconds = 0.1

func main() {
	log.Println("Starting consumer...")
	consumer := sink.NewConsumer(context.Background(), "0.0.0.0", 8080, sampleRate*seconds*4*2)

	err := consumer.Start()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer consumer.Close()

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = sampleRate
	deviceConfig.Alsa.NoMMap = 1

	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		_, err := io.ReadFull(consumer, pOutputSample)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onSamples,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Println("Listening...")
	fmt.Println("Press Enter to quit...")
	fmt.Scanln()
}
