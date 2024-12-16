package keys

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "strings"

    "github.com/baron-chain/cometbft-bc/libs/cli"
    "github.com/spf13/cobra"
    "sigs.k8s.io/yaml"
    sdk "github.com/baron-chain/cosmos-sdk/types"
    "github.com/baron-chain/cosmos-sdk/types/bech32"
)

const (
    flagFormat = "format"
)

// KeyOutput represents the key parsing output
type KeyOutput struct {
    HumanReadable string   `json:"human_readable,omitempty" yaml:"human_readable,omitempty"`
    HexBytes      string   `json:"hex_bytes,omitempty" yaml:"hex_bytes,omitempty"`
    Bech32Formats []string `json:"bech32_formats,omitempty" yaml:"bech32_formats,omitempty"`
}

func (ko KeyOutput) String() string {
    if ko.HumanReadable != "" {
        return fmt.Sprintf("Human readable part: %v\nBytes (hex): %s", ko.HumanReadable, ko.HexBytes)
    }
    out := make([]string, len(ko.Bech32Formats))
    for i, format := range ko.Bech32Formats {
        out[i] = fmt.Sprintf("  - %s", format)
    }
    return fmt.Sprintf("Bech32 Formats:\n%s", strings.Join(out, "\n"))
}

// ParseKeyStringCommand creates a command to parse address formats
func ParseKeyStringCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "parse <address>",
        Short: "Parse address between hex and bech32 formats",
        Long: `Convert addresses between hexadecimal and bech32 formats.
Supports both classic and quantum-safe address formats.`,
        Example: `$ baron-chain keys parse cosmos1...
$ baron-chain keys parse 0A090909...
$ baron-chain keys parse baron1...`,
        Args: cobra.ExactArgs(1),
        RunE: parseKey,
    }

    cmd.Flags().String(flagFormat, "text", "Output format (text|json|yaml)")
    return cmd
}

func parseKey(cmd *cobra.Command, args []string) error {
    addr := strings.TrimSpace(args[0])
    if addr == "" {
        return fmt.Errorf("empty address")
    }

    config, _ := sdk.GetSealedConfig(cmd.Context())
    output, _ := cmd.Flags().GetString(flagFormat)
    
    result, err := parseAddress(addr, config)
    if err != nil {
        return fmt.Errorf("failed to parse address: %w", err)
    }

    return displayOutput(cmd.OutOrStdout(), result, output)
}

func parseAddress(addr string, config *sdk.Config) (*KeyOutput, error) {
    // Try bech32 first
    if hrp, bz, err := bech32.DecodeAndConvert(addr); err == nil {
        return &KeyOutput{
            HumanReadable: hrp,
            HexBytes:      fmt.Sprintf("%X", bz),
        }, nil
    }

    // Try hex
    bz, err := hex.DecodeString(addr)
    if err != nil {
        return nil, fmt.Errorf("invalid address format: not bech32 or hex")
    }

    formats := make([]string, 0)
    for _, prefix := range getBech32Prefixes(config) {
        bech32Addr, err := bech32.ConvertAndEncode(prefix, bz)
        if err != nil {
            continue
        }
        formats = append(formats, bech32Addr)
    }

    return &KeyOutput{
        Bech32Formats: formats,
    }, nil
}

func getBech32Prefixes(config *sdk.Config) []string {
    return []string{
        config.GetBech32AccountAddrPrefix(),
        config.GetBech32AccountPubPrefix(),
        config.GetBech32ValidatorAddrPrefix(),
        config.GetBech32ValidatorPubPrefix(),
        config.GetBech32ConsensusAddrPrefix(),
        config.GetBech32ConsensusPubPrefix(),
    }
}

func displayOutput(w io.Writer, result *KeyOutput, format string) error {
    var (
        out []byte
        err error
    )

    switch strings.ToLower(format) {
    case "text", "yaml":
        out, err = yaml.Marshal(result)
    case "json":
        out, err = json.MarshalIndent(result, "", "  ")
    default:
        return fmt.Errorf("unsupported output format: %s", format)
    }

    if err != nil {
        return fmt.Errorf("failed to marshal output: %w", err)
    }

    _, err = fmt.Fprintln(w, string(out))
    return err
}
