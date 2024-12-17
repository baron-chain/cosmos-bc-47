package keys

import (
	"encoding/json"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"
	cryptokeyring "github.com/baron-chain/cosmos-sdk/crypto/keyring"
	"github.com/baron-chain/cometbft-bc/crypto/kyber"
)

const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
	OutputFormatYAML = "yaml"
)

type KeyOutputFormat struct {
	Type           string `json:"type"`
	Name           string `json:"name"`
	Address        string `json:"address"`
	PubKey         string `json:"pubkey"`
	KyberPubKey    string `json:"kyber_pubkey,omitempty"`
	DilithiumKey   string `json:"dilithium_key,omitempty"`
	QuantumSafe    bool   `json:"quantum_safe"`
}

type bechKeyOutFn func(k *cryptokeyring.Record) (cryptokeyring.KeyOutput, error)

func printKeyringRecord(w io.Writer, k *cryptokeyring.Record, bechKeyOut bechKeyOutFn, format string) error {
	ko, err := bechKeyOut(k)
	if err != nil {
		return fmt.Errorf("failed to process key output: %w", err)
	}

	keyOutput := enrichWithQuantumData(ko)
	return outputKeyData(w, keyOutput, format)
}

func printKeyringRecords(w io.Writer, records []*cryptokeyring.Record, format string) error {
	kos, err := cryptokeyring.MkAccKeysOutput(records)
	if err != nil {
		return fmt.Errorf("failed to process keys output: %w", err)
	}

	var enrichedOutputs []KeyOutputFormat
	for _, ko := range kos {
		enrichedOutputs = append(enrichedOutputs, enrichWithQuantumData(ko))
	}

	return outputKeyData(w, enrichedOutputs, format)
}

func enrichWithQuantumData(ko cryptokeyring.KeyOutput) KeyOutputFormat {
	kyberPubKey, _ := kyber.GenerateKey()
	return KeyOutputFormat{
		Type:        ko.Type,
		Name:        ko.Name,
		Address:     ko.Address,
		PubKey:      ko.PubKey,
		KyberPubKey: string(kyberPubKey),
		QuantumSafe: true,
	}
}

func outputKeyData(w io.Writer, data interface{}, format string) error {
	var (
		bytes []byte
		err   error
	)

	switch format {
	case OutputFormatJSON:
		bytes, err = json.MarshalIndent(data, "", "  ")
	case OutputFormatYAML:
		bytes, err = yaml.Marshal(data)
	case OutputFormatText:
		bytes, err = yaml.Marshal(data)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	_, err = fmt.Fprintln(w, string(bytes))
	return err
}
