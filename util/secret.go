package util

const (
	BpkdfIterations = 8192
	BpkdfKeySize    = 32
)

type MasterKey struct {
	Comment   string `json:"comment"`
	CreatedOn int    `json:"created_on"`
	Salt      string `json:"salt"`
	Nonce     string `json:"nonce"`
	Data      string `json:"data"`
}

type VaultMeta struct {
	MasterKeys []MasterKey `json:"master_keys"`
}

type AttributeMap map[string]*Attribute

func (m AttributeMap) FindFirstEyesOnly() string {
	for k, v := range m {
		if v.EyesOnly {
			return k
		}
	}
	return ""
}

func (m AttributeMap) EyesOnlyCount() int {
	i := 0
	for _, v := range m {
		if v.EyesOnly {
			i += 1
		}
	}
	return i
}

type Attribute struct {
	Value    string `json:"value"`
	EyesOnly bool   `json:"eyesonly"`
	File     bool   `json:"file"`
}

type Secret struct {
	Salt  string `json:"salt"`
	Nonce string `json:"nonce"`
	Data  string `json:"data"`
}
