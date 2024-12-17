package keys

import (
    "github.com/baron-chain/baron-sdk/crypto/pqc"
    "github.com/baron-chain/baron-sdk/types"
)

type KeyRequest struct {
    Name         string `json:"name"`
    Password     string `json:"password,omitempty"`
    Mnemonic     string `json:"mnemonic,omitempty"`
    Account      uint32 `json:"account,omitempty"`
    Index        uint32 `json:"index,omitempty"`
    QuantumAlgo  string `json:"quantum_algo,omitempty"`
}

func NewKeyRequest(name, password string) *KeyRequest {
    return &KeyRequest{
        Name:        name,
        Password:    password,
        QuantumAlgo: pqc.DefaultAlgorithm,
        Account:     0,
        Index:       0,
    }
}

type KeyRecovery struct {
    Password    string `json:"password"`
    Mnemonic    string `json:"mnemonic"`
    Account     uint32 `json:"account,omitempty"`
    Index       uint32 `json:"index,omitempty"`
    QuantumAlgo string `json:"quantum_algo,omitempty"`
}

func NewKeyRecovery(password, mnemonic string) *KeyRecovery {
    return &KeyRecovery{
        Password:    password,
        Mnemonic:    mnemonic,
        QuantumAlgo: pqc.DefaultAlgorithm,
    }
}

type KeyUpdate struct {
    OldPass     string `json:"old_password"`
    NewPass     string `json:"new_password"`
    Rotate      bool   `json:"rotate,omitempty"`
    NewAlgo     string `json:"new_algo,omitempty"`
}

func NewKeyUpdate(oldPass, newPass string) *KeyUpdate {
    return &KeyUpdate{
        OldPass:  oldPass,
        NewPass:  newPass,
        Rotate:   true,
        NewAlgo:  pqc.DefaultAlgorithm,
    }
}

type KeyDeletion struct {
    Password string `json:"password"`
    Confirm  bool   `json:"confirm"`
    Backup   bool   `json:"backup"`
}

func NewKeyDeletion(password string, requireBackup bool) *KeyDeletion {
    return &KeyDeletion{
        Password: password,
        Confirm:  true,
        Backup:   requireBackup,
    }
}

type KeyValidation struct {
    QuantumAlgo string         `json:"quantum_algo"`
    Security    types.Security `json:"security"`
    Version     string         `json:"version"`
}

func NewKeyValidation() *KeyValidation {
    return &KeyValidation{
        QuantumAlgo: pqc.DefaultAlgorithm,
        Security:    types.SecurityHigh,
        Version:     types.CurrentVersion,
    }
}
