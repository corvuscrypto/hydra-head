package main

import (
	"bytes"
	"crypto/cipher"
	"encoding/binary"
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

func (e *encryptedConnection) Write(data []byte) (n int, err error) {
	nonce := createNonce()
	var cipherText []byte
	cipherText = e.cipherBlock.Seal(cipherText, nonce, data, nil)
	return e.tcpConn.Write(append(nonce, cipherText...))
}

func (e *encryptedConnection) Read(dst []byte) (n int, err error) {
	var buffer = make([]byte, 1)
	n, err = e.tcpConn.Read(buffer)
	if err != nil {
		return
	}

	vlqLength := buffer[0]
	buf2 := make([]byte, vlqLength)
	n, err = e.tcpConn.Read(buf2)
	if err != nil {
		return
	}
	bytesReader := bytes.NewReader(buf2)
	length, err := binary.ReadUvarint(bytesReader)
	finBuf := make([]byte, length)
	n, err = e.tcpConn.Read(finBuf)
	var dest []byte
	dest, err = e.cipherBlock.Open(nil, finBuf[:12], finBuf[12:n], nil)
	n = len(dest)
	for i, b := range dest {
		dst[i] = b
	}

	return
}

func newMasterConn(t *net.TCPConn) (conn *encryptedConnection, err error) {
	conn = new(encryptedConnection)
	conn.tcpConn = t

	//create a raw gob (d)e(n)coder
	decoder := gob.NewDecoder(t)
	encoder := gob.NewEncoder(t)

	//this should be in the discovery file, but I'll move to there later
	//I just want to get things tested and move forward

	//first send the discovery offer
	initialDiscoveryPacket := &discoveryRequest{}
	initialDiscoveryPacket.packet = newPacket(DiscoveryRequest)
	initialDiscoveryPacket.SlaveID = ID
	initialDiscoveryPacket.Resources = []string{"exampleResource1", "exampleResource2"}
	err = encoder.Encode(initialDiscoveryPacket)
	if err != nil {
		log.Fatal(err)
	}
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
	log.Println("Sending key")
	//send the public key
	err = encoder.Encode(keyTransferPacket)
	if err != nil {
		t.Close()
		return
	}
	log.Println("receiving key")
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

	log.Println("constructing shared secret", err)
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

	log.Println("Connected to master")

	//make the new encrypted connection
	encConn, err := newMasterConn(tcpConn)
	if err != nil {
		encConn.tcpConn.Close()
	}

	// encoder := gob.NewEncoder(encConn)
	// decoder := gob.NewDecoder(encConn)

	log.Println("created encrypted connection")
	globalConnection = &masterConnection{}
	globalConnection.conn = encConn
	globalConnection.encoder = gob.NewEncoder(globalConnection.conn)
	globalConnection.decoder = gob.NewDecoder(globalConnection.conn)

	log.Println("Awaiting challenge")
	//Now wait for the challenge
	challenge := new(discoveryChallenge)
	err = globalConnection.decoder.Decode(challenge)
	if err != nil {
		return
	}

	log.Println("Received Challenge")
}
