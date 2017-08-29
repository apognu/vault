package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
	passphrase, err := terminal.ReadPassword(0)

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

func GetSecretFile(path string) (*util.Secret, error) {
	filePath := fmt.Sprintf("%s/%s", util.GetVaultPath(), path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	cipherJson, err := ioutil.ReadFile(filePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	var cipherData util.Secret
	err = json.Unmarshal(cipherJson, &cipherData)
	if err != nil {
		logrus.Fatalf("could not unmarshal secret: %s", err)
	}

	return &cipherData, err
}

func GetSecret(path string) (*util.Secret, util.AttributeMap) {
	cipherData, err := GetSecretFile(path)
	if err != nil {
		logrus.Fatalf("could not retrieve secret: %s", err)
	}

	// Get the passphrase from the console if the store is sealed
	masterKey := GetMasterKey(false, false, false)

	// Decrypt secret encrypted data
	attrs, err := DecryptData(cipherData, masterKey)
	if err != nil {
		logrus.Fatalf("could not decrypt secret: %s", err)
	}

	return cipherData, attrs
}

func SetSecret(path string, attrs util.AttributeMap, generatorLength int, generatorSymbols, edit bool, editedAttrs []string, rotation bool) {
	filePath := fmt.Sprintf("%s/%s", util.GetVaultPath(), path)

	// For each attribute, set its value
	for k, v := range attrs {
		// If eyes-only attribute, prompt for it on the command-line
		if v.Value == "" {
			pass, err := GetPassphrase(fmt.Sprintf("Value for '%s'", k), false)
			if err != nil {
				logrus.Fatalf("could not read attribute: %s", err)
			}
			attrs[k].Value = string(pass)
			attrs[k].EyesOnly = true
		} else if v.Value[0] == '@' {
			filePath := v.Value[1:]
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				logrus.Fatalf("could not open file %s: %s", filePath, err)
			}
			b64 := base64.StdEncoding.EncodeToString(content)

			attrs[k].Value = b64
			attrs[k].File = true
		} else if v.Value == "-" {
			attrs[k].Value = GeneratePassword(generatorLength, generatorSymbols)
			attrs[k].EyesOnly = true
		} else {
			attrs[k].EyesOnly = false
		}
	}

	masterKey := GetMasterKey(false, false, rotation)
	err := os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		logrus.Fatalf("could not create hierarchy: %s", err)
	}

	secretFile, err := os.Create(filePath)
	if err != nil {
		logrus.Fatalf("could not create secret: %s", err)
	}
	defer secretFile.Close()
	secretFile.Chmod(0600)

	// Get encrypted secret Go struct
	cipherData, err := EncryptData(attrs, masterKey)
	if err != nil {
		logrus.Fatalf("could not encrypt secret: %s", err)
	}

	cipherJson, err := json.Marshal(cipherData)
	if err != nil {
		logrus.Fatalf("could not marshal secret: %s", err)
	}

	_, err = secretFile.Write(cipherJson)
	if err != nil {
		logrus.Fatalf("could not write secret: %s", err)
	}

	if edit {
		logrus.Infof("secret '%s' edited successfully", path)
		util.GitCommit(path, util.GIT_EDIT, "")
	} else {
		logrus.Infof("secret '%s' created successfully", path)
		util.GitCommit(path, util.GIT_ADD, "")
	}
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
