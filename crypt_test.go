package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummy(t *testing.T) {
	assert.Equal(t, 1, 1, "should be equal")
}

func TestKeyGeneration(t *testing.T) {
	hash := generateKey([]byte("Sup3rS3cre7"))

	assert.Equal(t, "8780685daa543f122479d8c8510981bb8dc5b9bcd74a5fecb42ad3955ed22ee577f21bbe4facf7d7bb7074498c87c6adfc202c99be78091cb636ad34fb26ce65", fmt.Sprintf("%x", hash), "passphrase should be hashed")
}

func TestEncryption(t *testing.T) {
	passphrase := []byte("Sup3rS3cre7")
	attrs := map[string]string{
		"username": "apognu",
		"password": "strongpassword",
	}

	encryptedSecret, err := encryptData(attrs, passphrase, []string{})
	assert.Nil(t, err)
	assert.NotNil(t, encryptedSecret)

	decryptedAttrs, err := decryptData(encryptedSecret, passphrase)
	assert.Nil(t, err)

	assert.NotNil(t, decryptedAttrs)
	assert.Equal(t, "apognu", decryptedAttrs["username"], "username should be 'apognu'")
	assert.Equal(t, "strongpassword", decryptedAttrs["password"], "password should be 'strongpassword'")
}

func TestFailingEncryption(t *testing.T) {
	attrs := map[string]string{
		"username": "apognu",
		"password": "strongpassword",
	}

	encryptedSecret, err := encryptData(attrs, []byte("Sup3rS3cre7"), []string{})
	assert.Nil(t, err)
	assert.NotNil(t, encryptedSecret)

	decryptedAttrs, err := decryptData(encryptedSecret, []byte("WrongPassphrase"))
	assert.NotNil(t, err)
	assert.Nil(t, decryptedAttrs)
}
