package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thearturca/audio-sink/pkg"
)

var consumerCmd = &cobra.Command{
	Use:   "consumer",
	Short: "consumer allows you to receive audio from a producer",
	Long:  `consumer allows you to receive audio from a producer using udp`,
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")

		if host == "" {
			host = "0.0.0.0"
		}

		port := viper.GetInt("port")
		sampleRate := viper.GetUint32("sample_rate")

		log.Println("Starting consumer...")

		ctx := context.Background()
		consumer := audio_sink.NewConsumer(ctx, host, port)

		err := consumer.InitAudio(
			viper.GetString("device.name"),
			sampleRate,
			viper.GetBool("verbose"),
		)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = consumer.Start()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer consumer.Close()

		fmt.Printf("Listening on %s:%d\n", host, port)
		fmt.Println("Press Enter to quit...")
		fmt.Scanln()
	},
}
