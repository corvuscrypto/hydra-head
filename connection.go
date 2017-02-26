package main

import (
	"crypto/cipher"
	"encoding/gob"
	"net"
)

type encryptedConnection struct {
	tcpConn     *net.TCPConn
	cipherBlock cipher.AEAD
}

func (e encryptedConnection) Write(data []byte) (n int, err error) {
	var cipherText []byte
	e.cipherBlock.Seal(cipherText, nil, data, nil)
	return e.tcpConn.Write(cipherText)
}

func (e encryptedConnection) Read(dst []byte) (n int, err error) {
	var cipherText []byte
	n, err = e.tcpConn.Read(cipherText)
	e.cipherBlock.Open(dst, nil, cipherText, nil)
	return
}

type masterConnection struct {
	conn    *encryptedConnection
	encoder *gob.Encoder
	decoder *gob.Decoder
}
