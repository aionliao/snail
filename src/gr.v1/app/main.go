package main

import "github.com/kataras/iris"

func main() {
	iris.Get("/", func(ctx *iris.Context) {
		ctx.Write("Hello, %s", "World!")
	})
	iris.Listen(":8080")
}
