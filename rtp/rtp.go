package rtp

import (
	"encoding/binary"
	"fmt"
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
	P  uint8  // 1bit Padding, if set, the last byte mean the padding size
	X  uint8  // 1bit Extension
	CC uint8  // 4bit CSRC count
	M  uint8  // 1bit   // 1 frame end, 0
	PT uint8  // 7bit payload type,96
	SN uint16 // 16bit sequence number

	//
	// HeaderFlag uint32
	Timestamp uint32
	SSRC      uint32 // Synchronization Source identifier
	CSRC      []uint32
}

type RtpPacket struct {
	Index int
	Size  int

	RtpHeader   *RtpHeader
	PaddingSize byte
	Payload     []byte
}

const (
	csrcOffset int = 12
)

func (p *RtpPacket) Unmarshal(buf []byte) error {

	if p.RtpHeader == nil {
		p.RtpHeader = new(RtpHeader)
	}

	n, err := p.RtpHeader.Unmarshal(buf)
	if err != nil {
		return err
	}

	end := len(buf)
	if p.RtpHeader.P == 1 {
		if end <= n {
			return fmt.Errorf("%v", "too small buff")
		}
		p.PaddingSize = buf[end-1]
		end -= int(p.PaddingSize)
	}
	if end < n {
		return fmt.Errorf("%v", "too small buff")
	}
	p.Payload = buf[n:end]
	p.Size = len(p.Payload)

	return nil
}

func (rh *RtpHeader) Unmarshal(buff []byte) (int, error) {
	var n int
	var err error

	var byte1 = buff[0]
	var byte2 = buff[1]

	rh.V = byte1 >> 6 & 0x3
	rh.P = byte1 >> 5 & 0x1
	rh.X = byte1 >> 4 & 0x1
	rh.CC = byte1 & 0x0f
	rh.M = byte2 >> 7 & 0x1
	rh.PT = byte2 & 0x7f
	rh.SN = binary.BigEndian.Uint16(buff[2:4])

	rh.Timestamp = binary.BigEndian.Uint32(buff[4:8])
	rh.SSRC = binary.BigEndian.Uint32(buff[8:12])

	if rh.CC > 0 {
		rh.CSRC = make([]uint32, rh.CC)
		for i := 0; i < int(rh.CC); i++ {
			offset := csrcOffset + (i * 4)
			rh.CSRC[i] = binary.BigEndian.Uint32(buff[offset:])
		}
	}
	n = 12 + int(rh.CC)*4

	return n, err
}
