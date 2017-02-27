package main

import (
	"crypto/rand"
	"encoding/gob"
	"sync/atomic"
	"time"
)

type packetID uint64

var packetPerturbator uint64

func newPacketID() packetID {
	perturbator := atomic.AddUint64(&packetPerturbator, uint64(1)) % 1000
	now := time.Now()
	micro := uint64(now.Nanosecond()/1000) * 1000
	timestamp := uint64(now.Unix()*1000000000) + micro
	return packetID(timestamp + perturbator)
}

func receivePacket(d *gob.Decoder, expectedPacket interface{}) error {
	return d.Decode(expectedPacket)
}

func sendPacket(e *gob.Encoder, packet interface{}) {
	e.Encode(packet)
}

func createNonce() (nonce []byte) {
	nonce = make([]byte, 12)
	rand.Read(nonce)
	return
}
