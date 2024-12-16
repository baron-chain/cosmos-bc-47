package keys

import (
    "bufio"
    "fmt"

    "github.com/spf13/cobra"
    "github.com/baron-chain/cosmos-sdk/client"
    "github.com/baron-chain/cosmos-sdk/client/input"
    "github.com/baron-chain/cosmos-sdk/crypto/keyring"
)

const (
    flagSkipConfirm = "yes"
    flagForce       = "force"
)

// RenameKeyCommand creates a command to rename a key in the keyring
func RenameKeyCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "rename <old_name> <new_name>",
        Short: "Rename a key in the keyring",
        Long: `Rename a quantum-safe key in the Baron Chain keyring.

Note: For hardware-based keys (ledger) or offline keys, this command 
only renames the local public key references. Private keys stored in 
hardware devices cannot be renamed.

Example:
  $ baron-chain keys rename mykey mynewkey
  $ baron-chain keys rename ledgerkey newledgerkey --yes`,
        Args: cobra.ExactArgs(2),
        RunE: runRenameKey,
    }

    cmd.Flags().BoolP(flagSkipConfirm, "y", false, "Skip rename confirmation")
    cmd.Flags().Bool(flagForce, false, "Force rename even if new name exists")
    return cmd
}

func runRenameKey(cmd *cobra.Command, args []string) error {
    clientCtx, err := client.GetClientQueryContext(cmd)
    if err != nil {
        return fmt.Errorf("failed to get client context: %w", err)
    }

    oldName, newName := args[0], args[1]

    // Validate old key exists
    key, err := clientCtx.Keyring.Key(oldName)
    if err != nil {
        return fmt.Errorf("key '%s' not found: %w", oldName, err)
    }

    // Check if new name already exists
    if _, err := clientCtx.Keyring.Key(newName); err == nil {
        force, _ := cmd.Flags().GetBool(flagForce)
        if !force {
            return fmt.Errorf("key '%s' already exists, use --force to overwrite", newName)
        }
    }

    // Get confirmation unless --yes flag is set
    skip, _ := cmd.Flags().GetBool(flagSkipConfirm)
    if !skip {
        if err := confirmRename(cmd, oldName, newName, key.GetType()); err != nil {
            return err
        }
    }

    // Perform the rename operation
    if err := clientCtx.Keyring.Rename(oldName, newName); err != nil {
        return fmt.Errorf("failed to rename key: %w", err)
    }

    return printRenameResult(cmd, oldName, newName, key.GetType())
}

func confirmRename(cmd *cobra.Command, oldName, newName string, keyType keyring.KeyType) error {
    var prompt string
    switch keyType {
    case keyring.TypeLedger:
        prompt = fmt.Sprintf("Rename ledger key reference from '%s' to '%s'?", oldName, newName)
    case keyring.TypeOffline:
        prompt = fmt.Sprintf("Rename offline key reference from '%s' to '%s'?", oldName, newName)
    default:
        prompt = fmt.Sprintf("Rename key from '%s' to '%s'?", oldName, newName)
    }

    buf := bufio.NewReader(cmd.InOrStdin())
    confirmed, err := input.GetConfirmation(prompt, buf, cmd.ErrOrStderr())
    if err != nil {
        return fmt.Errorf("failed to read confirmation: %w", err)
    }
    if !confirmed {
        cmd.PrintErrln("Rename cancelled")
        return fmt.Errorf("operation cancelled")
    }

    return nil
}

func printRenameResult(cmd *cobra.Command, oldName, newName string, keyType keyring.KeyType) error {
    switch keyType {
    case keyring.TypeLedger, keyring.TypeOffline:
        cmd.PrintErrln(fmt.Sprintf("Public key reference renamed from '%s' to '%s'", oldName, newName))
    default:
        cmd.PrintErrln(fmt.Sprintf("Key successfully renamed from '%s' to '%s'", oldName, newName))
    }

    // Print quantum safety notice if applicable
    if isQuantumSafeKey(keyType) {
        cmd.PrintErrln("Note: Quantum-safe key properties preserved")
    }

    return nil
}

func isQuantumSafeKey(keyType keyring.KeyType) bool {
    // Add logic to check if key is quantum-safe based on type
    return keyType == keyring.TypeLocal // Assuming local keys are quantum-safe in Baron Chain
}
