package crypt

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"github.com/Sirupsen/logrus"
	"github.com/apognu/vault/util"
	"github.com/google/uuid"
)

func ListKeys() {
	meta := GetVaultMeta()

	util.FormatKeyList(meta.MasterKeys)
}

func AddKey(comment string) {
	masterKey := GetMasterKey(false, false)
	passphrase, err := GetPassphrase("New passphrase", true)
	if err != nil {
		logrus.Fatalf("could not read passphrase: %s", err)
	}
	passSalt := uuid.New().String()
	passKey := pbkdf2.Key(GenerateKey([]byte(passphrase)), []byte(passSalt), util.BpkdfIterations, util.BpkdfKeySize, sha512.New)

	nonce, aesgcm := GetCipher(passKey, nil)
	ciphertext := aesgcm.Seal(nil, nonce, masterKey, nil)

	meta := GetVaultMeta()
	meta.MasterKeys = append(meta.MasterKeys, util.MasterKey{
		Comment:   comment,
		CreatedOn: int(time.Now().Unix()),
		Salt:      fmt.Sprintf("%x", passSalt),   // Salt used to derive the key from the passphrase
		Nonce:     fmt.Sprintf("%x", nonce),      // Nonce used in encrypting the master key
		Data:      fmt.Sprintf("%x", ciphertext), // Encrypted master key
	})

	// Write vault metadata to metadata file
	metaFile, err := os.Create(fmt.Sprintf("%s/_vault.meta", util.GetVaultPath()))
	if err != nil {
		logrus.Fatalf("could not create vault metadata: %s", err)
	}
	defer metaFile.Close()
	metaFile.Chmod(0600)
	metaJson, err := json.Marshal(meta)
	if err != nil {
		logrus.Fatalf("could not marshal secret: %s", err)
	}

	_, err = metaFile.Write(metaJson)
	if err != nil {
		logrus.Fatalf("could not write secret: %s", err)
	}

	logrus.Info("key was successfully added")
}

func DeleteKey(id int) {
	meta := GetVaultMeta()
	if len(meta.MasterKeys) == 1 {
		logrus.Fatal("cannot delete the last key from the vault")
	}
	if id >= len(meta.MasterKeys) {
		logrus.Fatal("unknown key ID")
	}
	meta.MasterKeys = append(meta.MasterKeys[:id], meta.MasterKeys[id+1:]...)

	// Write vault metadata to metadata file
	metaFile, err := os.Create(fmt.Sprintf("%s/_vault.meta", util.GetVaultPath()))
	if err != nil {
		logrus.Fatalf("could not create vault metadata: %s", err)
	}
	defer metaFile.Close()
	metaFile.Chmod(0600)
	metaJson, err := json.Marshal(meta)
	if err != nil {
		logrus.Fatalf("could not marshal secret: %s", err)
	}

	_, err = metaFile.Write(metaJson)
	if err != nil {
		logrus.Fatalf("could not write secret: %s", err)
	}

	logrus.Info("key was successfully deleted")
}
