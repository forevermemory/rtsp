package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"rtsp/rtp"
	"strconv"
	"strings"
	"time"
)

type RtspClient struct {
	addr string
	conn net.Conn

	// rtsp connect
	seq           int
	userAgent     string
	playSession   string
	heartbeat     int
	heartbeatPrev time.Time

	// datas
	packetSize  int
	packetCache []byte // 1M
	packets     chan *rtp.RtpPacket
}

func NewRtspClient(addr string) (*RtspClient, error) {
	// rtsp://192.168.120.177:8554/edge/541021d8-56b4-44be-9d7e-438e220058bd/mark

	var ip = "192.168.120.177"
	var port = 8554
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", ip, port), time.Second)
	if err != nil {
		return nil, err
	}

	r := RtspClient{
		addr:          addr,
		conn:          conn,
		seq:           1,
		userAgent:     "LibVLC/3.0.8 (LIVE555 Streaming Media v2016.11.28)",
		heartbeatPrev: time.Now(),
		heartbeat:     0,

		packetSize:  0,
		packetCache: make([]byte, 1024000),
		packets:     make(chan *rtp.RtpPacket, 0x1000),
	}

	r.start()

	// only haikang has
	// go func() {
	// 	for {
	// 		time.Sleep(time.Second * 1)
	// 		hb := r.genmsg_HEARTBEAT()
	// 		r.conn.Write([]byte(hb))
	// 	}
	// }()

	// 一直负责接收即可,并且定时发送heartbeat
	go r.recv()

	r.parseRtpPacket()
	return &r, nil
}

func (r *RtspClient) parseRtpPacket() {
	var pkt *rtp.RtpPacket

	for {
		select {
		case pkt = <-r.packets:
			// rtpHeader = rtp.ParseRtpHeader(pkt.Buff[4:])
			// fmt.Printf("size:%d, V:%d, P:%d, X:%d, CC:%d, M:%d, PT:%d, SN:%d, ts:%d ssrc:%d\n",
			// 	pkt.Size,
			// 	rtpHeader.V, rtpHeader.P, rtpHeader.X, rtpHeader.CC,
			// 	rtpHeader.M, rtpHeader.PT, rtpHeader.SN, rtpHeader.Timestamp,
			// 	rtpHeader.SSRC)
			fmt.Println(pkt.Index)

		default:
			time.Sleep(time.Millisecond * 20)
		}
	}
}

func (r *RtspClient) recv() {

	var n int
	var err error
	var recvIndex int = 0
	var header = make([]byte, 4)
	var size uint16 = 0

	for {
		// first read rtp packet header
		n, err = r.conn.Read(header)
		// fmt.Printf("read err:%v, size:%d,data; %02x %02x %02x %02x %02x %02x %02x %02x \n", err, n,
		// 	buff[0], buff[1], buff[2], buff[3], buff[4], buff[5], buff[6], buff[7])
		if err != nil || n == 0 {
			fmt.Println("rtsp client read header err:", err)
			time.Sleep(time.Second)
			continue
		}

		// Magic:0x24  1byte
		// Channel:0   1byte
		// Size:       2byte
		if header[0] != 0x24 {
			fmt.Println("unknown packet:", string(header))
			continue
		}

		size = binary.BigEndian.Uint16(header[2:4])
		var buff = make([]byte, size)
		n, err = io.ReadFull(r.conn, buff)
		if err != nil || n == 0 {
			fmt.Println("rtsp client read data err:", err)
			time.Sleep(time.Second)
			continue
		}

		// put pkt to channel
		pkt := new(rtp.RtpPacket)
		if err := pkt.Unmarshal(buff); err != nil {
			fmt.Println("pkt Unmarshal buffer err:", err)
			continue
		}
		pkt.Index = recvIndex
		recvIndex += 1

		// if rtpHeader.CC > 0 {
		// 	// var a = *(*[]byte)(unsafe.Pointer(&rtpHeader.CSRC))
		// 	// var b = byte(a)
		// }

		var rtpHeader = pkt.RtpHeader
		fmt.Printf("size:%d, size:%d, V:%d, P:%d, X:%d, CC:%d, M:%d, PT:%d, SN:%d, ts:%d ssrc:%d\n",
			size, pkt.Size,
			rtpHeader.V, rtpHeader.P, rtpHeader.X, rtpHeader.CC,
			rtpHeader.M, rtpHeader.PT, rtpHeader.SN, rtpHeader.Timestamp,
			rtpHeader.SSRC)

	}

}

func (r *RtspClient) start() {
	var buf = make([]byte, 1024)
	var n int
	var err error

	// 1. options
	mgs_options := r.genmsg_OPTIONS()
	fmt.Println(mgs_options)
	n, err = r.conn.Write([]byte(mgs_options))
	fmt.Println("write:", n, err)
	n, err = r.conn.Read(buf)
	fmt.Println("read:", n, err)
	fmt.Println(string(buf[:n]))

	// 2. describe
	msg_describe := r.genmsg_DESCRIBE()
	fmt.Println(msg_describe)
	n, err = r.conn.Write([]byte(msg_describe))
	fmt.Println("write:", n, err)
	buf = make([]byte, 1024)
	n, err = r.conn.Read(buf)
	fmt.Println("read:", n, err)
	fmt.Println(string(buf[:n]))

	// 3.setup
	msg_setup := r.genmsg_SETUP()
	fmt.Println(msg_setup)
	n, err = r.conn.Write([]byte(msg_setup))
	fmt.Println("write:", n, err)
	buf = make([]byte, 1024)
	n, err = r.conn.Read(buf)
	fmt.Println("read:", n, err)
	fmt.Println(string(buf[:n]))

	r.decode_SETUP(string(buf[:n]))
	fmt.Println(r.playSession)

	// 4.play
	msg_play := r.genmsg_PLAY()
	fmt.Println(msg_play)
	n, err = r.conn.Write([]byte(msg_play))
	fmt.Println("write:", n, err)
	buf = make([]byte, 1024)
	n, err = r.conn.Read(buf)
	fmt.Println("read:", n, err)
	fmt.Println(string(buf[:n]))

	// .........
}

func (r *RtspClient) get_seq() int {
	r.seq += 1
	return r.seq
}

func (r *RtspClient) genmsg_OPTIONS() string {
	var s string
	s = "OPTIONS " + r.addr + " RTSP/1.0\r\n"
	s += "CSeq: " + strconv.Itoa(r.get_seq()) + "\r\n"
	s += "User-Agent: " + r.userAgent + "\r\n"
	s += "\r\n"
	return s
}

func (r *RtspClient) genmsg_DESCRIBE() string {
	var s string
	s = "DESCRIBE " + r.addr + " RTSP/1.0\r\n"
	s += "CSeq: " + strconv.Itoa(r.get_seq()) + "\r\n"
	s += "User-Agent: " + r.userAgent + "\r\n"
	s += "Accept: application/sdp\r\n"
	s += "\r\n"
	return s
}

func (r *RtspClient) genmsg_SETUP() string {

	// rn := rand.Intn(10000)
	// rn1 := rn + 40000
	// rn2 := rn + 40010
	var s string
	s = "SETUP " + r.addr + "/trackID=0" + " RTSP/1.0\r\n"
	s += "CSeq: " + strconv.Itoa(r.get_seq()) + "\r\n"
	s += "User-Agent: " + r.userAgent + "\r\n"
	// UDP
	// s += fmt.Sprintf("Transport: RTP/AVP;unicast;client_port=%d-%d\r\n", rn1, rn2)
	// TCP
	s += fmt.Sprintf("Transport: RTP/AVP/TCP;unicast\r\n")
	s += "\r\n"
	return s
}

func (r *RtspClient) decode_SETUP(s string) {

	/*
		RTSP/1.0 200 OK
		CSeq: 4
		Server: gortsplib
		Session: 95dbd4fe6def47eb8449d5281f2dc6c8;timeout=60
		Transport: RTP/AVP;unicast;client_port=52096-52097;server_port=8000-8001;ssrc=48C041CE
	*/

	items := strings.Split(s, "\r\n")

	for _, v := range items {
		// fmt.Println(v)
		tmp1 := strings.Split(v, ": ")
		if len(tmp1) > 1 {
			if tmp1[0] == "Session" {
				//
				r.playSession = strings.Split(tmp1[1], ";")[0]
			}
		}
	}
}

func (r *RtspClient) genmsg_GET_PARAMETER() {

}
func (r *RtspClient) genmsg_PLAY() string {
	var s string
	s = "PLAY " + r.addr + "/" + " RTSP/1.0\r\n"
	s += "CSeq: " + strconv.Itoa(r.get_seq()) + "\r\n"
	s += "User-Agent: " + r.userAgent + "\r\n"
	s += "Session: " + r.playSession + "\r\n"
	s += "Range: npt=0.000-" + "\r\n"
	s += "\r\n"
	return s
}

func (r *RtspClient) genmsg_HEARTBEAT() string {
	var s string
	s = "HEARTBEAT " + r.addr + " RTSP/1.0\r\n"
	s += "CSeq: " + strconv.Itoa(r.get_seq()) + "\r\n"
	s += "User-Agent: " + r.userAgent + "\r\n"
	s += "Session: " + r.playSession + "\r\n"
	s += "\r\n"
	return s
}

func (r *RtspClient) genmsg_TEARDOWN() string {
	var s string
	s = "TEARDOWN " + r.addr + " RTSP/1.0\r\n"
	s += "CSeq: " + strconv.Itoa(r.get_seq()) + "\r\n"
	s += "User-Agent: " + r.userAgent + "\r\n"
	s += "Session: " + r.playSession + "\r\n"
	s += "\r\n"
	return s
}
