package cmd

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"stm32-bootctl/pkg/bcp"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "",
	Long:  "",
	Run:   wrap(status),
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Uint("slot", 0, "Slot number (0 = all slots)")
}

// func status(cmd *cobra.Command, args []string) {
// 	jsonEnabled, _ := cmd.Flags().GetBool("json")
// 	if err := runStatus(cmd); err != nil {
// 		printError(err, jsonEnabled)
// 		os.Exit(1)
// 	}
// }

func status(cmd *cobra.Command, args []string) error {
	jsonEnabled, _ := cmd.Flags().GetBool("json")
	slot, _ := cmd.Flags().GetUint("slot")
	port, _ := cmd.Flags().GetString("port")
	baudrate, _ := cmd.Flags().GetUint("baud")

	c, err := bcp.Open(port, uint16(baudrate))
	if err != nil {
		return fmt.Errorf("open port %s: %w", port, err)
	}
	defer c.Close()

	if slot == 0 {
		for _, s := range []uint8{1, 2} {
			st, err := verifySlot(c, s)
			if err != nil {
				return &slotError{slot: s, err: err}
			}
			printImageStatus(st, jsonEnabled)
		}
		return nil
	}

	st, err := verifySlot(c, uint8(slot))
	if err != nil {
		return &slotError{slot: uint8(slot), err: err}
	}
	printImageStatus(st, jsonEnabled)
	return nil
}

type ImageStatus struct {
	Slot         uint8
	IsValid      bool
	VersionMajor *uint8
	VersionMinor *uint8
	VersionPatch *uint8
	Crc          *uint32
	Size         *uint32
}

func verifySlot(c *bcp.Client, slot uint8) (ImageStatus, error) {
	req, err := bcp.NewRequest(bcp.VerifyCommand, []byte{slot})
	if err != nil {
		return ImageStatus{}, nil
	}

	if err = c.Send(req); err != nil {
		return ImageStatus{}, nil
	}

	resp, err := c.Recv()
	if err != nil {
		return ImageStatus{}, err
	}

	if resp.Command != bcp.VerifyCommand {
		err = fmt.Errorf("unexpected command in response: 0x%02X", uint8(resp.Command))
		return ImageStatus{}, err
	}

	if len(resp.Data) != 12 {
		return ImageStatus{}, fmt.Errorf("incorrect payload length: expected 12 bytes, got %d", len(resp.Data))
	}

	switch resp.Data[0] {
	case 0:
		return ImageStatus{
			Slot:    slot,
			IsValid: false,
		}, nil
	case 1:
		res := ImageStatus{
			Slot:    slot,
			IsValid: true,
		}
		if len(resp.Data) < 12 {
			return ImageStatus{}, fmt.Errorf("short verify response: %d bytes", len(resp.Data))
		}

		major := resp.Data[1]
		minor := resp.Data[2]
		patch := resp.Data[3]
		crc := binary.BigEndian.Uint32(resp.Data[4:8])
		size := binary.BigEndian.Uint32(resp.Data[8:12])

		res.IsValid = true
		res.VersionMajor = &major
		res.VersionMinor = &minor
		res.VersionPatch = &patch
		res.Crc = &crc
		res.Size = &size

		return res, nil
	default:
		return ImageStatus{}, fmt.Errorf("incorrect first byte of payload")
	}
}

func printImageStatus(res ImageStatus, jsonEnabled bool) {
	if jsonEnabled {
		var out map[string]any
		if !res.IsValid {
			out = map[string]any{"slot": res.Slot, "valid": false}
		} else {
			out = map[string]any{
				"slot":  res.Slot,
				"valid": true,
				"version": map[string]any{
					"major":  *res.VersionMajor,
					"minor":  *res.VersionMinor,
					"patch":  *res.VersionPatch,
					"string": fmt.Sprintf("%d.%d.%d", *res.VersionMajor, *res.VersionMinor, *res.VersionPatch),
				},
				"crc":     *res.Crc,
				"crc_hex": fmt.Sprintf("0x%08X", *res.Crc),
				"size":    *res.Size,
			}
		}
		b, _ := json.Marshal(out)
		fmt.Println(string(b))
		return
	}

	fmt.Printf("Slot %d:\n", res.Slot)
	if !res.IsValid {
		fmt.Println("  empty / invalid image")
		return
	}
	fmt.Printf("  version : %d.%d.%d\n", *res.VersionMajor, *res.VersionMinor, *res.VersionPatch)
	fmt.Printf("  size    : %d bytes\n", *res.Size)
	fmt.Printf("  CRC32   : 0x%08X\n", *res.Crc)
}

type slotError struct {
	slot uint8
	err  error
}

func (e *slotError) Error() string {
	return fmt.Sprintf("slot %d: %v", e.slot, e.err)
}

func (e *slotError) Unwrap() error {
	return e.err
}

func printError(err error, jsonEnabled bool) {
	if jsonEnabled {
		out := map[string]any{}
		var se *slotError
		if errors.As(err, &se) {
			out["slot"] = se.slot
			out["error"] = se.err.Error()
		} else {
			out["error"] = err.Error()
		}
		b, _ := json.Marshal(out)
		fmt.Fprintln(os.Stderr, string(b))
		return
	}
	fmt.Fprintln(os.Stderr, "error: "+err.Error())
}
