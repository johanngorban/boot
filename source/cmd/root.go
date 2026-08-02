package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "boot",
	Short: "Utility for stm32-bootloader",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("json", false, "Enable JSON output")
	rootCmd.PersistentFlags().String("port", "/dev/ttyACM0", "Device port")
	rootCmd.PersistentFlags().Int("baud", 115200, "Baudrate")
}
