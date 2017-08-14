package crypt

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/Sirupsen/logrus"
)

var (
	userName = os.Getenv("USER")
	sealPath = fmt.Sprintf("/tmp/vault-%s.seal", userName)
)

func GetSealPath() string {
	currentUser, err := user.Current()
	if err != nil {
		return sealPath
	}
	meta := GetVaultMeta(false)
	runDir := fmt.Sprintf("/run/user/%s", currentUser.Uid)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		return sealPath
	}
	return fmt.Sprintf("%s/vault-%s.seal", runDir, meta.UUID)
}

func Unseal() {
	if IsUnsealed() {
		logrus.Fatal("store is already unsealed")
	}

	passphrase := GetMasterKey(false, true, false)
	sealFile, err := os.Create(GetSealPath())
	if err != nil {
		logrus.Fatalf("could not unseal store: %s", err)
	}
	defer sealFile.Close()
	sealFile.Chmod(0400)
	sealFile.Write(passphrase)

	logrus.Info("store is now unsealed")
}

func Seal(rotation bool) {
	if !IsUnsealed() {
		if !rotation {
			logrus.Fatal("store is already sealed")
		}
		return
	}

	err := os.Remove(GetSealPath())
	if err != nil {
		logrus.Fatalf("could not seal store: %s", err)
	}

	logrus.Info("store is now sealed")
}

func IsUnsealed() bool {
	if _, err := os.Stat(GetSealPath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func GetSeal() ([]byte, error) {
	key, err := ioutil.ReadFile(GetSealPath())
	if err != nil {
		return nil, err
	}
	return key, nil
}
