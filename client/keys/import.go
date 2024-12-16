package keys

import (
    "bufio"
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/baron-chain/cosmos-sdk/client"
    "github.com/baron-chain/cosmos-sdk/client/flags"
    "github.com/baron-chain/cosmos-sdk/client/input"
    "github.com/baron-chain/cosmos-sdk/crypto/keyring"
    "github.com/baron-chain/cometbft-bc/crypto/kyber"
)

const (
    flagKeyAlgorithm = "key-algorithm"
    defaultAlgorithm = "kyber"
)

func ImportKeyCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "import <name> <keyfile>",
        Short: "Import quantum-safe private keys",
        Long:  "Import a quantum-safe private key (Kyber/Dilithium supported) into the local keybase",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx, err := client.GetClientQueryContext(cmd)
            if err != nil {
                return fmt.Errorf("failed to get client context: %w", err)
            }

            keyBytes, err := os.ReadFile(args[1])
            if err != nil {
                return fmt.Errorf("failed to read keyfile: %w", err)
            }

            passphrase, err := input.GetPassword("Enter passphrase:", bufio.NewReader(clientCtx.Input))
            if err != nil {
                return fmt.Errorf("failed to read passphrase: %w", err)
            }

            algorithm, _ := cmd.Flags().GetString(flagKeyAlgorithm)
            return importKey(clientCtx.Keyring, args[0], keyBytes, passphrase, algorithm)
        },
    }

    cmd.Flags().String(flagKeyAlgorithm, defaultAlgorithm, "Quantum-safe algorithm (kyber/dilithium)")
    return cmd
}

func importKey(kr keyring.Keyring, name string, keyBytes []byte, passphrase, algorithm string) error {
    switch algorithm {
    case "kyber":
        return importKyberKey(kr, name, keyBytes, passphrase)
    case "dilithium":
        return importDilithiumKey(kr, name, keyBytes, passphrase)
    default:
        return fmt.Errorf("unsupported key algorithm: %s", algorithm)
    }
}

func importKyberKey(kr keyring.Keyring, name string, keyBytes []byte, passphrase string) error {
    key, err := kyber.DecryptPrivateKey(keyBytes, []byte(passphrase))
    if err != nil {
        return fmt.Errorf("failed to decrypt Kyber key: %w", err)
    }

    return kr.ImportPrivKey(name, key.String(), passphrase)
}

func ImportHexCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "import-hex <name> <hex>",
        Short: "Import quantum-safe hex keys",
        Long:  "Import hex encoded quantum-safe private key (Kyber/Dilithium supported)",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx, err := client.GetClientQueryContext(cmd)
            if err != nil {
                return fmt.Errorf("failed to get client context: %w", err)
            }

            algorithm, _ := cmd.Flags().GetString(flagKeyAlgorithm)
            return importHexKey(clientCtx.Keyring, args[0], args[1], algorithm)
        },
    }

    cmd.Flags().String(flagKeyAlgorithm, defaultAlgorithm, "Quantum-safe algorithm (kyber/dilithium)")
    return cmd
}

func importHexKey(kr keyring.Keyring, name, hexKey, algorithm string) error {
    switch algorithm {
    case "kyber":
        return kr.ImportPrivKeyHex(name, hexKey, "kyber")
    case "dilithium":
        return kr.ImportPrivKeyHex(name, hexKey, "dilithium")
    default:
        return fmt.Errorf("unsupported key algorithm: %s", algorithm)
    }
}
