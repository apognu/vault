package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ssh/terminal"
)

func getPassphrase(prompt string, confirm bool) ([]byte, error) {
	fmt.Printf("%s: ", prompt)
	passphrase, err := terminal.ReadPassword(0)
	fmt.Println("")

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

func generateKey(passphrase []byte) []byte {
	sha := sha512.New()
	sha.Write([]byte(passphrase))
	hash := sha.Sum(nil)

	return hash
}

func encryptData(attrs map[string]string, passphrase []byte, eyesOnly []string) (*Secret, error) {
	var hash []byte
	if isUnsealed() {
		seal, err := getSeal()
		if err != nil {
			logrus.Fatalf("could not retrieve seal: %s", err)
		}
		hash = seal
	} else {
		hash = generateKey(passphrase)
	}

	salt := uuid.New().String()
	key := pbkdf2.Key(hash, []byte(salt), 4096, 32, sha512.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainData, err := json.Marshal(attrs)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plainData, nil)

	return &Secret{
		Salt:     fmt.Sprintf("%x", salt),
		Nonce:    fmt.Sprintf("%x", nonce),
		Data:     fmt.Sprintf("%x", ciphertext),
		EyesOnly: eyesOnly,
	}, nil
}

func decryptData(secret *Secret, passphrase []byte) (map[string]string, error) {
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

	var hash []byte
	if isUnsealed() {
		seal, err := getSeal()
		if err != nil {
			logrus.Fatalf("could not retrieve seal: %s", err)
		}
		hash = seal
	} else {
		hash = generateKey(passphrase)
	}

	key := pbkdf2.Key(hash, salt, 4096, 32, sha512.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainJson, err := aesgcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, err
	}

	var attrs map[string]string
	err = json.Unmarshal(plainJson, &attrs)
	if err != nil {
		return nil, err
	}

	return attrs, nil
}
