package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/apognu/vault/util"
	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ssh/terminal"
)

func GetCipher(passphrase, nonce []byte) ([]byte, cipher.AEAD) {
	if nonce == nil {
		nonce = make([]byte, 12)
		if _, err := io.ReadFull(crand.Reader, nonce); err != nil {
			logrus.Fatalf("could not generate nonce: %s", err)
		}
	}

	block, err := aes.NewCipher(passphrase)
	if err != nil {
		logrus.Fatalf("could not create cipher: %s", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		logrus.Fatalf("could not create cipher: %s", err)
	}

	return nonce, aesgcm
}

func GetPassphrase(prompt string, confirm bool) ([]byte, error) {
	fmt.Printf("%s: ", prompt)
	passphrase, err := terminal.ReadPassword(0)
	fmt.Println("")

	if strings.TrimSpace(string(passphrase)) == "" {
		logrus.Fatal("could not use empty passphrase")
	}

	if confirm {
		fmt.Print("Confirm: ")
		confirmation, err := terminal.ReadPassword(0)
		fmt.Println("")
		if err != nil {
			logrus.Fatal("could not read confirmation passphrases")
		}

		if string(passphrase) != string(confirmation) {
			logrus.Fatal("passphrases do not match")
		}
	}

	return passphrase, err
}

func GenerateKey(passphrase []byte) []byte {
	sha := sha512.New()
	sha.Write([]byte(passphrase))
	hash := sha.Sum(nil)

	return hash
}

func EncryptData(attrs util.AttributeMap, passphrase []byte) (*util.Secret, error) {
	salt := uuid.New().String()
	key := pbkdf2.Key(passphrase, []byte(salt), util.BpkdfIterations, util.BpkdfKeySize, sha512.New)

	plainData, err := json.Marshal(attrs)
	if err != nil {
		return nil, err
	}

	nonce, aesgcm := GetCipher(key, nil)
	ciphertext := aesgcm.Seal(nil, nonce, plainData, nil)

	return &util.Secret{
		Salt:  fmt.Sprintf("%x", salt),
		Nonce: fmt.Sprintf("%x", nonce),
		Data:  fmt.Sprintf("%x", ciphertext),
	}, nil
}

func DecryptData(secret *util.Secret, passphrase []byte) (util.AttributeMap, error) {
	salt, err := hex.DecodeString(secret.Salt)
	if err != nil {
		return nil, err
	}
	nonce, err := hex.DecodeString(secret.Nonce)
	if err != nil {
		return nil, err
	}
	cipherData, err := hex.DecodeString(secret.Data)
	if err != nil {
		return nil, err
	}

	key := pbkdf2.Key(passphrase, salt, util.BpkdfIterations, util.BpkdfKeySize, sha512.New)
	_, aesgcm := GetCipher(key, nonce)

	plainJson, err := aesgcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, err
	}

	var attrs util.AttributeMap
	err = json.Unmarshal(plainJson, &attrs)
	if err != nil {
		return nil, err
	}

	return attrs, nil
}
