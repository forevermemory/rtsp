package rtp

import (
	"encoding/binary"
)

type RtpHeader struct {
	/*
		0                   1                   2                   3
		0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
		0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7
		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		|V=2|P|X|  CC   |M|     PT      |       sequence number         |
		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		|                           timestamp                           |
		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
		|           synchronization source (SSRC) identifier            |
		+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
		|            contributing source (CSRC) identifiers             |
		|                             ....                              |
		+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	*/

	V  uint8  // 2bit
	P  uint8  // 1bit Padding
	X  uint8  // 1bit Extension
	CC uint8  // 4bit CSRC size
	M  uint8  // 1bit
	PT uint8  // 7bit payload type,96
	SN uint16 // 16bit sequence number

	//
	// HeaderFlag uint32
	Timestamp uint32
	SSRC      uint32 // Synchronization Source identifier
	CSRC      []byte
}

func ParseRtpHeader(data []byte) *RtpHeader {
	var rh = new(RtpHeader)
	var byte1 = data[0]
	var byte2 = data[1]

	rh.V = (byte1) & 0b1100
	rh.P = (byte1) & 0b0010
	rh.X = (byte1) & 0b0001
	rh.CC = (byte1) & 0x0f
	rh.M = (byte2) & 0b00000001
	rh.PT = (byte2) & 0b11111110
	rh.SN = binary.BigEndian.Uint16(data[2:4])

	rh.Timestamp = binary.BigEndian.Uint32(data[4:8])
	rh.SSRC = binary.BigEndian.Uint32(data[8:12])

	if rh.CC > 0 {

	}

	return rh
}
