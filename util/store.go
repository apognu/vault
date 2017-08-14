package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

var vaultDir = fmt.Sprintf("%s/.vault", os.Getenv("HOME"))

func AssertVaultExists() {
	if _, err := os.Stat(GetVaultPath()); os.IsNotExist(err) {
		logrus.Fatal("vault does not exist, consider running init")
	}
}

func GetVaultPath() string {
	if os.Getenv("VAULT_PATH") != "" {
		vaultDir = os.Getenv("VAULT_PATH")
	}

	return strings.TrimSuffix(vaultDir, "/")
}
