package keys

import (
    "errors"
    "fmt"

    "github.com/baron-chain/cometbft-bc/libs/cli"
    "github.com/spf13/cobra"

    "github.com/baron-chain/cosmos-sdk/client"
    "github.com/baron-chain/cosmos-sdk/crypto/keyring"
    "github.com/baron-chain/cosmos-sdk/crypto/keys/multisig"
    "github.com/baron-chain/cosmos-sdk/crypto/keys/kyber"
    "github.com/baron-chain/cosmos-sdk/crypto/ledger"
    cryptotypes "github.com/baron-chain/cosmos-sdk/crypto/types"
    sdk "github.com/baron-chain/cosmos-sdk/types"
    sdkerr "github.com/baron-chain/cosmos-sdk/types/errors"
)

const (
    FlagAddress = "address"
    FlagPublicKey = "pubkey"
    FlagBechPrefix = "bech"
    FlagDevice = "device"
    FlagQuantumSafe = "quantum-safe"
    flagMultiSigThreshold = "multisig-threshold"
)

func ShowKeysCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "show [name_or_address [name_or_address...]]",
        Short: "Retrieve quantum-safe key information",
        Long: `Display key details with quantum-safe encryption support. 
Multiple names or addresses create an ephemeral multisig key under "multi".`,
        Args: cobra.MinimumNArgs(1),
        RunE: runShowCmd,
    }

    flags := cmd.Flags()
    flags.String(FlagBechPrefix, sdk.PrefixAccount, "Bech32 prefix encoding (acc|val|cons)")
    flags.BoolP(FlagAddress, "a", false, "Output address only")
    flags.BoolP(FlagPublicKey, "p", false, "Output public key only")
    flags.BoolP(FlagDevice, "d", false, "Output address in ledger device")
    flags.BoolP(FlagQuantumSafe, "q", true, "Use quantum-safe encryption")
    flags.Int(flagMultiSigThreshold, 1, "K out of N required signatures")

    return cmd
}

func runShowCmd(cmd *cobra.Command, args []string) error {
    clientCtx, err := client.GetClientQueryContext(cmd)
    if err != nil {
        return err
    }

    isQuantumSafe, _ := cmd.Flags().GetBool(FlagQuantumSafe)
    if isQuantumSafe {
        return handleQuantumSafeShow(cmd, args, clientCtx)
    }
    return handleClassicShow(cmd, args, clientCtx)
}

func handleQuantumSafeShow(cmd *cobra.Command, args []string, clientCtx client.Context) error {
    k := new(keyring.Record)
    var err error

    if len(args) == 1 {
        k, err = fetchQuantumSafeKey(clientCtx.Keyring, args[0])
    } else {
        k, err = handleMultiSigQuantumSafe(cmd, args, clientCtx)
    }
    if err != nil {
        return err
    }

    return displayKeyInfo(cmd, k, clientCtx.OutputFormat)
}

func fetchQuantumSafeKey(kb keyring.Keyring, keyref string) (*keyring.Record, error) {
    k, err := kb.Key(keyref)
    if err == nil || !sdkerr.IsOf(err, sdkerr.ErrIO, sdkerr.ErrKeyNotFound) {
        return k, err
    }

    accAddr, err := sdk.AccAddressFromBech32(keyref)
    if err != nil {
        return nil, err
    }

    k, err = kb.KeyByAddress(accAddr)
    if err != nil {
        return nil, sdkerr.Wrap(err, "invalid quantum-safe key")
    }

    if !isQuantumSafeKey(k) {
        return nil, errors.New("key is not quantum-safe")
    }

    return k, nil
}

func handleMultiSigQuantumSafe(cmd *cobra.Command, args []string, clientCtx client.Context) (*keyring.Record, error) {
    threshold, _ := cmd.Flags().GetInt(flagMultiSigThreshold)
    if err := validateMultisigThreshold(threshold, len(args)); err != nil {
        return nil, err
    }

    pks := make([]cryptotypes.PubKey, len(args))
    for i, keyref := range args {
        k, err := fetchQuantumSafeKey(clientCtx.Keyring, keyref)
        if err != nil {
            return nil, fmt.Errorf("error processing key %s: %v", keyref, err)
        }
        key, err := k.GetPubKey()
        if err != nil {
            return nil, err
        }
        pks[i] = key
    }

    multikey := multisig.NewLegacyAminoPubKey(threshold, pks)
    return keyring.NewMultiRecord("multi", multikey)
}

func isQuantumSafeKey(k *keyring.Record) bool {
    pub, err := k.GetPubKey()
    if err != nil {
        return false
    }
    _, isKyber := pub.(*kyber.PubKey)
    return isKyber
}

func displayKeyInfo(cmd *cobra.Command, k *keyring.Record, outputFormat string) error {
    isShowAddr, _ := cmd.Flags().GetBool(FlagAddress)
    isShowPubKey, _ := cmd.Flags().GetBool(FlagPublicKey)
    isShowDevice, _ := cmd.Flags().GetBool(FlagDevice)

    if isShowAddr && isShowPubKey {
        return errors.New("cannot use both --address and --pubkey")
    }

    bechPrefix, _ := cmd.Flags().GetString(FlagBechPrefix)
    bechKeyOut, err := getBechKeyOut(bechPrefix)
    if err != nil {
        return err
    }

    if isShowDevice {
        return handleDeviceDisplay(k, bechPrefix, isShowPubKey)
    }

    return printKeyringRecord(cmd.OutOrStdout(), k, bechKeyOut, outputFormat)
}
