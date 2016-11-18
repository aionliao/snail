package main

import (
	"bzcom/biubiu/media/protocol/hls"
	"bzcom/biubiu/media/protocol/httpflv"
	"bzcom/biubiu/media/protocol/rtmp"
	"log"
	"net"
)

// this is for demo
func main() {

	stream := rtmp.NewRtmpStream()

	rtmplisten, err := net.Listen("tcp", "127.0.0.1:1935")
	if err != nil {
		log.Fatal(err)
	}
	flvListen, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}
	hlsListen, err := net.Listen("tcp", "127.0.0.1:8082")
	if err != nil {
		log.Fatal(err)
	}

	hlsServer := hls.NewServer()
	go hlsServer.Serve(hlsListen)

	rtmpServer := rtmp.NewRtmpServer(stream, hlsServer)
	go rtmpServer.Serve(rtmplisten)

	hdlServer := httpflv.NewServer(stream)
	go hdlServer.Serve(flvListen)

	select {}
}
