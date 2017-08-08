package crypt

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
)

var (
	user     = os.Getenv("USER")
	sealPath = fmt.Sprintf("/tmp/vault-%s.seal", user)
)

func Unseal() {
	if IsUnsealed() {
		logrus.Fatal("store is already unsealed")
	}

	passphrase := GetMasterKey(false, true)
	sealFile, err := os.Create(sealPath)
	if err != nil {
		logrus.Fatalf("could not unseal store: %s", err)
	}
	defer sealFile.Close()
	sealFile.Chmod(0400)
	sealFile.Write(passphrase)

	logrus.Info("store is now unsealed")
}

func Seal() {
	err := os.Remove(sealPath)
	if err != nil {
		logrus.Fatalf("could not seal store: %s", err)
	}

	logrus.Info("store is now sealed")
}

func IsUnsealed() bool {
	if _, err := os.Stat(sealPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func GetSeal() ([]byte, error) {
	key, err := ioutil.ReadFile(sealPath)
	if err != nil {
		return nil, err
	}
	return key, nil
}
