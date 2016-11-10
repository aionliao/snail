package main

import (
	"fmt"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Welcome!\n")
}

// this is for demo
func main() {
	router := fasthttprouter.New()
	router.GET("/test", Index)
	fasthttp.ListenAndServe(":8080", router.Handler)
}
