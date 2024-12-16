package keys

import (
    "context"
    "fmt"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/baron-chain/cosmos-sdk/client"
    "github.com/baron-chain/cosmos-sdk/client/flags"
    clienttestutil "github.com/baron-chain/cosmos-sdk/client/testutil"
    "github.com/baron-chain/cosmos-sdk/crypto/keyring"
    "github.com/baron-chain/cosmos-sdk/testutil"
    clitestutil "github.com/baron-chain/cosmos-sdk/testutil/cli"
    "github.com/baron-chain/cosmos-sdk/testutil/testdata"
    sdk "github.com/baron-chain/cosmos-sdk/types"
)

func cleanupKeys(t *testing.T, kr keyring.Keyring, keys ...string) func() {
    return func() {
        for _, k := range keys {
            err := kr.Delete(k)
            if err != nil {
                t.Logf("failed to delete key %s: %v", k, err)
            }
        }
    }
}

func TestListCmd(t *testing.T) {
    cmd := ListKeysCmd()
    cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

    testCases := []struct {
        name        string
        args        []string
        expectKeys  bool
        expectAlgo  bool
        wantErr     bool
    }{
        {
            name: "list empty keyring",
            args: []string{
                fmt.Sprintf("--%s=false", flagListNames),
            },
            expectKeys: false,
        },
        {
            name: "list with quantum keys",
            args: []string{
                fmt.Sprintf("--%s=false", flagListNames),
            },
            expectKeys: true,
        },
        {
            name: "list names only",
            args: []string{
                fmt.Sprintf("--%s=true", flagListNames),
            },
            expectKeys: true,
        },
        {
            name: "list with algorithms",
            args: []string{
                fmt.Sprintf("--%s=false", flagListNames),
                fmt.Sprintf("--%s=true", flagShowAlgo),
            },
            expectKeys: true,
            expectAlgo: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            kbHome := t.TempDir()
            mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
            cdc := clienttestutil.MakeTestCodec(t)

            kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
            assert.NoError(t, err)

            // Add test quantum-safe keys if needed
            if tc.expectKeys {
                // Add Kyber test key
                _, err = kb.NewAccount(
                    "kyber-key",
                    testdata.TestMnemonic,
                    "",
                    "", // No HD path for quantum keys
                    "kyber",
                )
                assert.NoError(t, err)

                // Add Dilithium test key
                _, err = kb.NewAccount(
                    "dilithium-key",
                    testdata.TestMnemonic,
                    "",
                    "",
                    "dilithium",
                )
                assert.NoError(t, err)

                t.Cleanup(cleanupKeys(t, kb, "kyber-key", "dilithium-key"))
            }

            clientCtx := client.Context{}.
                WithKeyring(kb).
                WithKeyringDir(kbHome)
            ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

            cmd.SetArgs(tc.args)
            err = cmd.ExecuteContext(ctx)

            if tc.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
        })
    }
}

func TestListKeyTypesCmd(t *testing.T) {
    cmd := ListKeyTypesCmd()
    kbHome := t.TempDir()
    mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
    cdc := clienttestutil.MakeTestCodec(t)

    kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
    assert.NoError(t, err)

    clientCtx := client.Context{}.
        WithKeyringDir(kbHome).
        WithKeyring(kb)

    out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{})
    assert.NoError(t, err)

    // Verify quantum-safe algorithms are listed
    output := out.String()
    assert.Contains(t, output, "kyber")
    assert.Contains(t, output, "dilithium")
}

func TestKeyringOutput(t *testing.T) {
    cmd := ListKeysCmd()
    kbHome := t.TempDir()
    mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
    cdc := clienttestutil.MakeTestCodec(t)

    kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
    assert.NoError(t, err)

    // Create test quantum keys
    _, err = kb.NewAccount("test-kyber", testdata.TestMnemonic, "", "", "kyber")
    assert.NoError(t, err)
    t.Cleanup(cleanupKeys(t, kb, "test-kyber"))

    clientCtx := client.Context{}.
        WithKeyring(kb).
        WithOutput(cmd.OutOrStdout())

    // Test JSON output
    cmd.SetArgs([]string{"--output=json"})
    err = cmd.ExecuteContext(context.WithValue(context.Background(), client.ClientContextKey, &clientCtx))
    assert.NoError(t, err)
}
