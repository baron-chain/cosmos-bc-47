package crypto_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	tmcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/xsalsa20symmetric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bcrypt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	_ "github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/types"
)

const (
	testPassphrase = "passphrase"
	testKeyType    = "secp256k1"
	testKeyName    = "Bob"
)

func TestPrivKeyArmorOperations(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	armored := crypto.EncryptArmorPrivKey(priv, testPassphrase, "")

	t.Run("wrong passphrase", func(t *testing.T) {
		_, _, err := crypto.UnarmorDecryptPrivKey(armored, "wrongpassphrase")
		require.Error(t, err)
	})

	t.Run("correct passphrase", func(t *testing.T) {
		decrypted, algo, err := crypto.UnarmorDecryptPrivKey(armored, testPassphrase)
		require.NoError(t, err)
		require.Equal(t, string(hd.Secp256k1Type), algo)
		require.True(t, priv.Equals(decrypted))
	})

	t.Run("empty armor string", func(t *testing.T) {
		decrypted, algo, err := crypto.UnarmorDecryptPrivKey("", testPassphrase)
		require.True(t, errors.Is(err, io.EOF))
		require.Nil(t, decrypted)
		require.Empty(t, algo)
	})

	t.Run("wrong key type", func(t *testing.T) {
		wrongArmored := crypto.ArmorPubKeyBytes(priv.PubKey().Bytes(), "")
		_, _, err := crypto.UnarmorDecryptPrivKey(wrongArmored, testPassphrase)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unrecognized armor type")
	})

	t.Run("wrong kdf header", func(t *testing.T) {
		saltBytes, encBytes := encryptPrivKey(t, priv, testPassphrase)
		headerWrongKdf := map[string]string{
			"kdf":  "wrong",
			"salt": fmt.Sprintf("%X", saltBytes),
			"type": testKeyType,
		}
		wrongArmored := crypto.EncodeArmor("TENDERMINT PRIVATE KEY", headerWrongKdf, encBytes)
		_, _, err := crypto.UnarmorDecryptPrivKey(wrongArmored, testPassphrase)
		require.EqualError(t, err, "unrecognized KDF type: wrong")
	})
}

func TestPubKeyArmorOperations(t *testing.T) {
	var cdc codec.Codec
	require.NoError(t, depinject.Inject(configurator.NewAppConfig(), &cdc))
	cstore := keyring.NewInMemory(cdc)

	k, _, err := cstore.NewMnemonic(testKeyName, keyring.English, types.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)

	key, err := k.GetPubKey()
	require.NoError(t, err)
	keyBytes := legacy.Cdc.Amino.MustMarshalBinaryBare(key)

	t.Run("standard armor", func(t *testing.T) {
		armored := crypto.ArmorPubKeyBytes(keyBytes, "")
		pubBytes, algo, err := crypto.UnarmorPubKeyBytes(armored)
		require.NoError(t, err)
		pub, err := legacy.PubKeyFromBytes(pubBytes)
		require.NoError(t, err)
		require.Equal(t, string(hd.Secp256k1Type), algo)
		require.True(t, pub.Equals(key))
	})

	t.Run("custom algorithm", func(t *testing.T) {
		armored := crypto.ArmorPubKeyBytes(keyBytes, "unknown")
		pubBytes, algo, err := crypto.UnarmorPubKeyBytes(armored)
		require.NoError(t, err)
		pub, err := legacy.PubKeyFromBytes(pubBytes)
		require.NoError(t, err)
		require.Equal(t, "unknown", algo)
		require.True(t, pub.Equals(key))
	})

	t.Run("private key armor", func(t *testing.T) {
		armored, err := cstore.ExportPrivKeyArmor(testKeyName, testPassphrase)
		require.NoError(t, err)
		_, _, err = crypto.UnarmorPubKeyBytes(armored)
		require.EqualError(t, err, `couldn't unarmor bytes: unrecognized armor type "TENDERMINT PRIVATE KEY", expected: "TENDERMINT PUBLIC KEY"`)
	})

	t.Run("header versions", func(t *testing.T) {
		testCases := []struct {
			name    string
			header  map[string]string
			expAlgo string
			expErr  string
		}{
			{
				name:    "version 0.0.0",
				header:  map[string]string{"version": "0.0.0", "type": "unknown"},
				expAlgo: testKeyType,
			},
			{
				name:   "missing version",
				header: map[string]string{"type": "unknown"},
				expErr: "header's version field is empty",
			},
			{
				name:   "unknown version",
				header: map[string]string{"type": "unknown", "version": "unknown"},
				expErr: "unrecognized version: unknown",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				armored := crypto.EncodeArmor("TENDERMINT PUBLIC KEY", tc.header, keyBytes)
				bz, algo, err := crypto.UnarmorPubKeyBytes(armored)
				if tc.expErr != "" {
					require.EqualError(t, err, tc.expErr)
					require.Nil(t, bz)
					require.Empty(t, algo)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.expAlgo, algo)
			})
		}
	})
}

func TestInfoBytesArmor(t *testing.T) {
	testData := []byte("test")
	armored := crypto.ArmorInfoBytes(testData)
	unarmored, err := crypto.UnarmorInfoBytes(armored)
	require.NoError(t, err)
	require.True(t, bytes.Equal(testData, unarmored))

	t.Run("errors", func(t *testing.T) {
		testCases := []struct {
			name    string
			armored string
			expErr  string
		}{
			{
				name:    "empty string",
				armored: "",
				expErr:  "EOF",
			},
			{
				name: "wrong version",
				armored: crypto.EncodeArmor("TENDERMINT KEY INFO",
					map[string]string{"type": "Info", "version": "0.0.1"},
					[]byte("plain-text")),
				expErr: "unrecognized version: 0.0.1",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				unarmored, err := crypto.UnarmorInfoBytes(tc.armored)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErr)
				require.Nil(t, unarmored)
			})
		}
	})
}

func TestBasicArmor(t *testing.T) {
	blockType := "MINT TEST"
	data := []byte("somedata")
	armored := crypto.EncodeArmor(blockType, nil, data)

	blockType2, _, data2, err := crypto.DecodeArmor(armored)
	require.NoError(t, err)
	assert.Equal(t, blockType, blockType2)
	assert.Equal(t, data, data2)
}

func BenchmarkBcryptGenerateFromPassword(b *testing.B) {
	passphrase := []byte(testPassphrase)
	for param := 9; param < 16; param++ {
		b.Run(fmt.Sprintf("security-param-%d", param), func(b *testing.B) {
			b.ReportAllocs()
			saltBytes := tmcrypto.CRandBytes(16)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := bcrypt.GenerateFromPassword(saltBytes, passphrase, param)
				require.NoError(b, err)
			}
		})
	}
}

// Helper function to encrypt private key for testing
func encryptPrivKey(t *testing.T, privKey cryptotypes.PrivKey, passphrase string) (saltBytes, encBytes []byte) {
	saltBytes = tmcrypto.CRandBytes(16)
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), crypto.BcryptSecurityParameter)
	require.NoError(t, err)
	key = tmcrypto.Sha256(key)
	privKeyBytes := legacy.Cdc.Amino.MustMarshalBinaryBare(privKey)
	return saltBytes, xsalsa20symmetric.EncryptSymmetric(privKeyBytes, key)
}
