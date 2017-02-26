package main

import (
	"crypto/cipher"
	"encoding/gob"
	"log"
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

func newMasterConn(t *net.TCPConn) (conn *encryptedConnection, err error) {
	conn = new(encryptedConnection)
	conn.tcpConn = t
	//create a new private key for exchange
	priv, X, Y, err := createNewKey()
	if err != nil {
		return nil, err
	}
	//TODO: Write key exchange
	conn.cipherBlock, err = createNewCipher(priv, X, Y)
	return
}

type masterConnection struct {
	conn    *encryptedConnection
	encoder *gob.Encoder
	decoder *gob.Decoder
}

var globalConnection *masterConnection

func connectToMaster() {

	masterAddr, err := net.ResolveTCPAddr("tcp", globalConfig.Master.MasterAddress+":"+globalConfig.Master.MasterPort)
	if err != nil {
		log.Fatal("Unable to resolve address for master:", err)
	}

	tcpConn, err := net.DialTCP("tcp", nil, masterAddr)
	if err != nil {
		log.Fatal(err)
	}

	//make the new encrypted connection
	newMasterConn(tcpConn)
}
