package main

import (
	"bzcom/biubiu/media/protocol/rtmp"
	"fmt"
	"log"
	"net"

	"github.com/valyala/fasthttp"
)

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Welcome bjoy!\n")
}

// this is for demo
func main() {
	l, err := net.Listen("tcp", ":1935")
	if err != nil {
		log.Fatal(err)
	}
	rtmpServer := rtmp.NewRtmpServer(rtmp.NewRtmpStream())
	rtmpServer.Serve(l)
}
