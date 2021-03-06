package crypt

import (
	"fmt"
	"testing"

	"github.com/apognu/vault/util"
	"github.com/stretchr/testify/assert"
)

func TestDummy(t *testing.T) {
	assert.Equal(t, 1, 1, "should be equal")
}

func TestKeyGeneration(t *testing.T) {
	hash := GenerateKey([]byte("Sup3rS3cre7"))

	assert.Equal(t, "8780685daa543f122479d8c8510981bb8dc5b9bcd74a5fecb42ad3955ed22ee577f21bbe4facf7d7bb7074498c87c6adfc202c99be78091cb636ad34fb26ce65", fmt.Sprintf("%x", hash), "passphrase should be hashed")
}

func TestEncryption(t *testing.T) {
	passphrase := []byte("Sup3rS3cre7")
	attrs := util.AttributeMap{
		"username": &util.Attribute{Value: "apognu"},
		"password": &util.Attribute{Value: "strongpassword"},
	}

	encryptedSecret, err := EncryptData(attrs, passphrase)
	assert.Nil(t, err)
	assert.NotNil(t, encryptedSecret)

	decryptedAttrs, err := DecryptData(encryptedSecret, passphrase)
	assert.Nil(t, err)

	assert.NotNil(t, decryptedAttrs)
	assert.Equal(t, "apognu", decryptedAttrs["username"].Value, "username should be 'apognu'")
	assert.Equal(t, "strongpassword", decryptedAttrs["password"].Value, "password should be 'strongpassword'")
}

func TestFailingEncryption(t *testing.T) {
	attrs := util.AttributeMap{
		"username": &util.Attribute{Value: "apognu"},
		"password": &util.Attribute{Value: "strongpassword"},
	}

	encryptedSecret, err := EncryptData(attrs, []byte("Sup3rS3cre7"))
	assert.Nil(t, err)
	assert.NotNil(t, encryptedSecret)

	decryptedAttrs, err := DecryptData(encryptedSecret, []byte("WrongPassphrase"))
	assert.NotNil(t, err)
	assert.Nil(t, decryptedAttrs)
}
