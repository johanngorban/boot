package cmd

import (
	"boot-util/pkg/bcp"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "",
	Run:   wrap(version),
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func version(cmd *cobra.Command, args []string) error {
	jsonEnabled, _ := cmd.Flags().GetBool("json")
	port, _ := cmd.Flags().GetString("port")
	baudrate, _ := cmd.Flags().GetInt("baud")

	c, err := bcp.Open(port, baudrate)
	if err != nil {
		return fmt.Errorf("open port %s: %w", port, err)
	}
	defer c.Close()

	req, err := bcp.NewRequest(bcp.VersionCommand, nil)
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
		return fmt.Errorf("version fetching failed: %s", resp.StatusName())
	}

	if len(resp.Data) != 3 {
		return fmt.Errorf("incorrect payload data: expected 3 bytes, got %d", len(resp.Data))
	}

	major := resp.Data[0]
	minor := resp.Data[1]
	patch := resp.Data[2]

	if !jsonEnabled {
		fmt.Printf("boot-utilloader version: v%d.%d.%d\n", major, minor, patch)
	} else {
		out := map[string]any{
			"version": map[string]any{
				"major":  major,
				"minor":  minor,
				"patch":  patch,
				"string": fmt.Sprintf("v%d.%d.%d", major, minor, patch),
			},
		}

		bytes, _ := json.Marshal(out)
		fmt.Println(string(bytes))
	}

	return nil
}
