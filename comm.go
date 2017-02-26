package main

//go:generate stringer -type=packetType
type packetType uint8

//The different packet types we expect
const (
	Ping packetType = iota
	Acknowledgement
	KeyTransfer
	DiscoveryRequest
	DiscoveryChallenge
	DiscoveryAccept
	DiscoveryReject
	DataRequest
	DataContinue
	DataResponse
	StatusRequest
	StatusResponse
	ErrorResponse
)

func newPacket(t packetType) packet {
	basePacket := new(packet)
	basePacket.Type = t
	return *basePacket
}

//Base struct to compose all other packets
type packet struct {
	Type packetType
}

type pingPacket struct {
	packet
	ID packetID
}

type acknowledgementPacket struct {
	packet
	ID packetID
}

type keyTransfer struct {
	packet
	X     []byte
	Y     []byte
	XSign int
	YSign int
}

type discoveryRequest struct {
	packet
	ID        string
	Resources []string
}

type discoveryChallenge struct {
	packet
	Nonce []byte
}

type discoveryChallengeResponse struct {
	packet
	Hash []byte
}

type discoveryAccept struct {
	packet
}

type discoveryReject struct {
	packet
	Reason string
}

type dataRequestPacket struct {
	packet
	ID       packetID
	Resource string
	Filter   string
	OrderBy  string
}

type dataResponsePacket struct {
	packet
	ID       packetID
	DataLeft int
	Data     []byte
}

type statusRequest struct {
	packet
	ID     packetID
	Fields []string
}

type statusResponse struct {
	packet
	ID   packetID
	Data []byte
}

type errorResponse struct {
	packet
	ID    packetID
	Error string
}
