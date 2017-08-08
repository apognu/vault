package crypt

import (
	"math/rand"
	"time"
)

var basicChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GeneratePassword(length int) string {
	rand.Seed(time.Now().UTC().UnixNano())

	password := make([]rune, length)
	for i := range password {
		password[i] = basicChars[rand.Intn(len(basicChars))]
	}

	return string(password)
}
