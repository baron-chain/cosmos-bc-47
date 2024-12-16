package keys

import (
    "context"
    "fmt"
    "testing"

    design99keyring "github.com/99designs/keyring"
    "github.com/stretchr/testify/suite"
    "github.com/baron-chain/cosmos-sdk/client"
    "github.com/baron-chain/cosmos-sdk/client/flags"
    clienttestutil "github.com/baron-chain/cosmos-sdk/client/testutil"
    "github.com/baron-chain/cosmos-sdk/codec"
    "github.com/baron-chain/cosmos-sdk/crypto/keyring"
    "github.com/baron-chain/cosmos-sdk/crypto/keys/kyber"
    "github.com/baron-chain/cosmos-sdk/crypto/keys/dilithium"
    cryptotypes "github.com/baron-chain/cosmos-sdk/crypto/types"
    "github.com/baron-chain/cosmos-sdk/testutil"
)

type setter interface {
    SetItem(item design99keyring.Item) error
}

type MigrateTestSuite struct {
    suite.Suite

    dir          string
    appName      string
    cdc          codec.Codec
    kyberKey     cryptotypes.PrivKey
    dilithiumKey cryptotypes.PrivKey
}

func (s *MigrateTestSuite) SetupSuite() {
    s.dir = s.T().TempDir()
    s.cdc = clienttestutil.MakeTestCodec(s.T())
    s.appName = "baron-chain"
    s.kyberKey = kyber.GenPrivKey()
    s.dilithiumKey = dilithium.GenPrivKey()
}

func (s *MigrateTestSuite) TestMigrateToQuantumSafe() {
    cmd := MigrateCommand()
    cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

    mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
    kb, err := keyring.New(s.appName, keyring.BackendTest, s.dir, mockIn, s.cdc)
    s.Require().NoError(err)

    // Test migration to Kyber
    s.Run("Migrate to Kyber", func() {
        record, err := keyring.NewLocalRecord("kyber-test", s.kyberKey, s.kyberKey.PubKey())
        s.Require().NoError(err)
        
        serialized, err := s.cdc.Marshal(record)
        s.Require().NoError(err)

        item := design99keyring.Item{
            Key:         "kyber-test",
            Data:        serialized,
            Description: "Kyber quantum-safe key",
        }

        setter, ok := kb.(setter)
        s.Require().True(ok)
        s.Require().NoError(setter.SetItem(item))

        clientCtx := client.Context{}.WithKeyring(kb)
        ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

        cmd.SetArgs([]string{
            fmt.Sprintf("--%s=kyber", flagQuantumKey),
            fmt.Sprintf("--%s=false", flagDryRun),
        })
        s.Require().NoError(cmd.ExecuteContext(ctx))
    })

    // Test migration to Dilithium
    s.Run("Migrate to Dilithium", func() {
        record, err := keyring.NewLocalRecord("dilithium-test", s.dilithiumKey, s.dilithiumKey.PubKey())
        s.Require().NoError(err)
        
        serialized, err := s.cdc.Marshal(record)
        s.Require().NoError(err)

        item := design99keyring.Item{
            Key:         "dilithium-test",
            Data:        serialized,
            Description: "Dilithium quantum-safe key",
        }

        setter, ok := kb.(setter)
        s.Require().True(ok)
        s.Require().NoError(setter.SetItem(item))

        clientCtx := client.Context{}.WithKeyring(kb)
        ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

        cmd.SetArgs([]string{
            fmt.Sprintf("--%s=dilithium", flagQuantumKey),
            fmt.Sprintf("--%s=false", flagDryRun),
        })
        s.Require().NoError(cmd.ExecuteContext(ctx))
    })
}

func (s *MigrateTestSuite) TestDryRunMigration() {
    cmd := MigrateCommand()
    mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
    kb, err := keyring.New(s.appName, keyring.BackendTest, s.dir, mockIn, s.cdc)
    s.Require().NoError(err)

    // Setup test key
    record, err := keyring.NewLocalRecord("test-key", s.kyberKey, s.kyberKey.PubKey())
    s.Require().NoError(err)
    
    serialized, err := s.cdc.Marshal(record)
    s.Require().NoError(err)

    item := design99keyring.Item{
        Key:         "test-key",
        Data:        serialized,
        Description: "Test key for dry run",
    }

    setter, ok := kb.(setter)
    s.Require().True(ok)
    s.Require().NoError(setter.SetItem(item))

    clientCtx := client.Context{}.WithKeyring(kb)
    ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

    cmd.SetArgs([]string{
        fmt.Sprintf("--%s=true", flagDryRun),
        fmt.Sprintf("--%s=kyber", flagQuantumKey),
    })
    s.Require().NoError(cmd.ExecuteContext(ctx))
}

func (s *MigrateTestSuite) TestListMigratedKeys() {
    cmd := ListKeysCmd()
    cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

    mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
    kb, err := keyring.New(s.appName, keyring.BackendTest, s.dir, mockIn, s.cdc)
    s.Require().NoError(err)

    // Add both types of quantum-safe keys
    for _, tc := range []struct {
        name string
        key  cryptotypes.PrivKey
    }{
        {"kyber-key", s.kyberKey},
        {"dilithium-key", s.dilithiumKey},
    } {
        record, err := keyring.NewLocalRecord(tc.name, tc.key, tc.key.PubKey())
        s.Require().NoError(err)
        
        serialized, err := s.cdc.Marshal(record)
        s.Require().NoError(err)

        item := design99keyring.Item{
            Key:         tc.name,
            Data:        serialized,
            Description: "Quantum-safe key",
        }

        setter, ok := kb.(setter)
        s.Require().True(ok)
        s.Require().NoError(setter.SetItem(item))
    }

    clientCtx := client.Context{}.WithKeyring(kb)
    ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

    cmd.SetArgs([]string{
        fmt.Sprintf("--%s=%s", flags.FlagHome, s.dir),
        fmt.Sprintf("--%s=true", flagShowAlgo),
    })
    s.Require().NoError(cmd.ExecuteContext(ctx))
}

func TestMigrateTestSuite(t *testing.T) {
    suite.Run(t, new(MigrateTestSuite))
}
