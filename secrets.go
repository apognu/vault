package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/atotto/clipboard"
)

func createVault() error {
	if os.Getenv("HOME") == "" {
		logrus.Fatal("vault assumes HOME environment variable is set")
	}

	if _, err := os.Stat(vaultDir); !os.IsNotExist(err) {
		return nil
	}
	return os.MkdirAll(vaultDir, 0700)
}

func listSecrets(path string) {
	dirPath := fmt.Sprintf("%s/%s", vaultDir, path)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		logrus.Fatal("secret does not exist")
	}

	FormatDirectory(path, 0)
}

func getSecret(path string) (*Secret, map[string]string) {
	filePath := fmt.Sprintf("%s/%s", vaultDir, path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logrus.Fatal("secret does not exist")
	}

	cipherHex, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Fatalf("could not retrieve secret: %s", err)
	}

	cipherJson, err := hex.DecodeString(string(cipherHex))
	if err != nil {
		logrus.Fatalf("could not decode secret: %s", err)
	}

	var cipherData Secret
	err = json.Unmarshal(cipherJson, &cipherData)
	if err != nil {
		logrus.Fatalf("could not unmarshal secret: %s", err)
	}

	// Get the passphrase from the console if the store is sealed
	var passphrase []byte
	if !isUnsealed() {
		seal, err := getPassphrase("Enter passphrase", false)
		if err != nil {
			logrus.Fatalf("could not read passphrase: %s", err)
		}
		passphrase = seal
	}

	// Decrypt secret encrypted data
	attrs, err := decryptData(&cipherData, passphrase)
	if err != nil {
		logrus.Fatalf("could not decrypt secret: %s", err)
	}

	return &cipherData, attrs
}

func showSecret(path string, print bool, clip bool, clipAttr string) {
	cipherData, attrs := getSecret(path)

	if clipAttr == "" {
		if len(cipherData.EyesOnly) == 1 {
			clipAttr = cipherData.EyesOnly[0]
		} else {
			clipAttr = "password"
		}
	}

	if clip {
		if attrs[clipAttr] != "" {
			clipboard.WriteAll(attrs[clipAttr])
			logrus.Infof("attribute '%s' of '%s' was copied to your clipboard", clipAttr, path)
			return
		} else {
			logrus.Fatalf("could not read attribute '%s'", clipAttr)
		}
	}

	FormatAttributes(path, attrs, cipherData.EyesOnly, print)
}

func addSecret(path string, attrs map[string]string, eyesOnly []string, edit bool, editedAtts []string) {
	// Check if the secret already exists in ADD mode
	filePath := fmt.Sprintf("%s/%s", vaultDir, path)
	if !edit {
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			logrus.Fatal("secret already exists")
		}
	}

	// For each attribute, set its value
	for k, v := range attrs {
		// If eyes-only attirbute, prompt for it on the command-line
		if v == "" {
			pass, err := getPassphrase(fmt.Sprintf("Value for '%s'", k), false)
			if err != nil {
				logrus.Fatalf("could not read attribute: %s", err)
			}
			attrs[k] = string(pass)

			// Store the eyes-only state
			if !StringArrayContains(eyesOnly, k) {
				eyesOnly = append(eyesOnly, k)
			}
		} else {
			// If the attribute is NOT eyes-only, potentially remove it from the list
			if StringArrayContains(editedAtts, k) {
				eyesOnly = RemoveFromSlice(eyesOnly, k)
			}
		}
	}

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

	var passphrase []byte
	if !isUnsealed() {
		seal, err := getPassphrase("Enter passphrase", true)
		if err != nil {
			logrus.Fatalf("could not read passphrase: %s", err)
		}
		passphrase = seal
	}

	// Get encrypted secret Go struct
	cipherData, err := encryptData(attrs, passphrase, eyesOnly)
	if err != nil {
		logrus.Fatalf("could not encrypt secret: %s", err)
	}

	cipherJson, err := json.Marshal(cipherData)
	if err != nil {
		logrus.Fatalf("could not marshal secret: %s", err)
	}

	_, err = secretFile.Write([]byte(fmt.Sprintf("%x", cipherJson)))
	if err != nil {
		logrus.Fatalf("could not write secret: %s", err)
	}

	if edit {
		logrus.Infof("secret '%s' edited successfully", path)
		gitCommit(path, GIT_EDIT)
	} else {
		logrus.Infof("secret '%s' created successfully", path)
		gitCommit(path, GIT_ADD)
	}
}

func editSecret(path string, newAttrs map[string]string, deletedAttrs []string) {
	cipherData, attrs := getSecret(path)
	editedAttrs := make([]string, 0)

	// Replace old attributes with new ones
	for k, _ := range newAttrs {
		attrs[k] = newAttrs[k]
		editedAttrs = append(editedAttrs, k)
	}

	// Remove deleted attributes from the map
	for _, k := range deletedAttrs {
		delete(attrs, k)
	}

	addSecret(path, attrs, cipherData.EyesOnly, true, editedAttrs)
}

func deleteSecret(path string) {
	err := os.Remove(fmt.Sprintf("%s/%s", vaultDir, path))
	if err != nil {
		logrus.Fatalf("could not remove secret: %s", err)
	}

	logrus.Infof("secret '%s' deleted successfully", path)
	gitCommit(path, GIT_DELETE)

	// Remove any empty parent directory
	for {
		dir, _ := filepath.Split(filepath.Clean(path))
		if dir == "" {
			break
		}

		err := os.Remove(fmt.Sprintf("%s/%s", vaultDir, dir))
		if err != nil {
			return
		}

		path = dir
	}

}
