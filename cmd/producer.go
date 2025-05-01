package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thearturca/audio-sink/pkg"
)

var producerCmd = &cobra.Command{
	Use:   "producer",
	Short: "producer allows you to send audio from a producer to a consumer",
	Long:  `producer allows you to send audio from a producer to a consumer using udp`,
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")

		if host == "" {
			fmt.Println("Host is required")
			os.Exit(1)
		}

		port := viper.GetInt("port")
		sampleRate := viper.GetUint32("sample_rate")

		ctx := context.Background()
		producer := audio_sink.NewProducer(ctx, host, port)

		err := producer.InitAudio(
			viper.GetString("device.name"),
			viper.GetString("device.type"),
			sampleRate,
			viper.GetBool("verbose"),
		)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = producer.Start()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer producer.Close()

		fmt.Printf("Streaming to %s:%d\n", host, port)
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	},
}
