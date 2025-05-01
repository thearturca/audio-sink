package cmd

import (
	"fmt"

	"github.com/gen2brain/malgo"
	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "devices allows you to list audio devices",
	Long:  `devices allows you to list audio devices`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {})

		cobra.CheckErr(err)

		defer func() {
			_ = ctx.Uninit()
			ctx.Free()
		}()

		devices, err := ctx.Devices(malgo.Capture)
		cobra.CheckErr(err)

		fmt.Printf("Capture devices:\n%s\n\n", formatDevices(devices))

		devices, err = ctx.Devices(malgo.Playback)
		cobra.CheckErr(err)

		fmt.Printf("Playback devices:\n%s\n\n", formatDevices(devices))
	},
}

func formatDevices(devices []malgo.DeviceInfo) string {

	var result string

	for _, device := range devices {
		result += fmt.Sprintf("%s\n", device.Name())
	}

	return result
}
