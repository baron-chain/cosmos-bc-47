package keys_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/baron-chain/cosmos-sdk/client/keys"
	"github.com/baron-chain/cometbft-bc/crypto/kyber"
)

func TestKeyConstructors(t *testing.T) {
	t.Run("AddNewKey", func(t *testing.T) {
		newKey := keys.NewAddNewKey("validator1", "quantum-safe-pwd", "test-mnemonic", 0, 0)
		require.Equal(t, keys.AddNewKey{
			Name:     "validator1",
			Password: "quantum-safe-pwd",
			Mnemonic: "test-mnemonic",
			Account:  0,
			Index:    0,
		}, newKey)
	})

	t.Run("RecoverKey", func(t *testing.T) {
		recoveredKey := keys.NewRecoverKey("quantum-safe-pwd", "test-mnemonic", 0, 0)
		require.Equal(t, keys.RecoverKey{
			Password: "quantum-safe-pwd",
			Mnemonic: "test-mnemonic",
			Account:  0,
			Index:    0,
		}, recoveredKey)
	})

	t.Run("UpdateKey", func(t *testing.T) {
		updatedKey := keys.NewUpdateKeyReq("old-pwd", "new-quantum-pwd")
		require.Equal(t, keys.UpdateKeyReq{
			OldPassword: "old-pwd",
			NewPassword: "new-quantum-pwd",
		}, updatedKey)
	})

	t.Run("DeleteKey", func(t *testing.T) {
		deletedKey := keys.NewDeleteKeyReq("quantum-safe-pwd")
		require.Equal(t, keys.DeleteKeyReq{
			Password: "quantum-safe-pwd",
		}, deletedKey)
	})
}

func TestQuantumKeyGeneration(t *testing.T) {
	t.Run("KyberKeyPair", func(t *testing.T) {
		privateKey, publicKey, err := kyber.GenerateKey()
		require.NoError(t, err)
		require.NotNil(t, privateKey)
		require.NotNil(t, publicKey)
	})
}
