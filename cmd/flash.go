package cmd

import (
	"fmt"
	"os"
	"stm32-bootctl/pkg/bcp"
	"stm32-bootctl/pkg/fwp"

	"github.com/spf13/cobra"
)

var flashCmd = &cobra.Command{
	Use:   "flash",
	Short: "Flash a firmware image into a slot",
	Run:   wrap(flash),
}

func init() {
	rootCmd.AddCommand(flashCmd)
	flashCmd.Flags().Uint("slot", 1, "Slot number")
	flashCmd.MarkFlagRequired("slot")
	flashCmd.Flags().String("file", "", "Firmware file destination")
	flashCmd.MarkFlagRequired("file")
}

func flash(cmd *cobra.Command, args []string) error {
	jsonEnabled, _ := cmd.Flags().GetBool("json")
	slot, _ := cmd.Flags().GetUint("slot")

	port, _ := cmd.Flags().GetString("port")
	baudrate, _ := cmd.Flags().GetUint("baud")

	path, _ := cmd.Flags().GetString("file")
	image, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	c, err := bcp.Open(port, uint16(baudrate))
	if err != nil {
		return fmt.Errorf("open port %s: %w", port, err)
	}
	defer c.Close()

	fwpClient, err := fwp.Open(port, int(baudrate))
	if err != nil {
		return err
	}
	defer fwpClient.Close()

	req, err := bcp.NewRequest(bcp.FlashCommand, []byte{uint8(slot)})
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	if err = c.Send(req); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	resp, err := c.Recv()
	if err != nil {
		return fmt.Errorf("recv: %w", err)
	}

	if resp.Command != req.Command {
		return fmt.Errorf("unexpected command in response: 0x%02X", uint8(resp.Command))
	}

	if !resp.IsOk() {
		return fmt.Errorf("flash failed: %s", resp.StatusName())
	}

	reportOk(jsonEnabled, "flash", slot)

	if err := fwpClient.Transfer(image, 5, nil); err != nil {
		return fmt.Errorf("firmware transfer failed: %v", err)
	}

	fmt.Println("Firmware transferred successfully")

	return nil
}
