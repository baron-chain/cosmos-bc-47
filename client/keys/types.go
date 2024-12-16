package keys

// AddNewKey represents a request to create a new key with optional quantum-safe encryption
type AddNewKey struct {
    Name         string `json:"name"`
    Password     string `json:"password"`
    Mnemonic     string `json:"mnemonic"`
    Account      int    `json:"account,string,omitempty"`
    Index        int    `json:"index,string,omitempty"`
    QuantumSafe  bool   `json:"quantum_safe,omitempty"`
    KeyAlgorithm string `json:"key_algorithm,omitempty"` // kyber, dilithium, etc.
}

// NewAddNewKey creates a new quantum-safe key request with default settings
func NewAddNewKey(name, password, mnemonic string, account, index int) AddNewKey {
    return AddNewKey{
        Name:         name,
        Password:     password,
        Mnemonic:     mnemonic,
        Account:      account,
        Index:        index,
        QuantumSafe:  true,
        KeyAlgorithm: "kyber",
    }
}

// RecoverKey represents a request to recover a key with quantum-safe support
type RecoverKey struct {
    Password     string `json:"password"`
    Mnemonic     string `json:"mnemonic"`
    Account      int    `json:"account,string,omitempty"`
    Index        int    `json:"index,string,omitempty"`
    QuantumSafe  bool   `json:"quantum_safe,omitempty"`
    KeyAlgorithm string `json:"key_algorithm,omitempty"`
}

// NewRecoverKey creates a new quantum-safe key recovery request
func NewRecoverKey(password, mnemonic string, account, index int) RecoverKey {
    return RecoverKey{
        Password:     password,
        Mnemonic:     mnemonic,
        Account:      account,
        Index:        index,
        QuantumSafe:  true,
        KeyAlgorithm: "kyber",
    }
}

// UpdateKeyReq represents a request to update key passwords with quantum-safe verification
type UpdateKeyReq struct {
    OldPassword string `json:"old_password"`
    NewPassword string `json:"new_password"`
    RotateKey   bool   `json:"rotate_key,omitempty"`    // Option to rotate quantum keys
    ReEncrypt   bool   `json:"re_encrypt,omitempty"`    // Re-encrypt with new quantum algo
    Algorithm   string `json:"algorithm,omitempty"`      // New quantum algorithm if rotating
}

// NewUpdateKeyReq creates a new key update request with quantum-safe options
func NewUpdateKeyReq(old, new string) UpdateKeyReq {
    return UpdateKeyReq{
        OldPassword: old,
        NewPassword: new,
        RotateKey:   false,
        ReEncrypt:   true,
        Algorithm:   "kyber",
    }
}

// DeleteKeyReq represents a request to delete a key with additional verification
type DeleteKeyReq struct {
    Password        string `json:"password"`
    ConfirmDelete   bool   `json:"confirm_delete,omitempty"`
    BackupRequired  bool   `json:"backup_required,omitempty"`
}

// NewDeleteKeyReq creates a new key deletion request with safety checks
func NewDeleteKeyReq(password string) DeleteKeyReq {
    return DeleteKeyReq{
        Password:       password,
        ConfirmDelete: true,
        BackupRequired: true,
    }
}

// KeyValidation represents key validation options
type KeyValidation struct {
    QuantumSafe     bool   `json:"quantum_safe"`
    SecurityLevel   string `json:"security_level,omitempty"`    // high, medium, low
    Algorithm       string `json:"algorithm"`
    Version        string `json:"version,omitempty"`
}

// NewKeyValidation creates default key validation settings
func NewKeyValidation() KeyValidation {
    return KeyValidation{
        QuantumSafe:   true,
        SecurityLevel: "high",
        Algorithm:     "kyber",
        Version:      "1.0",
    }
}
