package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func reportOk(jsonEnabled bool, command string, slot uint) {
	if jsonEnabled {
		out := map[string]any{
			"ok":      true,
			"command": command,
			"slot":    slot,
		}
		b, _ := json.Marshal(out)
		fmt.Println(string(b))
		return
	}
	fmt.Fprintf(os.Stderr, "%s slot %d: ok\n", command, slot)
}

func reportError(jsonEnabled bool, command string, slot uint, err error) {
	if jsonEnabled {
		out := map[string]any{
			"ok":      false,
			"command": command,
			"slot":    slot,
			"error":   err.Error(),
		}
		b, _ := json.Marshal(out)
		fmt.Println(string(b))
		return
	}
	fmt.Fprintf(os.Stderr, "%s error: %s\n", command, err)
}

func wrap(handler func(*cobra.Command, []string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		jsonEnabled, _ := cmd.Flags().GetBool("json")
		if err := handler(cmd, args); err != nil {
			printError(err, jsonEnabled)
			os.Exit(1)
		}
	}
}
