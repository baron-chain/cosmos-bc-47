package keys

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/baron-chain/cosmos-sdk/client"
    "github.com/baron-chain/cosmos-sdk/client/flags"
    clienttestutil "github.com/baron-chain/cosmos-sdk/client/testutil"
    "github.com/baron-chain/cosmos-sdk/crypto/keyring"
    "github.com/baron-chain/cosmos-sdk/testutil"
    sdk "github.com/baron-chain/cosmos-sdk/types"
)

func TestImportCmd(t *testing.T) {
    cdc := clienttestutil.MakeTestCodec(t)
    
    testCases := []struct {
        name           string
        keyringBackend string
        userInput      string
        keyAlgorithm   string
        expectError    bool
    }{
        {
            name:           "kyber key import success",
            keyringBackend: keyring.BackendTest,
            userInput:      "123456789\n",
            keyAlgorithm:   "kyber",
        },
        {
            name:           "dilithium key import success",
            keyringBackend: keyring.BackendTest,
            userInput:      "123456789\n",
            keyAlgorithm:   "dilithium",
        },
        {
            name:           "invalid passphrase",
            keyringBackend: keyring.BackendTest,
            userInput:      "wrong\n",
            keyAlgorithm:   "kyber",
            expectError:    true,
        },
        {
            name:           "file backend with kyber",
            keyringBackend: keyring.BackendFile,
            userInput:      "123456789\n12345678\n12345678\n",
            keyAlgorithm:   "kyber",
        },
    }

    // Quantum-safe test key (Kyber format)
    quantumSafeKey := `-----BEGIN BARON CHAIN QUANTUM KEY-----
algorithm: kyber
salt: Q790BB721D1C094260EA84F5E5B72289
kdf: argon2id

HbP+c6JmeJy9JXe2rbbF1QtCX1gLqGcDQPBXiCtFvP7/8wTZtVOPj8vREzhZ9ElO
3P7YnrzPQThG0Q+ZnRSbl9MAS8uFAM4mqm5r/Ys=
=f3l4
-----END BARON CHAIN QUANTUM KEY-----
`

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            cmd := ImportKeyCommand()
            cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
            mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

            kbHome := t.TempDir()
            kb, err := keyring.New(sdk.KeyringServiceName(), tc.keyringBackend, kbHome, nil, cdc)
            require.NoError(t, err)

            clientCtx := client.Context{}.
                WithKeyringDir(kbHome).
                WithKeyring(kb).
                WithInput(mockIn).
                WithCodec(cdc)
            ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

            t.Cleanup(func() {
                cleanupKeys(t, kb, "keyname1")
                os.RemoveAll(kbHome)
            })

            keyfile := filepath.Join(kbHome, "key.asc")
            require.NoError(t, os.WriteFile(keyfile, []byte(quantumSafeKey), 0o600))

            mockIn.Reset(tc.userInput)
            cmd.SetArgs([]string{
                "keyname1",
                keyfile,
                fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, tc.keyringBackend),
                fmt.Sprintf("--%s=%s", flagKeyAlgorithm, tc.keyAlgorithm),
            })

            err = cmd.ExecuteContext(ctx)
            if tc.expectError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestImportHexCmd(t *testing.T) {
    cdc := clienttestutil.MakeTestCodec(t)
    
    testCases := []struct {
        name           string
        keyringBackend string
        hexKey         string
        keyAlgorithm   string
        expectError    bool
    }{
        {
            name:           "kyber hex key import",
            keyringBackend: keyring.BackendTest,
            hexKey:        "0x7b3e57952e835ed30eea86a2993ac2a61c03e74f2085b3635bd94aa4d7ae0cfdf",
            keyAlgorithm:   "kyber",
        },
        {
            name:           "dilithium hex key import",
            keyringBackend: keyring.BackendTest,
            hexKey:        "0x8c4e57952e835ed30eea86a2993ac2a61c03e74f2085b3635bd94aa4d7ae0cfdf",
            keyAlgorithm:   "dilithium",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            cmd := ImportHexCommand()
            cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

            kbHome := t.TempDir()
            kb, err := keyring.New(sdk.KeyringServiceName(), tc.keyringBackend, kbHome, nil, cdc)
            require.NoError(t, err)

            clientCtx := client.Context{}.
                WithKeyringDir(kbHome).
                WithKeyring(kb).
                WithCodec(cdc)
            ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

            t.Cleanup(func() {
                cleanupKeys(t, kb, "keyname1")
                os.RemoveAll(kbHome)
            })

            cmd.SetArgs([]string{
                "keyname1",
                tc.hexKey,
                fmt.Sprintf("--%s=%s", flagKeyAlgorithm, tc.keyAlgorithm),
                fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, tc.keyringBackend),
            })

            err = cmd.ExecuteContext(ctx)
            if tc.expectError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
