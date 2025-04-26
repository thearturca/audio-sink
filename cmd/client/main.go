package main

import (
	sink "audio-sink/internal"
	"context"
	"fmt"
	"os"

	"github.com/gen2brain/malgo"
)

const sampleRate = 48000
const seconds = 0.1

func main() {
	client := sink.NewClient(context.Background(), "192.168.88.254", 8080)
	err := client.Start()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer client.Close()

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

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatF32
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = sampleRate
	deviceConfig.Alsa.NoMMap = 1

	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		_, err := client.Write(pInputSamples)

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

	fmt.Println("Press enter to exit...")
	fmt.Scanln()
}
