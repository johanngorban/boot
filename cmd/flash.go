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
	flashCmd.Flags().Bool("no-verify", false, "Skip post-flash verification")
}

func flash(cmd *cobra.Command, args []string) error {
	jsonEnabled, _ := cmd.Flags().GetBool("json")
	slot, _ := cmd.Flags().GetUint("slot")
	if slot > 255 {
		return fmt.Errorf("slot %d out of range (0-255)", slot)
	}

	port, _ := cmd.Flags().GetString("port")
	baudrate, _ := cmd.Flags().GetInt("baud")
	noVerify, _ := cmd.Flags().GetBool("no-verify")

	path, _ := cmd.Flags().GetString("file")
	image, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(image) == 0 {
		return fmt.Errorf("firmware file %s is empty", path)
	}

	if err := requestFlash(port, baudrate, uint8(slot)); err != nil {
		return err
	}

	var progress func(int, int)
	if !jsonEnabled {
		progress = func(done, total int) {
			fmt.Fprintf(os.Stderr, "\rTransferring %d/%d bytes (%d%%)", done, total, done*100/total)
			if done >= total {
				fmt.Fprintln(os.Stderr)
			}
		}
	}

	if err := transferFirmware(port, baudrate, image, progress); err != nil {
		return err
	}

	if !noVerify {
		if err := verifyFlashed(port, baudrate, uint8(slot)); err != nil {
			return err
		}
	}

	reportOk(jsonEnabled, "flash", slot)

	return nil
}

func requestFlash(port string, baudrate int, slot uint8) error {
	c, err := bcp.Open(port, baudrate)
	if err != nil {
		return fmt.Errorf("open port %s: %w", port, err)
	}
	defer c.Close()

	req, err := bcp.NewRequest(bcp.FlashCommand, []byte{slot})
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if err := c.Send(req); err != nil {
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
	return nil
}

func transferFirmware(port string, baudrate int, image []byte, progress func(int, int)) error {
	c, err := fwp.Open(port, baudrate)
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.Transfer(image, 5, progress); err != nil {
		return fmt.Errorf("firmware transfer failed: %w", err)
	}
	return nil
}

func verifyFlashed(port string, baudrate int, slot uint8) error {
	c, err := bcp.Open(port, baudrate)
	if err != nil {
		return fmt.Errorf("open port %s: %w", port, err)
	}
	defer c.Close()

	st, err := verifySlot(c, slot)
	if err != nil {
		return fmt.Errorf("verify failed: %w", err)
	}
	if !st.IsValid {
		return fmt.Errorf("post-flash verify reports image as invalid")
	}
	return nil
}
