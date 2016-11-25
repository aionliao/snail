package main

import (
	"bzcom/biubiu/media/protocol/hls"
	"bzcom/biubiu/media/protocol/httpflv"
	"bzcom/biubiu/media/protocol/rtmp"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func catchSignal() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGUSR1)
	<-sig
	log.Println("recieved signal!")
	panic("")
}

// this is for demo
func main() {

	go catchSignal()

	go func() {
		log.Println(http.ListenAndServe(":8089", nil))
	}()

	stream := rtmp.NewRtmpStream()

	rtmplisten, err := net.Listen("tcp", ":1935")
	if err != nil {
		log.Fatal(err)
	}
	flvListen, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	hlsListen, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatal(err)
	}

	hlsServer := hls.NewServer()
	go hlsServer.Serve(hlsListen)

	rtmpServer := rtmp.NewRtmpServer(stream, hlsServer)
	go rtmpServer.Serve(rtmplisten)

	// rtmpclient := rtmp.NewRtmpClient(stream, nil)
	// rtmpclient.Dial("rtmp://127.0.0.1:1935/live/test", "publish")

	hdlServer := httpflv.NewServer(stream)
	go hdlServer.Serve(flvListen)

	select {}
}
