package main

import (
	"crypto/cipher"
	"encoding/gob"
	"log"
	"math/big"
	"net"
	"time"
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

	//create a raw gob (d)e(n)coder
	decoder := gob.NewDecoder(t)
	encoder := gob.NewEncoder(t)

	//create a new private key for exchange
	priv, X, Y, err := createNewKey()
	if err != nil {
		return nil, err
	}

	keyTransferPacket := &keyTransfer{
		newPacket(KeyTransfer),
		X.Bytes(),
		Y.Bytes(),
		X.Sign(),
		Y.Sign(),
	}

	//send the public key
	err = encoder.Encode(keyTransferPacket)
	if err != nil {
		t.Close()
		return
	}
	//receive the master's public key
	masterKeyPacket := new(keyTransfer)
	err = decoder.Decode(masterKeyPacket)
	if err != nil {
		t.Close()
		return
	}

	//reconstruct public key from the packet rx'd
	masterX := big.NewInt(0).SetBytes(masterKeyPacket.X)
	if masterKeyPacket.XSign == -1 {
		masterX = masterX.Neg(masterX)
	}
	masterY := big.NewInt(0).SetBytes(masterKeyPacket.Y)
	if masterKeyPacket.YSign == -1 {
		masterY = masterY.Neg(masterY)
	}

	conn.cipherBlock, err = createNewCipher(priv, masterX, masterY)
	return
}

type masterConnection struct {
	conn    *encryptedConnection
	encoder *gob.Encoder
	decoder *gob.Decoder
}

var globalConnection *masterConnection

func connectToMaster() {

	masterAddr, err := net.ResolveTCPAddr("tcp", globalConfig.Master.Address+":"+globalConfig.Master.Port)
	if err != nil {
		log.Fatal("Unable to resolve address for master: ", err)
	}

	tcpConn, err := net.DialTCP("tcp", nil, masterAddr)
	if err != nil {
		//reattempt connection up to 5 times using doubling backoff
		backoffStartMilli := uint(250)
		for i := uint(0); i < 5; i++ {
			<-time.Tick(time.Millisecond * time.Duration(backoffStartMilli<<i))
			log.Printf("Re-attempting to connect (attempt #%d; waited: %.2f s)\n", i+1, float64(backoffStartMilli<<i)/1000)
			tcpConn, err = net.DialTCP("tcp", nil, masterAddr)
			if err == nil {
				break
			}
		}
		//if still unable to connect fatally report error
		if err != nil {
			log.Fatal(err)
		}
	}

	//make the new encrypted connection
	newMasterConn(tcpConn)
}
