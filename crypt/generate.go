package crypt

import (
	"math/rand"
	"time"
)

var basicChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var symbolChars = []rune(`!"#$%&'()*+,-./:;<=>?@[\]$^_{|}~`)

func GeneratePassword(length int, symbols bool) string {
	rand.Seed(time.Now().UTC().UnixNano())

	password := make([]rune, length)
	for i := range password {
		var char rune
		if symbols {
			if rand.Intn(4) < 1 {
				char = symbolChars[rand.Intn(len(symbolChars))]
			} else {
				char = basicChars[rand.Intn(len(basicChars))]
			}
		} else {
			char = basicChars[rand.Intn(len(basicChars))]
		}
		password[i] = char
	}

	return string(password)
}
