package keys

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/baron-chain/cosmos-sdk/client"
    "github.com/baron-chain/cosmos-sdk/crypto/keyring"
)

const (
    flagListNames = "list-names"
    flagShowAlgo  = "show-algorithm"
)

func ListKeysCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List all quantum-safe keys",
        Long:  `List all quantum-safe public keys stored in the keyring with their names, addresses and algorithms.`,
        RunE:  runListCmd,
    }

    cmd.Flags().BoolP(flagListNames, "n", false, "List names only")
    cmd.Flags().BoolP(flagShowAlgo, "a", false, "Show key algorithm (kyber/dilithium)")
    return cmd
}

func runListCmd(cmd *cobra.Command, _ []string) error {
    clientCtx, err := client.GetClientQueryContext(cmd)
    if err != nil {
        return fmt.Errorf("failed to get client context: %w", err)
    }

    records, err := clientCtx.Keyring.List()
    if err != nil {
        return fmt.Errorf("failed to list keys: %w", err)
    }

    if len(records) == 0 {
        cmd.Printf("No quantum-safe keys found in keyring\n")
        return nil
    }

    showNames, _ := cmd.Flags().GetBool(flagListNames)
    showAlgo, _ := cmd.Flags().GetBool(flagShowAlgo)

    if showNames {
        return printKeyNames(cmd, records)
    }

    return printKeyDetails(cmd, records, showAlgo, clientCtx.OutputFormat)
}

func printKeyNames(cmd *cobra.Command, records []*keyring.Record) error {
    for _, k := range records {
        cmd.Printf("%s\n", k.Name)
    }
    return nil
}

func printKeyDetails(cmd *cobra.Command, records []*keyring.Record, showAlgo bool, format string) error {
    if format == "json" {
        return printJSON(cmd, records, showAlgo)
    }

    for _, k := range records {
        if showAlgo {
            cmd.Printf("Name: %s\nAddress: %s\nAlgorithm: %s\n\n", 
                k.Name, k.GetAddress(), getKeyAlgorithm(k))
        } else {
            cmd.Printf("Name: %s\nAddress: %s\n\n", 
                k.Name, k.GetAddress())
        }
    }
    return nil
}

func getKeyAlgorithm(record *keyring.Record) string {
    switch record.PubKey.Type() {
    case "kyber":
        return "Kyber (Quantum-Safe)"
    case "dilithium":
        return "Dilithium (Quantum-Safe)"
    default:
        return "Unknown"
    }
}

func ListKeyTypesCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "list-key-types",
        Short: "List supported quantum-safe key types",
        Long:  `Display all supported quantum-safe key algorithms in Baron Chain`,
        RunE: func(cmd *cobra.Command, _ []string) error {
            clientCtx, err := client.GetClientQueryContext(cmd)
            if err != nil {
                return fmt.Errorf("failed to get client context: %w", err)
            }

            cmd.Println("Supported quantum-safe key algorithms:")
            cmd.Println("- kyber     (Post-Quantum Key Encapsulation)")
            cmd.Println("- dilithium (Post-Quantum Digital Signatures)")
            
            return nil
        },
    }
}

func printJSON(cmd *cobra.Command, records []*keyring.Record, showAlgo bool) error {
    type keyInfo struct {
        Name      string `json:"name"`
        Address   string `json:"address"`
        Algorithm string `json:"algorithm,omitempty"`
    }

    var output []keyInfo
    for _, k := range records {
        info := keyInfo{
            Name:    k.Name,
            Address: k.GetAddress().String(),
        }
        if showAlgo {
            info.Algorithm = getKeyAlgorithm(k)
        }
        output = append(output, info)
    }

    return printKeyringRecords(cmd.OutOrStdout(), output, "json")
}
