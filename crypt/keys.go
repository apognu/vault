package crypt

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"github.com/Sirupsen/logrus"
	"github.com/apognu/vault/util"
	"github.com/google/uuid"
)

func ListKeys() {
	meta := GetVaultMeta(false)

	util.FormatKeyList(meta.MasterKeys)
}

func AddKey(comment string) {
	masterKey := GetMasterKey(false, false, false)
	passphrase, err := GetPassphrase("New passphrase", true)
	if err != nil {
		logrus.Fatalf("could not read passphrase: %s", err)
	}
	passSalt := uuid.New().String()
	passKey := pbkdf2.Key(GenerateKey([]byte(passphrase)), []byte(passSalt), util.BpkdfIterations, util.BpkdfKeySize, sha512.New)

	nonce, aesgcm := GetCipher(passKey, nil)
	ciphertext := aesgcm.Seal(nil, nonce, masterKey, nil)

	meta := GetVaultMeta(false)
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
	util.GitCommit("_vault.meta", util.GIT_EDIT, fmt.Sprintf("Created key '%s'", comment))
}

func DeleteKey(id int) {
	meta := GetVaultMeta(false)
	if len(meta.MasterKeys) == 1 {
		logrus.Fatal("cannot delete the last key from the vault")
	}
	if id >= len(meta.MasterKeys) {
		logrus.Fatal("unknown key ID")
	}
	comment := meta.MasterKeys[id].Comment
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
	util.GitCommit("_vault.meta", util.GIT_DELETE, fmt.Sprintf("Deleted key '%s'", comment))
}

func RotateKey() {
	fmt.Println(`WARNING: rotating the vault's master key will invalidate all the user passphrases except the one used here.`)
	fmt.Println(`If the process fails for any reasons, please check your vault repository is clean before going any further.`)
	fmt.Print("Are you sure you want to rotate the vault's master key ? (y/N) ")

	answer := ""
	fmt.Scanln(&answer)
	if strings.TrimSpace(strings.ToLower(answer)) != "y" {
		logrus.Fatal("aborting...")
	}

	Seal(true)

	// Retrieve initial passphrase
	passphrase := GetMasterKey(false, true, false)
	passSalt := uuid.New().String()
	passKey := pbkdf2.Key(passphrase, []byte(passSalt), util.BpkdfIterations, util.BpkdfKeySize, sha512.New)

	// Generate random master key
	keyBytes := make([]byte, 4096)
	_, err := rand.Read(keyBytes)
	if err != nil {
		logrus.Fatalf("could not generate random key: %s", err)
	}

	// Encrypt master key with key derived from initial passphrase
	masterSalt, err := uuid.NewUUID()
	if err != nil {
		logrus.Fatalf("could not generate salt: %s", err)
	}

	key := pbkdf2.Key(keyBytes, []byte(masterSalt.String()), util.BpkdfIterations, util.BpkdfKeySize, sha512.New)
	nonce, aesgcm := GetCipher(passKey, nil)
	ciphertext := aesgcm.Seal(nil, nonce, key, nil)

	meta := &util.VaultMeta{
		MasterKeys: []util.MasterKey{
			{
				Comment:   "Key generated on vault rotation",
				CreatedOn: int(time.Now().Unix()),
				Salt:      fmt.Sprintf("%x", passSalt),   // Salt used to derive the key from the passphrase
				Nonce:     fmt.Sprintf("%x", nonce),      // Nonce used in encrypting the master key
				Data:      fmt.Sprintf("%x", ciphertext), // Encrypted master key
			},
		},
	}

	// Write vault metadata to metadata file
	metaFile, err := os.Create(fmt.Sprintf("%s/_vault.meta.new", util.GetVaultPath()))
	if err != nil {
		logrus.Errorf("could not create vault metadata: %s", err)
		cancelKeyRotation()
	}
	defer metaFile.Close()
	metaFile.Chmod(0600)
	metaJson, err := json.Marshal(meta)
	if err != nil {
		logrus.Errorf("could not marshal secret: %s", err)
		cancelKeyRotation()
	}

	_, err = metaFile.Write(metaJson)
	if err != nil {
		logrus.Errorf("could not write secret: %s", err)
		cancelKeyRotation()
	}

	err = filepath.Walk(util.GetVaultPath(), rotateSecretKey)
	if err != nil {
		cancelKeyRotation()
	}

	os.Rename(fmt.Sprintf("%s/_vault.meta.new", util.GetVaultPath()), fmt.Sprintf("%s/_vault.meta", util.GetVaultPath()))
	util.GitCommit("-A", util.GIT_EDIT, "Rotated vault master key")

	logrus.Info("vault master key rotation successful")
}

func rotateSecretKey(path string, info os.FileInfo, err error) error {
	if strings.HasSuffix(path, ".git") {
		return filepath.SkipDir
	}
	if strings.HasSuffix(path, "_vault.meta") || strings.HasSuffix(path, "_vault.meta.new") {
		return nil
	}
	if f, _ := os.Stat(path); f.IsDir() {
		return nil
	}

	secretPathTokens := strings.Split(path, util.GetVaultPath())
	secretPath := secretPathTokens[1]
	_, attrs := GetSecret(secretPath)

	SetSecret(secretPath, attrs, 0, true, []string{}, true)

	return nil
}

func cancelKeyRotation() {
	util.RunGitCommand(false, "reset", "--hard")
	os.Remove(fmt.Sprintf("%s/_vault.meta.new", util.GetVaultPath()))

	logrus.Fatalf("could not rotate keys, rolling back vault")
}
