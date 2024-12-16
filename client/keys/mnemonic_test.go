package keys

import (
    "fmt"
    "strings"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/baron-chain/cosmos-sdk/testutil"
)

func TestMnemonicGeneration(t *testing.T) {
    testCases := []struct {
        name          string
        args          []string
        input         string
        expectError   bool
        errorMessage  string
        entropySize   int
        quantumSafe   bool
    }{
        {
            name:        "default quantum-safe generation",
            args:        []string{},
            expectError: false,
            entropySize: defaultEntropySize,
            quantumSafe: true,
        },
        {
            name:        "512-bit quantum-safe entropy",
            args:        []string{fmt.Sprintf("--%s=512", flagEntropySize)},
            expectError: false,
            entropySize: 512,
            quantumSafe: true,
        },
        {
            name:         "insufficient entropy size",
            args:         []string{fmt.Sprintf("--%s=128", flagEntropySize)},
            expectError:  true,
            errorMessage: fmt.Sprintf("entropy size must be at least %d bits for quantum safety", minEntropySize),
        },
        {
            name: "user entropy with no input",
            args: []string{fmt.Sprintf("--%s=1", flagUserEntropy)},
            expectError: true,
            errorMessage: "EOF",
        },
        {
            name:         "user entropy too short",
            args:         []string{fmt.Sprintf("--%s=1", flagUserEntropy)},
            input:        "short\n",
            expectError:  true,
            errorMessage: "insufficient entropy: got 5 chars, need at least",
        },
        {
            name:         "good entropy but rejected",
            args:         []string{fmt.Sprintf("--%s=1", flagUserEntropy)},
            input:        strings.Repeat("quantum", 20) + "\nn\n",
            expectError:  false,
        },
        {
            name:         "good entropy and accepted",
            args:         []string{fmt.Sprintf("--%s=1", flagUserEntropy)},
            input:        strings.Repeat("quantum", 20) + "\ny\n",
            expectError:  false,
        },
        {
            name: "quantum-safe disabled",
            args: []string{
                fmt.Sprintf("--%s=false", flagQuantumSafe),
                fmt.Sprintf("--%s=256", flagEntropySize),
            },
            expectError: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            cmd := MnemonicKeyCommand()
            mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

            if tc.input != "" {
                mockIn.Reset(tc.input)
            }

            cmd.SetArgs(tc.args)
            err := cmd.Execute()

            if tc.expectError {
                require.Error(t, err)
                if tc.errorMessage != "" {
                    require.Contains(t, err.Error(), tc.errorMessage)
                }
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestEntropyEnhancement(t *testing.T) {
    testCases := []struct {
        name     string
        entropy  []byte
        size     int
        validate func(t *testing.T, original, enhanced []byte)
    }{
        {
            name:    "256-bit entropy enhancement",
            entropy: make([]byte, 32),
            size:    256,
            validate: func(t *testing.T, original, enhanced []byte) {
                require.NotEqual(t, original, enhanced)
                require.Equal(t, len(original), len(enhanced))
            },
        },
        {
            name:    "512-bit entropy enhancement",
            entropy: make([]byte, 64),
            size:    512,
            validate: func(t *testing.T, original, enhanced []byte) {
                require.NotEqual(t, original, enhanced)
                require.Equal(t, len(original), len(enhanced))
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Generate some test entropy
            originalEntropy, err := getSystemEntropy(tc.size)
            require.NoError(t, err)
            require.Equal(t, len(tc.entropy), len(originalEntropy))

            // Enhance the entropy
            enhancedEntropy := enhanceEntropyForQuantumSafety(originalEntropy)

            // Run validation
            tc.validate(t, originalEntropy, enhancedEntropy)
        })
    }
}

func TestUserEntropyValidation(t *testing.T) {
    cmd := MnemonicKeyCommand()
    mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

    testCases := []struct {
        name         string
        input        string
        entropySize  int
        expectError  bool
        errorMessage string
    }{
        {
            name:         "valid entropy input",
            input:        strings.Repeat("quantum", 20) + "\ny\n",
            entropySize:  256,
            expectError:  false,
        },
        {
            name:         "short entropy input",
            input:        "short\n",
            entropySize:  256,
            expectError:  true,
            errorMessage: "insufficient entropy",
        },
        {
            name:         "rejected confirmation",
            input:        strings.Repeat("quantum", 20) + "\nn\n",
            entropySize:  256,
            expectError:  true,
            errorMessage: "entropy input cancelled",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            mockIn.Reset(tc.input)
            entropy, err := getUserEntropy(cmd, tc.entropySize)

            if tc.expectError {
                require.Error(t, err)
                if tc.errorMessage != "" {
                    require.Contains(t, err.Error(), tc.errorMessage)
                }
            } else {
                require.NoError(t, err)
                require.NotNil(t, entropy)
                require.Equal(t, tc.entropySize/8, len(entropy))
            }
        })
    }
}
