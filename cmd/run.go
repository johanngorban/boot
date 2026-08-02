package cmd

import (
	"fmt"
	"stm32-bootctl/pkg/bcp"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "",
	Run:   wrap(run),
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().Uint("slot", 1, "Slot number")
	runCmd.MarkFlagRequired("slot")
}

func run(cmd *cobra.Command, args []string) error {
	jsonEnabled, _ := cmd.Flags().GetBool("json")
	port, _ := cmd.Flags().GetString("port")
	baudrate, _ := cmd.Flags().GetInt("baud")
	slot, _ := cmd.Flags().GetUint("slot")
	if slot > 255 {
		return fmt.Errorf("slot %d out of range (0-255)", slot)
	}

	c, err := bcp.Open(port, baudrate)
	if err != nil {
		return fmt.Errorf("open port %s: %w", port, err)
	}
	defer c.Close()

	req, err := bcp.NewRequest(bcp.RunCommand, []byte{uint8(slot)})
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

	if resp.IsOk() {
		if jsonEnabled {
			fmt.Printf("slot #%d starting", slot)
		} else {
			fmt.Printf("Starting slot #%d\n", slot)
		}
	} else {
		return fmt.Errorf("run failed: %s", resp.StatusName())
	}

	return nil
}
