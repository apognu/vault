package main

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

func unseal() {
	if isUnsealed() {
		logrus.Fatal("store is already unsealed")
	}

	passphrase, err := getPassphrase("Enter passphrase")
	if err != nil {
		logrus.Fatalf("could not unseal store: %s", err)
	}

	key := generateKey([]byte(passphrase))

	sealFile, err := os.Create(sealPath)
	if err != nil {
		logrus.Fatalf("could not unseal store: %s", err)
	}
	defer sealFile.Close()
	sealFile.Chmod(0400)
	sealFile.Write(key)

	logrus.Info("store is now unsealed")
}

func seal() {
	err := os.Remove(sealPath)
	if err != nil {
		logrus.Fatalf("could not seal store: %s", err)
	}

	logrus.Info("store is now sealed")
}

func isUnsealed() bool {
	if _, err := os.Stat(sealPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func getSeal() ([]byte, error) {
	key, err := ioutil.ReadFile(sealPath)
	if err != nil {
		return nil, err
	}
	return key, nil
}
