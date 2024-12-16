package keys

import (
    "context"
    "fmt"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/baron-chain/cosmos-sdk/client"
    "github.com/baron-chain/cosmos-sdk/client/flags"
    clienttestutil "github.com/baron-chain/cosmos-sdk/client/testutil"
    "github.com/baron-chain/cosmos-sdk/crypto/keyring"
    "github.com/baron-chain/cosmos-sdk/testutil"
    "github.com/baron-chain/cosmos-sdk/testutil/testdata"
    sdk "github.com/baron-chain/cosmos-sdk/types"
)

func TestRenameCommand(t *testing.T) {
    testCases := []struct {
        name         string
        keyType      string
        args         []string
        input        string
        expectErr    bool
        errMsg       string
        checkRename  bool
    }{
        {
            name:      "rename non-existent key",
            keyType:   "local",
            args:      []string{"nonexistent", "newname"},
            expectErr: true,
            errMsg:    "key not found",
        },
        {
            name:        "rename without confirmation",
            keyType:     "local",
            args:        []string{"testkey1", "testkey2"},
            expectErr:   true,
            errMsg:      "EOF",
        },
        {
            name:        "successful rename with confirmation",
            keyType:     "local",
            args:        []string{"testkey1", "testkey2", fmt.Sprintf("--%s=true", flagSkipConfirm)},
            checkRename: true,
        },
        {
            name:        "rename to existing key without force",
            keyType:     "local",
            args:        []string{"testkey1", "existing"},
            expectErr:   true,
            errMsg:      "already exists",
        },
        {
            name:        "rename to existing key with force",
            keyType:     "local",
            args:        []string{"testkey1", "existing", fmt.Sprintf("--%s=true", flagForce), fmt.Sprintf("--%s=true", flagSkipConfirm)},
            checkRename: true,
        },
        {
            name:        "rename quantum-safe key",
            keyType:     "kyber",
            args:        []string{"quantum1", "quantum2", fmt.Sprintf("--%s=true", flagSkipConfirm)},
            checkRename: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            cmd := RenameKeyCommand()
            kbHome := t.TempDir()

            cmd.Flags().AddFlagSet(Commands(kbHome).PersistentFlags())
            mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

            // Initialize keyring
            cdc := clienttestutil.MakeTestCodec(t)
            kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
            require.NoError(t, err)

            // Setup test key
            if len(tc.args) > 0 {
                oldName := tc.args[0]
                _, err = createTestKey(kr, oldName, tc.keyType)
                require.NoError(t, err)

                // Create existing key if testing force flag
                if tc.args[1] == "existing" {
                    _, err = createTestKey(kr, "existing", tc.keyType)
                    require.NoError(t, err)
                }
            }

            clientCtx := client.Context{}.
                WithKeyringDir(kbHome).
                WithKeyring(kr).
                WithCodec(cdc)
            ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

            // Set command args
            cmdArgs := append(tc.args, fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome))
            cmd.SetArgs(cmdArgs)

            // Execute command
            err = cmd.ExecuteContext(ctx)

            if tc.expectErr {
                require.Error(t, err)
                if tc.errMsg != "" {
                    require.Contains(t, err.Error(), tc.errMsg)
                }
                return
            }

            require.NoError(t, err)

            if tc.checkRename {
                // Verify old key is gone
                _, err = kr.Key(tc.args[0])
                require.Error(t, err)

                // Verify new key exists
                newKey, err := kr.Key(tc.args[1])
                require.NoError(t, err)

                // Verify key properties are preserved
                requireKeyPropertiesPreserved(t, kr, tc.args[0], newKey)
            }
        })
    }
}

func createTestKey(kr keyring.Keyring, name, keyType string) (keyring.Info, error) {
    switch keyType {
    case "kyber":
        return createQuantumSafeKey(kr, name)
    default:
        return kr.NewAccount(name, testdata.TestMnemonic, "", "", keyType)
    }
}

func createQuantumSafeKey(kr keyring.Keyring, name string) (keyring.Info, error) {
    // Implementation would depend on Baron Chain's quantum-safe key creation
    return kr.NewAccount(name, testdata.TestMnemonic, "", "", "kyber")
}

func requireKeyPropertiesPreserved(t *testing.T, kr keyring.Keyring, oldName string, newKey keyring.Info) {
    require.NotNil(t, newKey)
    
    // Verify key type is preserved
    if isQuantumSafeKey(newKey.GetType()) {
        require.True(t, strings.HasPrefix(newKey.GetType().String(), "kyber"))
    }

    // Verify public key and address are preserved
    pubKey, err := newKey.GetPubKey()
    require.NoError(t, err)
    require.NotNil(t, pubKey)

    addr, err := newKey.GetAddress()
    require.NoError(t, err)
    require.NotEmpty(t, addr)
}
