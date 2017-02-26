package main

import (
	"crypto/rand"
	"fmt"
)

var ID string

func createID() {
	//generate just random data for now
	buffer := make([]byte, 16)
	rand.Read(buffer)
	ID = fmt.Sprintf("%x", buffer)
}

func main() {
	createID()
	fmt.Println(ID)
	//run the connection routine
	connectToMaster()
	select {}
}
