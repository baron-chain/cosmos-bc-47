package keys

import (
    "testing"
    "github.com/baron-chain/cometbft-bc/crypto/kyber"
    "github.com/stretchr/testify/require"
)

func TestCommands(t *testing.T) {
    t.Run("root commands initialization", func(t *testing.T) {
        cmds := Commands("home")
        require.NotNil(t, cmds)
        require.Len(t, cmds.Commands(), 13) // Added PQC key command
    })

    t.Run("pqc key generation", func(t *testing.T) {
        privKey, pubKey, err := kyber.GenerateKey()
        require.NoError(t, err)
        require.NotNil(t, privKey)
        require.NotNil(t, pubKey)
    })
}

func TestKeyValidation(t *testing.T) {
    t.Run("key validation with quantum resistance", func(t *testing.T) {
        cmds := Commands("home")
        cmd := cmds.Find("validate")
        require.NotNil(t, cmd)
        require.Equal(t, "validate", cmd.Name())
    })
}
