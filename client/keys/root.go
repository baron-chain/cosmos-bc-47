package keys

import (
    "github.com/spf13/cobra"
    "github.com/baron-chain/cometbft-bc/libs/cli"
    "github.com/baron-chain/cosmos-sdk/client/flags"
)

const (
    defaultOutputFormat = "text"
    keyringCmdName     = "keys"
)

// Commands registers all key management commands for Baron Chain
func Commands(defaultNodeHome string) *cobra.Command {
    cmd := &cobra.Command{
        Use:   keyringCmdName,
        Short: "Manage quantum-safe keys",
        Long: `Baron Chain quantum-safe key management commands.

Available Key Types:
    kyber       Post-quantum key encapsulation mechanism (KEM)
    dilithium   Post-quantum digital signature algorithm

Supported Keyring Backends:
    os          Operating system's secure credential store
    file        Encrypted file-based keystore in app's config directory
    kwallet     KDE Wallet Manager (requires external setup)
    pass        Unix pass command line utility (requires GnuPG)
    test        Insecure disk storage (testing only)

For external backend setup:
    KWallet: https://github.com/KDE/kwallet
    Pass:    https://www.passwordstore.org/
    GnuPG:   https://gnupg.org/

Note: File backend will prompt for password on each access.`,
    }

    // Add key management commands
    cmd.AddCommand(
        // Key Generation
        MnemonicKeyCommand(),
        AddKeyCommand(),
        
        // Key Import/Export
        ImportKeyCommand(),
        ImportKeyHexCommand(),
        ExportKeyCommand(),
        
        // Key Management
        ListKeysCmd(),
        ShowKeysCmd(),
        RenameKeyCommand(),
        DeleteKeyCommand(),
        
        // Utility Commands
        ListKeyTypesCmd(),
        ParseKeyStringCommand(),
        MigrateCommand(),
    )

    // Add persistent flags
    addPersistentFlags(cmd, defaultNodeHome)

    return cmd
}

func addPersistentFlags(cmd *cobra.Command, defaultNodeHome string) {
    persistentFlags := cmd.PersistentFlags()

    persistentFlags.String(
        flags.FlagHome,
        defaultNodeHome,
        "Application home directory for key storage",
    )

    persistentFlags.String(
        cli.OutputFlag,
        defaultOutputFormat,
        "Output format (text|json)",
    )

    // Add keyring-specific flags
    flags.AddKeyringFlags(persistentFlags)

    // Add quantum-safe specific flags
    addQuantumSafeFlags(persistentFlags)
}

func addQuantumSafeFlags(flags *pflag.FlagSet) {
    flags.String(
        flagKeyAlgorithm,
        defaultAlgorithm,
        "Quantum-safe algorithm (kyber|dilithium)",
    )
    
    flags.Int(
        flagEntropySize,
        defaultEntropySize,
        "Entropy size in bits for quantum-safe key generation",
    )

    flags.Bool(
        flagForceQuantum,
        true,
        "Enforce quantum-safe key generation",
    )
}

// GetCommands returns all available key commands
func GetCommands() []*cobra.Command {
    return []*cobra.Command{
        MnemonicKeyCommand(),
        AddKeyCommand(),
        ImportKeyCommand(),
        ImportKeyHexCommand(),
        ExportKeyCommand(),
        ListKeysCmd(),
        ShowKeysCmd(),
        RenameKeyCommand(),
        DeleteKeyCommand(),
        ListKeyTypesCmd(),
        ParseKeyStringCommand(),
        MigrateCommand(),
    }
}
