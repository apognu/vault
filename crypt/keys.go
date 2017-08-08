package crypt

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/apognu/vault/util"
	"github.com/google/uuid"

	"golang.org/x/crypto/pbkdf2"

	"github.com/Sirupsen/logrus"
)

var passphraseCache []byte

func createVault() error {
	if _, err := os.Stat(util.GetVaultPath()); !os.IsNotExist(err) {
		return nil
	}
	return os.MkdirAll(util.GetVaultPath(), 0700)
}

func InitVault() {
	// Retrieve initial passphrase
	passPhrase, err := GetPassphrase("Initial vault passphrase", true)
	if err != nil {
		logrus.Fatalf("could not read passphrase: %s", err)
	}
	passSalt := uuid.New().String()
	passKey := pbkdf2.Key(GenerateKey([]byte(passPhrase)), []byte(passSalt), 4096, 32, sha512.New)

	// Generate random master key
	keyBytes := make([]byte, 4096)
	_, err = rand.Read(keyBytes)
	if err != nil {
		logrus.Fatalf("could not generate random key: %s", err)
	}

	// Encrypt master key with key derived from initial passphrase
	masterSalt, err := uuid.NewUUID()
	if err != nil {
		logrus.Fatalf("could not generate salt: %s", err)
	}

	key := pbkdf2.Key(keyBytes, []byte(masterSalt.String()), 4096, 32, sha512.New)
	nonce, aesgcm := GetCipher(passKey, nil)
	ciphertext := aesgcm.Seal(nil, nonce, key, nil)

	meta := &util.VaultMeta{
		MasterKeys: []util.MasterKey{
			{
				Salt:  fmt.Sprintf("%x", passSalt),   // Salt used to derive the key from the passphrase
				Nonce: fmt.Sprintf("%x", nonce),      // Nonce used in encrypting the master key
				Data:  fmt.Sprintf("%x", ciphertext), // Encrypted master key
			},
		},
	}

	createVault()

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

	_, err = metaFile.Write([]byte(fmt.Sprintf("%x", metaJson)))
	if err != nil {
		logrus.Fatalf("could not write secret: %s", err)
	}

	logrus.Info("vault created successfully")
}

func GetMasterKey(confirm, getPassphrase bool) []byte {
	// Retrieve hashed passphrase either from console or seal
	var passphrase []byte
	if len(passphraseCache) == 0 {
		if !IsUnsealed() {
			pass, err := GetPassphrase("Enter passphrase", confirm)
			if err != nil {
				logrus.Fatalf("could not read passphrase: %s", err)
			}
			passphrase = GenerateKey(pass)
		} else {
			seal, err := GetSeal()
			if err != nil {
				logrus.Fatalf("could not retrieve passphrase from seal: %s", err)
			}
			passphrase = seal
		}
	} else {
		passphrase = passphraseCache
	}

	// Retrieve vault metadata
	metaFile, err := ioutil.ReadFile(fmt.Sprintf("%s/_vault.meta", util.GetVaultPath()))
	if err != nil {
		logrus.Fatalf("could not open vault metadata: %s", err)
	}
	metaJson, err := hex.DecodeString(string(metaFile))
	if err != nil {
		logrus.Fatalf("could not read vault metadata salt: %s", err)
	}
	var meta util.VaultMeta
	err = json.Unmarshal(metaJson, &meta)
	if err != nil {
		logrus.Fatalf("could not read vault metadata: %s", err)
	}

	// Try and find a key slot than can be decrypted with provided key
	for _, mkey := range meta.MasterKeys {
		salt, err := hex.DecodeString(mkey.Salt)
		if err != nil {
			logrus.Fatalf("could not read vault metadata salt: %s", err)
		}
		nonce, err := hex.DecodeString(mkey.Nonce)
		if err != nil {
			logrus.Fatalf("could not read vault metadata nonce: %s", err)
		}
		data, err := hex.DecodeString(mkey.Data)
		if err != nil {
			logrus.Fatalf("could not read vault metadata data: %s", err)
		}

		key := pbkdf2.Key(passphrase, []byte(salt), 4096, 32, sha512.New)
		nonce, aesgcm := GetCipher(key, nonce)
		masterKey, err := aesgcm.Open(nil, nonce, data, nil)
		if err != nil {
			// Go to the next key slot
			continue
		}

		passphraseCache = passphrase

		if getPassphrase {
			return passphrase
		} else {
			return masterKey
		}
	}

	logrus.Fatalf("could not find matching passphrase")
	return []byte{}
}
