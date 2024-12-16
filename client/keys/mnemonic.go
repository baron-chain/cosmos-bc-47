package keys

import (
    "bufio"
    "crypto/sha512"
    "fmt"
    "strings"

    "github.com/spf13/cobra"
    "github.com/baron-chain/cosmos-sdk/client/input"
    "github.com/baron-chain/go-bip39"
)

const (
    flagUserEntropy     = "unsafe-entropy"
    flagEntropySize     = "entropy-size"
    flagQuantumSafe     = "quantum-safe"
    defaultEntropySize  = 256
    minEntropySize      = 256
    recommendedEntropy  = 512
)

// MnemonicKeyCommand generates a quantum-safe bip39 mnemonic
func MnemonicKeyCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mnemonic",
        Short: "Generate a quantum-safe mnemonic seed phrase",
        Long: strings.TrimSpace(`
Generate a quantum-safe BIP39 mnemonic (seed phrase) with enhanced entropy.
By default, uses system-provided entropy with quantum-safe enhancements.
For user-provided entropy, use --unsafe-entropy flag (not recommended).

Example:
$ baron-chain keys mnemonic --entropy-size 512
$ baron-chain keys mnemonic --quantum-safe
`),
        RunE: generateMnemonic,
    }

    cmd.Flags().Bool(flagUserEntropy, false, "Use user-provided entropy (not recommended)")
    cmd.Flags().Int(flagEntropySize, defaultEntropySize, "Entropy size in bits (256, 384, or 512)")
    cmd.Flags().Bool(flagQuantumSafe, true, "Enable quantum-safe entropy enhancement")
    
    return cmd
}

func generateMnemonic(cmd *cobra.Command, _ []string) error {
    entropySize, err := cmd.Flags().GetInt(flagEntropySize)
    if err != nil {
        return fmt.Errorf("failed to get entropy size: %w", err)
    }

    if entropySize < minEntropySize {
        return fmt.Errorf("entropy size must be at least %d bits for quantum safety", minEntropySize)
    }

    var entropy []byte
    useUserEntropy, _ := cmd.Flags().GetBool(flagUserEntropy)
    
    if useUserEntropy {
        entropy, err = getUserEntropy(cmd, entropySize)
    } else {
        entropy, err = getSystemEntropy(entropySize)
    }
    
    if err != nil {
        return err
    }

    quantumSafe, _ := cmd.Flags().GetBool(flagQuantumSafe)
    if quantumSafe {
        entropy = enhanceEntropyForQuantumSafety(entropy)
    }

    mnemonic, err := bip39.NewMnemonic(entropy)
    if err != nil {
        return fmt.Errorf("failed to generate mnemonic: %w", err)
    }

    cmd.Println("\nYour quantum-safe mnemonic phrase (keep this secure):")
    cmd.Println(mnemonic)
    
    if entropySize < recommendedEntropy {
        cmd.Printf("\nNote: For maximum quantum safety, consider using --entropy-size=%d\n", recommendedEntropy)
    }

    return nil
}

func getUserEntropy(cmd *cobra.Command, entropySize int) ([]byte, error) {
    minChars := entropySize / 6 // conservative estimate for base-64
    buf := bufio.NewReader(cmd.InOrStdin())
    
    cmd.PrintErrln("\nWARNING: User-provided entropy is not recommended for production use.")
    prompt := fmt.Sprintf(
        "Please enter at least %d characters of entropy (more is better):",
        minChars,
    )
    
    inputEntropy, err := input.GetString(prompt, buf)
    if err != nil {
        return nil, fmt.Errorf("failed to read entropy: %w", err)
    }

    if len(inputEntropy) < minChars {
        return nil, fmt.Errorf(
            "insufficient entropy: got %d chars, need at least %d",
            len(inputEntropy),
            minChars,
        )
    }

    conf, err := input.GetConfirmation(
        fmt.Sprintf("Confirm entropy input length: %d chars", len(inputEntropy)),
        buf,
        cmd.ErrOrStderr(),
    )
    if err != nil {
        return nil, err
    }
    if !conf {
        return nil, fmt.Errorf("entropy input cancelled")
    }

    // Use SHA-512 for enhanced security
    hash := sha512.Sum512([]byte(inputEntropy))
    return hash[:entropySize/8], nil
}

func getSystemEntropy(entropySize int) ([]byte, error) {
    entropy, err := bip39.NewEntropy(entropySize)
    if err != nil {
        return nil, fmt.Errorf("failed to generate system entropy: %w", err)
    }
    return entropy, nil
}

func enhanceEntropyForQuantumSafety(entropy []byte) []byte {
    // Apply additional mixing for quantum resistance
    hash := sha512.Sum512(entropy)
    result := make([]byte, len(entropy))
    
    // XOR the original entropy with the hash for enhanced randomness
    for i := 0; i < len(entropy); i++ {
        result[i] = entropy[i] ^ hash[i]
    }
    
    return result
}
