package keys

import (
    "context"
    "fmt"
    "testing"

    "github.com/stretchr/testify/require"

    "github.com/baron-chain/cosmos-sdk/client"
    "github.com/baron-chain/cosmos-sdk/client/flags"
    clienttestutil "github.com/baron-chain/cosmos-sdk/client/testutil"
    "github.com/baron-chain/cosmos-sdk/crypto/hd"
    "github.com/baron-chain/cosmos-sdk/crypto/keyring"
    "github.com/baron-chain/cosmos-sdk/crypto/keys/kyber"
    "github.com/baron-chain/cosmos-sdk/crypto/keys/multisig"
    cryptotypes "github.com/baron-chain/cosmos-sdk/crypto/types"
    "github.com/baron-chain/cosmos-sdk/testutil"
    "github.com/baron-chain/cosmos-sdk/testutil/testdata"
    sdk "github.com/baron-chain/cosmos-sdk/types"
)

func TestMultiSigKeyProperties(t *testing.T) {
    tmpKey1 := kyber.GenPrivKeyFromSecret([]byte("quantumSecret"))
    pk := multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{tmpKey1.PubKey()})
    
    k, err := keyring.NewMultiRecord("quantumMultisig", pk)
    require.NoError(t, err)
    require.Equal(t, "quantumMultisig", k.Name)
    require.Equal(t, keyring.TypeMulti, k.GetType())

    pub, err := k.GetPubKey()
    require.NoError(t, err)
    require.NotEmpty(t, pub.Address().String())

    addr, err := k.GetAddress()
    require.NoError(t, err)
    require.NotEmpty(t, sdk.MustBech32ifyAddressBytes("baron", addr))
}

func TestShowKeysCmd(t *testing.T) {
    cmd := ShowKeysCmd()
    require.NotNil(t, cmd)
    require.Equal(t, "false", cmd.Flag(FlagAddress).DefValue)
    require.Equal(t, "false", cmd.Flag(FlagPublicKey).DefValue)
    require.Equal(t, "true", cmd.Flag(FlagQuantumSafe).DefValue)
}

func TestRunShowCmd(t *testing.T) {
    cmd := ShowKeysCmd()
    cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
    mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

    kbHome := t.TempDir()
    cdc := clienttestutil.MakeTestCodec(t)
    kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
    require.NoError(t, err)

    clientCtx := client.Context{}.
        WithKeyringDir(kbHome).
        WithCodec(cdc)
    ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

    testCases := []struct {
        name      string
        args      []string
        expectErr string
    }{
        {
            "invalid key",
            []string{"invalid"},
            "invalid is not a valid name or address: decoding bech32 failed",
        },
        {
            "multiple invalid keys",
            []string{"invalid1", "invalid2"},
            "invalid1 is not a valid name or address: decoding bech32 failed",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            cmd.SetArgs(tc.args)
            err := cmd.ExecuteContext(ctx)
            require.Contains(t, err.Error(), tc.expectErr)
        })
    }

    // Test quantum-safe key operations
    t.Run("quantum safe key operations", func(t *testing.T) {
        quantumKey := "quantumTestKey"
        path := hd.NewFundraiserParams(1, sdk.CoinType, 0).String()
        
        _, err = kb.NewAccount(quantumKey, testdata.TestMnemonic, "", path, kyber.Kyber)
        require.NoError(t, err)

        cmd.SetArgs([]string{
            quantumKey,
            fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
            fmt.Sprintf("--%s=%s", FlagBechPrefix, sdk.PrefixAccount),
            fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
            fmt.Sprintf("--%s=true", FlagQuantumSafe),
        })
        require.NoError(t, cmd.ExecuteContext(ctx))
    })
}

func TestValidateMultisigThreshold(t *testing.T) {
    testCases := []struct {
        name      string
        k         int
        nKeys     int
        expectErr bool
    }{
        {"invalid zero values", 0, 0, true},
        {"invalid threshold", 1, 0, true},
        {"valid single key", 1, 1, false},
        {"valid multi key", 2, 3, false},
        {"invalid multi key", 3, 2, true},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := validateMultisigThreshold(tc.k, tc.nKeys)
            if tc.expectErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestGetBechKeyOut(t *testing.T) {
    testCases := []struct {
        name       string
        bechPrefix string
        expectErr  bool
    }{
        {"empty prefix", "", true},
        {"invalid prefix", "invalid", true},
        {"account prefix", sdk.PrefixAccount, false},
        {"validator prefix", sdk.PrefixValidator, false},
        {"consensus prefix", sdk.PrefixConsensus, false},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := getBechKeyOut(tc.bechPrefix)
            if tc.expectErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                require.NotNil(t, got)
            }
        })
    }
}
