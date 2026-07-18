/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"stm32-bootctl/pkg/bcp"

	"github.com/spf13/cobra"
)

// flashCmd represents the flash command
var flashCmd = &cobra.Command{
	Use:   "flash",
	Short: "Flash a firmware image into a slot",
	Long:  "Flash a firmware image into a slot",
	Run:   flash,
}

func init() {
	rootCmd.AddCommand(flashCmd)
	flashCmd.Flags().Uint("slot", 1, "Slot number")
	flashCmd.MarkFlagRequired("slot")
}

func flash(cmd *cobra.Command, args []string) {
	jsonEnabled, _ := cmd.Flags().GetBool("json")
	slot, _ := cmd.Flags().GetUint("slot")

	var err error
	defer func() {
		if err != nil {
			flashReportError(jsonEnabled, slot, err)
			os.Exit(1)
		}
		flashReportOk(jsonEnabled, slot)
	}()

	port, _ := cmd.Flags().GetString("port")
	baudrate, _ := cmd.Flags().GetUint("baud")

	c, err := bcp.Open(port, uint16(baudrate))
	if err != nil {
		err = fmt.Errorf("open port %s: %w", port, err)
		return
	}
	defer c.Close()

	req, err := bcp.NewRequest(bcp.FlashCommand, []byte{uint8(slot)})
	if err != nil {
		err = fmt.Errorf("build request: %w", err)
		return
	}

	if err = c.Send(req); err != nil {
		err = fmt.Errorf("send: %w", err)
		return
	}

	resp, err := c.Recv()
	if err != nil {
		err = fmt.Errorf("recv: %w", err)
		return
	}

	if resp.Command != req.Command {
		err = fmt.Errorf("unexpected command in response: 0x%02X", uint8(resp.Command))
		return
	}

	if !resp.IsOk() {
		err = fmt.Errorf("flash failed: %s", resp.StatusName())
		return
	}
}

func flashReportOk(jsonEnabled bool, slot uint) {
	if jsonEnabled {
		out := map[string]any{
			"ok":      true,
			"command": "flash",
			"slot":    slot,
		}
		b, _ := json.Marshal(out)
		fmt.Println(string(b))
		return
	}
	fmt.Fprintf(os.Stderr, "flash slot %d: ok\n", slot)
}

func flashReportError(jsonEnabled bool, slot uint, err error) {
	if jsonEnabled {
		out := map[string]any{
			"ok":      false,
			"command": "flash",
			"slot":    slot,
			"error":   err.Error(),
		}
		b, _ := json.Marshal(out)
		fmt.Println(string(b))
		return
	}
	fmt.Fprintln(os.Stderr, "error:", err)
}
