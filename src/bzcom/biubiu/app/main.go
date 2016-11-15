package main

import (
	"bzcom/biubiu/media/protocol/rtmp"
	"fmt"
	"log"
	"net"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Welcome to snail streaming server!\n")
}

// this is for demo
func main() {
	router := fasthttprouter.New()
	router.GET("/", Index)
	go fasthttp.ListenAndServe(":8080", router.Handler)

	l, err := net.Listen("tcp", ":1935")
	if err != nil {
		log.Fatal(err)
	}
	rtmpServer := rtmp.NewRtmpServer(rtmp.NewRtmpStream())
	rtmpServer.Serve(l)
}
