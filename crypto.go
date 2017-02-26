package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

var secret []byte

func createNewKey() ([]byte, *big.Int, *big.Int, error) {
	return elliptic.GenerateKey(elliptic.P521(), rand.Reader)
}

func createNewCipher(priv []byte, slaveX, slaveY *big.Int) (cipher.AEAD, error) {
	sharedSecretX, sharedSecretY := elliptic.P521().ScalarMult(slaveX, slaveY, priv)
	hash := sha256.Sum256(append(sharedSecretX.Bytes(), sharedSecretY.Bytes()...))
	block, _ := aes.NewCipher(hash[:])
	return cipher.NewGCM(block)
}
