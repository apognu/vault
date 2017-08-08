package crypt

type MasterKey struct {
	Salt  string `json:"salt"`
	Nonce string `json:"nonce"`
	Data  string `json:"data"`
}

type VaultMeta struct {
	MasterKeys []MasterKey `json:"master_keys"`
}

type Secret struct {
	Salt     string   `json:"salt"`
	Nonce    string   `json:"nonce"`
	Data     string   `json:"data"`
	EyesOnly []string `json:"eyesonly"`
}
