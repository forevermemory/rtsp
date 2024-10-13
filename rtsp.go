package main

import (
	"fmt"
	"rtsp/client"
)

func main() {
	var a = []byte{0x24}
	fmt.Println(string(a))
	return
	var addr = "rtsp://192.168.120.177:8554/edge/541021d8-56b4-44be-9d7e-438e220058bd/mark"
	r, err := client.NewRtspClient(addr)
	fmt.Println(r, err)
}
