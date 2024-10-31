package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/cometbft/cometbft/crypto"
	"golang.org/x/crypto/openpgp/armor" //nolint:staticcheck

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bcrypt"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/xsalsa20symmetric"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	blockTypePrivKey = "TENDERMINT PRIVATE KEY"
	blockTypeKeyInfo = "TENDERMINT KEY INFO"
	blockTypePubKey  = "TENDERMINT PUBLIC KEY"
	defaultAlgo      = "secp256k1"
	headerVersion    = "version"
	headerType       = "type"
	headerKDF        = "kdf"
	headerSalt       = "salt"
	bcryptKDF        = "bcrypt"
	version0         = "0.0.0"
	version1         = "0.0.1"
)

// BcryptSecurityParameter defines the security level for bcrypt key generation
var BcryptSecurityParameter = 12

// ArmorInfoBytes encrypts info bytes with armor encoding
func ArmorInfoBytes(bz []byte) string {
	header := map[string]string{
		headerType:    "Info",
		headerVersion: version0,
	}
	return EncodeArmor(blockTypeKeyInfo, header, bz)
}

// ArmorPubKeyBytes encrypts public key bytes with armor encoding
func ArmorPubKeyBytes(bz []byte, algo string) string {
	header := map[string]string{
		headerVersion: version1,
	}
	if algo != "" {
		header[headerType] = algo
	}
	return EncodeArmor(blockTypePubKey, header, bz)
}

// UnarmorInfoBytes decrypts armored info bytes
func UnarmorInfoBytes(armorStr string) ([]byte, error) {
	bz, header, err := unarmorBytes(armorStr, blockTypeKeyInfo)
	if err != nil {
		return nil, err
	}

	if header[headerVersion] != version0 {
		return nil, fmt.Errorf("unrecognized version: %v", header[headerVersion])
	}

	return bz, nil
}

// UnarmorPubKeyBytes decrypts armored public key bytes and returns the key bytes, algorithm and any error
func UnarmorPubKeyBytes(armorStr string) ([]byte, string, error) {
	bz, header, err := unarmorBytes(armorStr, blockTypePubKey)
	if err != nil {
		return nil, "", fmt.Errorf("couldn't unarmor bytes: %v", err)
	}

	switch header[headerVersion] {
	case version0:
		return bz, defaultAlgo, nil
	case version1:
		algo := header[headerType]
		if algo == "" {
			algo = defaultAlgo
		}
		return bz, algo, nil
	case "":
		return nil, "", fmt.Errorf("header's version field is empty")
	default:
		return nil, "", fmt.Errorf("unrecognized version: %v", header[headerVersion])
	}
}

// EncryptArmorPrivKey encrypts and armors a private key
func EncryptArmorPrivKey(privKey cryptotypes.PrivKey, passphrase, algo string) string {
	saltBytes, encBytes := encryptPrivKey(privKey, passphrase)
	header := map[string]string{
		headerKDF:  bcryptKDF,
		headerSalt: fmt.Sprintf("%X", saltBytes),
	}
	if algo != "" {
		header[headerType] = algo
	}
	return EncodeArmor(blockTypePrivKey, header, encBytes)
}

// UnarmorDecryptPrivKey decrypts an armored private key and returns the key, algorithm and any error
func UnarmorDecryptPrivKey(armorStr, passphrase string) (privKey cryptotypes.PrivKey, algo string, err error) {
	blockType, header, encBytes, err := DecodeArmor(armorStr)
	if err != nil {
		return nil, "", err
	}

	if err := validatePrivKeyHeader(blockType, header); err != nil {
		return nil, "", err
	}

	saltBytes, err := hex.DecodeString(header[headerSalt])
	if err != nil {
		return nil, "", fmt.Errorf("error decoding salt: %v", err.Error())
	}

	privKey, err = decryptPrivKey(saltBytes, encBytes, passphrase)
	if header[headerType] == "" {
		header[headerType] = defaultAlgo
	}

	return privKey, header[headerType], err
}

// EncodeArmor creates an armored string from the input data and headers
func EncodeArmor(blockType string, headers map[string]string, data []byte) string {
	buf := new(bytes.Buffer)
	w, err := armor.Encode(buf, blockType, headers)
	if err != nil {
		panic(fmt.Errorf("could not encode ascii armor: %s", err))
	}
	
	if _, err := w.Write(data); err != nil {
		panic(fmt.Errorf("could not encode ascii armor: %s", err))
	}
	
	if err := w.Close(); err != nil {
		panic(fmt.Errorf("could not encode ascii armor: %s", err))
	}
	
	return buf.String()
}

// DecodeArmor decodes an armored string and returns the block type, headers, data and any error
func DecodeArmor(armorStr string) (string, map[string]string, []byte, error) {
	buf := bytes.NewBufferString(armorStr)
	block, err := armor.Decode(buf)
	if err != nil {
		return "", nil, nil, err
	}
	
	data, err := io.ReadAll(block.Body)
	if err != nil {
		return "", nil, nil, err
	}
	
	return block.Type, block.Header, data, nil
}

// Helper functions

func validatePrivKeyHeader(blockType string, header map[string]string) error {
	if blockType != blockTypePrivKey {
		return fmt.Errorf("unrecognized armor type: %v", blockType)
	}

	if header[headerKDF] != bcryptKDF {
		return fmt.Errorf("unrecognized KDF type: %v", header[headerKDF])
	}

	if header[headerSalt] == "" {
		return fmt.Errorf("missing salt bytes")
	}

	return nil
}

func unarmorBytes(armorStr, blockType string) ([]byte, map[string]string, error) {
	bType, header, bz, err := DecodeArmor(armorStr)
	if err != nil {
		return nil, nil, err
	}

	if bType != blockType {
		return nil, nil, fmt.Errorf("unrecognized armor type %q, expected: %q", bType, blockType)
	}

	return bz, header, nil
}

func encryptPrivKey(privKey cryptotypes.PrivKey, passphrase string) (saltBytes []byte, encBytes []byte) {
	saltBytes = crypto.CRandBytes(16)
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), BcryptSecurityParameter)
	if err != nil {
		panic(sdkerrors.Wrap(err, "error generating bcrypt key from passphrase"))
	}

	key = crypto.Sha256(key)
	privKeyBytes := legacy.Cdc.MustMarshal(privKey)
	return saltBytes, xsalsa20symmetric.EncryptSymmetric(privKeyBytes, key)
}

func decryptPrivKey(saltBytes []byte, encBytes []byte, passphrase string) (cryptotypes.PrivKey, error) {
	key, err := bcrypt.GenerateFromPassword(saltBytes, []byte(passphrase), BcryptSecurityParameter)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "error generating bcrypt key from passphrase")
	}

	key = crypto.Sha256(key)
	privKeyBytes, err := xsalsa20symmetric.DecryptSymmetric(encBytes, key)
	if err != nil {
		if err.Error() == "Ciphertext decryption failed" {
			return nil, sdkerrors.ErrWrongPassword
		}
		return nil, err
	}

	return legacy.PrivKeyFromBytes(privKeyBytes)
}
